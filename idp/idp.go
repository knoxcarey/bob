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


package idp

import (
	"io/ioutil"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
        "golang.org/x/oauth2"
)


// Identity provider
type IDPConfig struct {
	Name            string                      // Name of the identity provider
	Endpoint        string                      // Endpoint for ID services
	Revocation      string                      // Revocation endpoint, cf RFC 7009
	ClientID        string                      // Client ID embedded directly
	ClientIDEnv     string                      // Environment variable with Client ID
	ClientSecret    string                      // Client secret embedded directly
	ClientSecretEnv string                      // Environment variable with client secret
	RedirectURL     string                      // URL the provider should redirect to
	// Need to add config options like icons for redirect page
}

// Internal struct for recording provider info
type Provider struct {
	Name       string                           // Friendly name 
	context    *context.Context                 // OAUTH2.0 context
	config     *oauth2.Config                   // OAUTH2.0 configuration structure
	verifier   *oidc.IDTokenVerifier            // Token verifier
	provider   *oidc.Provider                   // Pointer to provider in OIDC library
	idpconfig  *IDPConfig                       // Pointer to struct read from config file
}

// Structure for recording an outstanding auth request
type authRequest struct {
	idpi int                                    // Index of dentity provider auth request went to
	url  string                                 // Original URL that was requested
}

// Structure to contain response data from identity provider
type Auth struct {
	URL         string                          // URL originally requested
	AccessToken string                          // Access token for session
	IDToken     string                          // Identity token
	ExpiresIn   int                             // Timeout for session
	Name        string                          // Authenticated user's name
	ProviderIdx int                             // Index of provider that authenticated
}


var providers []Provider                            // List of identity providers
var requests  map[string]authRequest                // Maps of requests by ephemeral nonce


// Initialize module globals
func init() {
	requests  = make(map[string]authRequest)
}


// Register Auth type so that it can be serialized (into cookie)
func init() {
	gob.Register(&Auth{})
}


// Return list of providers
func Providers() []Provider {
	return providers
}


// Read a configuration file and add an identity provider
func AddIDPFromConfig(file string) {

	// Read the configuration file
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("unable to read configuration file ", file)
	}

	// Unmarshal identity provider config from file
	var idpc IDPConfig
	if err = json.Unmarshal(buffer, &idpc); err != nil {
		log.Fatal("malformed config file ", file)
	}

	// Set client ID and secret if specified by environment variables
	if idpc.ClientIDEnv != "" {
		idpc.ClientID = os.Getenv(idpc.ClientIDEnv)
	}

	if idpc.ClientSecretEnv != "" {
		idpc.ClientSecret = os.Getenv(idpc.ClientSecretEnv)
	}

	// Create OpenID Connect/OAUTH 2.0 structures
        ctx := context.Background()

	// Fetch details of OIDC provider
        provider, err := oidc.NewProvider(ctx, idpc.Endpoint)
        if err != nil {
                log.Fatal("could not reach identity provider: ", idpc.Name)
        }

	// Extract revocation endpoint from provider claims
	var claims map[string]string
	provider.Claims(&claims)
	idpc.Revocation = claims["revocation_endpoint"]
	
        oidcConfig := &oidc.Config{ClientID: idpc.ClientID}

        verifier := provider.Verifier(oidcConfig)

        config := oauth2.Config{
                ClientID:     idpc.ClientID,
                ClientSecret: idpc.ClientSecret,
                Endpoint:     provider.Endpoint(),
                RedirectURL:  idpc.RedirectURL,
                Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "ga4gh"},
        }

	idp := Provider{
		Name: idpc.Name,
		context:  &ctx,
		config:   &config,
		verifier: verifier,
		provider: provider,
		idpconfig: &idpc,
	}
		
	// Add to the list of providers
	providers = append(providers, idp)
}


// Handle redirect to IdP indexed by pi
func Authenticate(pi int, w http.ResponseWriter, r *http.Request) {
	state := nonce(32)
	idp := &providers[pi]
	url, err := url.QueryUnescape(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "bad request", http.StatusInternalServerError)
	}
	requests[state] = authRequest{
		idpi: pi,
		url: url,
	}
	http.Redirect(w, r, idp.config.AuthCodeURL(state), http.StatusFound)
}


// Send revocation request to IdP
func Logout(pi int, accessToken string) {
	idp := &providers[pi]
	auth := fmt.Sprintf("%s:%s", idp.idpconfig.ClientID, idp.idpconfig.ClientSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	form := url.Values{}
	form.Add("token", accessToken)
	form.Add("token_type_hint", "access_token")
	url := idp.idpconfig.Revocation
	if r, e := http.NewRequest("POST", url, strings.NewReader(form.Encode())); e == nil {
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Authorization", "Basic " + encoded)
		client := &http.Client{}
		client.Do(r)
	}
}


// Handle callback from IdP
func Callback(w http.ResponseWriter, r *http.Request) (Auth, error) {
	// Extract state from IDP response
	state := r.URL.Query().Get("state")

	// Determine which IDP that request was to
	idpi := requests[state].idpi
	idp := &providers[idpi]

	// Make sure requests map gets cleaned up when we're done
	defer delete(requests, state)
	
	// Get the OAUTH token
	oauth2Token, err := idp.config.Exchange(*idp.context, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return Auth{}, err
	}

	// Get raw version of ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return Auth{}, err
	}

	// Verify it
	if _, err := idp.verifier.Verify(*idp.context, rawIDToken); err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return Auth{}, err
	}

	// Fetch userinfo
	userInfo, err := idp.provider.UserInfo(*idp.context, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return Auth{}, err
	}

	var claims map[string]interface{}
	userInfo.Claims(&claims)
	name := claims["given_name"].(string) + " " + claims["family_name"].(string)
	
	resp := Auth{
		URL: requests[state].url,
		AccessToken: oauth2Token.Extra("access_token").(string),
		IDToken: rawIDToken,
		ExpiresIn: int(oauth2Token.Extra("expires_in").(float64)),
		ProviderIdx: idpi,
		Name: name,
	}

	return resp, nil
}



// Generate a random nonce string
var digits = [...]string{"0","1","2","3","4","5","6","7","8","9","a","b","c","d","e","f"}
func nonce(len int) string {
	n := ""
	for i := 0; i < len; i++ {
		n += digits[rand.Intn(16)]
	}
	return n
}
