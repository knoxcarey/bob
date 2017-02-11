package idp

import (
	"fmt"
	// oidc "github.com/coreos/go-oidc"
	// "golang.org/x/net/context"
        // "golang.org/x/oauth2"
)



// Identity provider
type IDProvider struct {
	Name            string                      // Name of the identity provider
	Endpoint        string                      // Endpoint for ID services
	ClientID        string                      // Client ID embedded directly
	ClientIDEnv     string                      // Environment variable with Client ID
	ClientSecret    string                      // Client secret embedded directly
	ClientSecretEnv string                      // Environment variable with client secret
	RedirectURL     string                      // URL the provider should redirect to
}




func init() {
	fmt.Println("Initialized idp")
}

