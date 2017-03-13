/***************************************************************************
 Copyright 2017 William Knox Carey

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 ***************************************************************************/

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"	
	"net/url"
	"strconv"
	"time"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


// A net/http handler function that also takes an authentication session argument
type authenticatedHandler func (w http.ResponseWriter, r *http.Request, a *idp.Auth)

// Upgrade structure for websocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


// Lets the user choose from among the registered ID providers
func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/template/login.html"))
	s := struct {
		Providers []idp.Provider
		Page      string
	}{idp.Providers(), r.URL.Query().Get("page")}
	t.Execute(w, s)
}


// Redirects to a chosen identity provider
func loginRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider, err := strconv.Atoi(vars["provider"])
	if err != nil {
		http.Error(w, "Invalid identity provider", http.StatusInternalServerError)
		return
	} else {
		idp.Authenticate(provider, w, r)
	}
}


// Handles OpenID Connect identity provider callbacks
func callbackHandler(w http.ResponseWriter, r *http.Request) {

	// Process identity provider callback, checking tokens, etc.
	auth, err := idp.Callback(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Store session authentication information in cookie
	setCookie(w, r, auth, auth.ExpiresIn)
	
	// Redirect to original page
	http.Redirect(w, r, auth.URL, http.StatusFound)	
}


// Authentication middleware. If not authenticated, redirect to login.
func authenticated(f authenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a idp.Auth
		err := getCookie(r, &a)
		if (err != nil) || (a.IDToken == "") {
			url := "/login?page=" + url.QueryEscape(r.URL.String())
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			f(w, r, &a)
		}
	}
}


// Render the main query page
func queryPageHandler(w http.ResponseWriter, r *http.Request, a *idp.Auth) {
	t := template.Must(template.ParseFiles("static/template/query.html"))
	url := "ws://" + host + ":" + strconv.Itoa(port) + "/ws"
	s := struct {
		Name    string
		URL     string
		Timeout int
		Count   int
	}{a.Name, url, timeout, beacon.Count()}
	t.Execute(w, s)
}


// Handle beacon query; return results asynchronously via websocket
func queryAsyncHandler(w http.ResponseWriter, r *http.Request, a *idp.Auth) {
	num := beacon.Count()
	ch := make(chan beacon.BeaconResponse, num)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}

	query := make(map[string][]string)

	if err := json.Unmarshal(msg, &query); err != nil {
		return
	}

	beacon.QueryBeaconsAsync(query, a.AccessToken, a.IDToken, ch)

	// Collect responses, forwarding over websocket, or timeout
	for i := 0; i < num; i++ {
		select {
		case r := <-ch:
			data, _ := json.Marshal(r)
			conn.WriteMessage(websocket.TextMessage, data)
		case <- time.After(time.Second * time.Duration(timeout)):
			return
		}
	} 	
}


// Handle logout request
func logoutHandler(w http.ResponseWriter, r *http.Request, a *idp.Auth) {
	idp.Logout(a.ProviderIdx, a.AccessToken)
	setCookie(w, r, a, logout)
	
	// Redirect to login
	http.Redirect(w, r, "/login?page=/", http.StatusFound)
}


// Entry point
func main() {
	fmt.Printf("BoB is listening on %s:%d\n", host, port)

	fs := http.FileServer(http.Dir("static/"))
	
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	r.HandleFunc("/login", loginPageHandler)
	r.HandleFunc("/login/{provider}", loginRedirectHandler)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/", authenticated(queryPageHandler))	
	r.HandleFunc("/ws", authenticated(queryAsyncHandler))
	r.HandleFunc("/logout", authenticated(logoutHandler))

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
