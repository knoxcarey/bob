package main
/* TODO:
   * Serve webpage to make query, with js connection via websocket to receive responses
   * Extract common code from beacon versions
   * Automated tests
   * Documentation
   * Dockerfile
   * Integrate OpenID connect 
     * Display interfaces for logging in with various IDPs
     * Send auth headers to beacons along with queries
*/

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"	
	"net/url"
	"strconv"
	"time"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


// Key for cookie encryption
var cookieKey = "7fb62642f70d42e48b1e4b4a48ac94d6"

// Cookie store
var store = sessions.NewCookieStore([]byte(cookieKey))

// Upgrade structure for websocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize: 1024,
	WriteBufferSize:1024,
}



// Authentication middleware. If not authenticated, redirect to login.
func authenticated(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth")
		if (err != nil) || (session.Values["authenticated"] != true) {
			url := "/login?page=" + url.QueryEscape(r.URL.String())
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			f(w, r)
		}
	}
}


// Render login page
func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/template/login.html"))
	s := struct {
		Providers []idp.Provider
		Page      string
	}{idp.Providers(), r.URL.Query().Get("page")}
	t.Execute(w, s)
}


// Redirect to identity provider
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


// Render query page
func queryHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) == 0 {
		queryPageHandler(w, r)
	} else {
		queryAPIHandler(w, r)
	}
}


// Render query page
func queryPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/template/query.html"))
	s := struct {
		Name string
	}{""}
	t.Execute(w, s)
}


// Handle beacon query
func queryAPIHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	accessToken := session.Values["access_token"].(string)
	idToken := session.Values["id_token"].(string)

	query := beacon.BeaconQuery(r.URL.Query())
	results := beacon.QueryBeaconsSync(query, accessToken, idToken, timeout)
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(results)
}


// Handle beacon query; return results asynchronously via websocket
func queryAsyncAPIHandler(w http.ResponseWriter, r *http.Request) {
	num := beacon.Count()
	ch := make(chan beacon.BeaconResponse, num)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Could not open websocket:", err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		query := make(map[string][]string)

		if err := json.Unmarshal(msg, &query); err != nil {
			fmt.Println("Could not unmarshal JSON: ", err)
			return
		}

		fmt.Println("Query: ", query)
		
		session, err := store.Get(r, "auth")
		if err != nil {
			fmt.Println("Could not get cookie session: ", err)
			return
		}

		accessToken := session.Values["access_token"].(string)
		idToken := session.Values["id_token"].(string)

		beacon.QueryBeaconsAsync(query, accessToken, idToken, ch)

		// Collect responses, forwarding over websocket, or timeout
		for i := 0; i < num; i++ {
			select {
			case r := <-ch:
				data, _ := json.Marshal(r)
				if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
					http.Error(w, "Websocket error", http.StatusInternalServerError)
				}
			case <- time.After(time.Second * time.Duration(timeout)):
				break
			}
		} 	
	}
}


// Handle OpenID Connect identity provider callbacks
func callbackHandler(w http.ResponseWriter, r *http.Request) {

	// Process identity provider allback, checking tokens, etc.
	auth, err := idp.Callback(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Set cookie data
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	session.Values["authenticated"] = true
	session.Values["access_token"] = auth.AccessToken
	session.Values["id_token"] = auth.IDToken
	session.Options = &sessions.Options{
		Path: "/",
		MaxAge: auth.ExpiresIn,
		HttpOnly: true,
	}
	session.Save(r, w)
		
	// Redirect to original page
	http.Redirect(w, r, auth.URL, http.StatusFound)	
}


// Entry point
func main() {
	fmt.Printf("BoB is listening on port %d\n", port)

	fs := http.FileServer(http.Dir("static/"))
	
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	r.HandleFunc("/query", authenticated(queryHandler))
	r.HandleFunc("/queryws", queryAsyncAPIHandler)
	r.HandleFunc("/login/{provider}", loginRedirectHandler)
	r.HandleFunc("/login", loginPageHandler)
	r.HandleFunc("/auth/bob/callback", callbackHandler)

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
