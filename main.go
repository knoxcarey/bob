package main

/* TODO:
   * Consider interface for beacons, abstracting different versions
   * Serve webpage to make query, with js connection via websocket to receive responses
   * Make sure non-exported stuff (that isn't needed for JSON) is lowercase
   * Automated tests
   * Documentation
   * Dockerfile
   * Integrate OpenID connect 
     * https://github.com/coreos/go-oidc
     * Read IDP configurations from config files, fetch keys, etc. 
     * Display interfaces for logging in with various IDPs
     * Send auth headers to beacons along with queries
*/

import (
	"fmt"
	"net/http"

	// "github.com/knoxcarey/bob/idp"
	"github.com/knoxcarey/bob/beacon"
)


// Handle HTTP query
func queryHandler(w http.ResponseWriter, r *http.Request) {
	query := beacon.BeaconQuery(r.URL.Query())
	results := beacon.QueryBeaconsSync(query, timeout)
	w.Header().Set("Content-Type", "application/json")
	w.Write(results)
}


// Entry point
func main() {
	fmt.Printf("BoB is listening on port %d\n", port)
	http.HandleFunc("/", queryHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
