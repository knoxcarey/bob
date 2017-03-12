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
	"encoding/gob"
	"net/http"	
	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
	"github.com/knoxcarey/bob/idp"
)


// Key for cookie encryption
var cookieKey = "7fb62642f70d42e48b1e4b4a48ac94d6"

// Cookie store
var store = sessions.NewCookieStore([]byte(cookieKey))


// Register cookie structure for easy (de)serialization
func init() {
	gob.Register(&idp.Auth{})
}


// Read and deserialize cookie
func getAuthCookie(r *http.Request) (idp.Auth, error) {
	session, err := store.Get(r, "auth")
	if err != nil {
		return idp.Auth{}, err
	}
	var a idp.Auth
	if err := mapstructure.Decode(session.Values["auth"], &a); err != nil {
		return idp.Auth{}, err
	} else {
		return a, nil
	}
}


// Serialize and write cookie
func setAuthCookie(w http.ResponseWriter, r *http.Request, a idp.Auth) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}

	session.Values["auth"] = a
	session.Options = &sessions.Options{
		Path: "/",
		MaxAge: a.ExpiresIn,
		HttpOnly: true,
	}

	session.Save(r, w)
	return nil
}


// Log out of current IDP
func sessionLogout(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}

	session.Options = &sessions.Options{
		Path: "/",
		MaxAge: -1,
		HttpOnly: true,
	}

	session.Save(r, w)
	return nil
}
