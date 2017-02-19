package main

/* TODO:
   * Serve webpage to make query, with js connection via websocket to receive responses
   * Extract common code from beacon versions
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

	"github.com/knoxcarey/bob/idp"
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
	// Want a generic handler that catches all requests (except callback?) which:
	//   a) checks for a cookie with auth info in it
	//   b) if no cookie, redirects to auth page, which gives choice of IDPs
	//
	// Also, the callback should set a cookie -- expiring same time as token
	//
	// Need logout endpoint as well to delete cookie

	// Alternatively, interface can show no session, which can be used to make
	// anonymous queries, then only get redirected to auth page on request?
	
	http.HandleFunc("/", idp.Authenticate)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/auth/bob/callback", idp.Callback)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
