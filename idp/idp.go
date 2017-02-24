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
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
        "golang.org/x/oauth2"
)


// Identity provider
type IDPConfig struct {
	Name            string                      // Name of the identity provider
	Endpoint        string                      // Endpoint for ID services
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
}

// Structure for recording an outstanding auth request
type authRequest struct {
	idp *Provider                               // Identity provider auth request went to
	url string                                  // Original URL that was requested
}

// Structure to contain response data from identity provider
type AuthResponse struct {
	URL         string
	AccessToken string
	IDToken     string
	ExpiresIn   int
}


var providers []Provider                            // List of identity providers
var requests  map[string]authRequest                // Maps of requests by ephemeral nonce


// Initialize module globals
func init() {
	requests  = make(map[string]authRequest)
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

	// Unmarshal identity provider config
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

        provider, err := oidc.NewProvider(ctx, idpc.Endpoint)
        if err != nil {
                log.Fatal("could not reach identity provider: ", idpc.Name)
        }

        oidcConfig := &oidc.Config{
                ClientID:       idpc.ClientID,
                SkipNonceCheck: true,
        }

        verifier := provider.Verifier(oidcConfig)

        config := oauth2.Config{
                ClientID:     idpc.ClientID,
                ClientSecret: idpc.ClientSecret,
                Endpoint:     provider.Endpoint(),
                RedirectURL:  "http://127.0.0.1:8080/auth/bob/callback", // FIXME!
                Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "ga4gh"},
        }

	idp := Provider{
		Name: idpc.Name,
		context:  &ctx,
		config:   &config,
		verifier: verifier,
	}
		
	// Add to the list of providers
	providers = append(providers, idp)
}


func Authenticate(pi int, w http.ResponseWriter, r *http.Request) {
	state := nonce(32)
	idp := &providers[pi]
	url, err := url.QueryUnescape(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "bad request", http.StatusInternalServerError)
	}
	requests[state] = authRequest{
		idp: idp,
		url: url,
	}
	http.Redirect(w, r, idp.config.AuthCodeURL(state), http.StatusFound)
}



func Callback(w http.ResponseWriter, r *http.Request) (AuthResponse, error) {
	// Extract state from IDP response
	state := r.URL.Query().Get("state")

	// Determine which IDP that request was to
	idp := requests[state].idp

	// Make sure requests map gets cleaned up when we're done
	defer delete(requests, state)
	
	// Get the OAUTH token
	oauth2Token, err := idp.config.Exchange(*idp.context, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return AuthResponse{}, err
	}

	// Get raw version of ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return AuthResponse{}, err
	}

	// Verify it
	if _, err := idp.verifier.Verify(*idp.context, rawIDToken); err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return AuthResponse{}, err
	}

	resp := AuthResponse{
		URL: requests[state].url,
		AccessToken: oauth2Token.Extra("access_token").(string),
		IDToken: rawIDToken,
		ExpiresIn: int(oauth2Token.Extra("expires_in").(float64)),
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
