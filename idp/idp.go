package idp

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"net/http"
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

type idp struct {
	context    *context.Context                 // OAUTH2.0 context
	config     *oauth2.Config                   // OAUTH2.0 configuration structure
	verifier   *oidc.IDTokenVerifier            // Token verifier
}

type authRequest struct {
	idp *idp                                    // Identity provider auth request went to
	url string                                  // Original URL that was requested
}

var providers []idp                                 // List of identity providers
var requests  map[string]authRequest                // Maps of requests by ephemeral nonce


// Initialize module globals
func init() {
	requests  = make(map[string]authRequest)
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
                RedirectURL:  "/auth/bob/callback",
                Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "ga4gh"},
        }

	idp := idp{
		context:  &ctx,
		config:   &config,
		verifier: verifier,
	}
		
	// Add to the list of providers
	providers = append(providers, idp)
}


func Authenticate(w http.ResponseWriter, r *http.Request) {
	state := "42"           // FIXME! should be a nonce
	idp := &providers[0]    // FIXME! should indicate proper provider
	requests[state] = authRequest{
		idp: idp,
		url: url,       // FIXME! read from request?
	}
	http.Redirect(w, r, idp.config.AuthCodeURL(state), http.StatusFound)
}



func Callback(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	// Get raw version of ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	// Verify it
	idToken, err := idp.verifier.Verify(*idp.context, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// NB: this bit may not be necessary...
	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// End NB: this bit may not be necessary...
	
	// FIXME! -- SET COOKIE
	
	// Redirect back to originally requested page
	http.Redirect(w, r, requests[state].url, http.StatusFound)
}

