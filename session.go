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
	"net/http"	
	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
)


// Key for cookie encryption
var cookieKey = "7fb62642f70d42e48b1e4b4a48ac94d6"

// Cookie store
var store = sessions.NewCookieStore([]byte(cookieKey))

// Expiration to force session logout
var logout = -1


// Read and deserialize cookie
func getCookie(r *http.Request, value interface{}) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}
	if err := mapstructure.Decode(session.Values["auth"], value); err != nil {
		return err
	} else {
		return nil
	}	
}


// Serialize and Write cookie
func setCookie(w http.ResponseWriter, r *http.Request, v interface{}, exp int) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}

	session.Values["auth"] = v
	session.Options = &sessions.Options{
		Path: "/",
		MaxAge: exp,
		HttpOnly: true,
	}

	session.Save(r, w)
	return nil	
}
