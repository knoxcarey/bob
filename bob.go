package main

/* TODO:
   * Consider interface for beacons, abstracting different versions
   * Serve webpage to make query, with js connection via websocket to receive responses
   * Split into multiple packages
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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// oidc "github.com/coreos/go-oidc"
	// "golang.org/x/net/context"
        // "golang.org/x/oauth2"
)


// Beacon contains information about an individual beacon
type Beacon struct {
	Name              string                    // Name we give the beacon internally
	Version           string                    // Beacon API version
	Endpoint          string                    // URL for beacon
	DatasetIds        []string                  // Beacon dataset to query
	QueryMap          map[string]string         // Names of the query fields for this beacon
	AssemblyMap       map[string]string         // Names of the references for this beacon
	AdditionalFields  map[string]string         // Additional query fields to include
}

// BeaconQuery is a type synonym for query
type BeaconQuery map[string]string

// QueryMap maps standard query keywords to non-standard ones
type QueryMap    map[string]string

// AssemblyMap maps standard assembly names to non-standard ones
type AssemblyMap map[string]string

// BeaconResponse contains the response from the beacon
type BeaconResponse struct {
	Name       string             `json:"name"` 
	Status     int                `json:"status"`
	Responses  map[string]string  `json:"responses,omitempty"` 
	Error      map[string]string  `json:"error,omitempty"`              
}

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



// BeaconConfig holds the contents of JSON configuration file
type BeaconConfig struct {
	Beacons     []Beacon     `json:"beacons,omitempty`     // List of beacons
	IDProviders []IDProvider `json:"idProviders,omitempty` // Identity providers
}


// These constants are standard strings used in beacon API
const (
	CHROMOSOME      = "chromosome"
	START           = "start"
	ALTERNATE_BASES = "alternateBases"
	REFERENCE_BASES = "referenceBases"
	ASSEMBLY_ID     = "assemblyId"
	DATASET_IDS     = "datasetIds"
	GRCH37          = "GRCh37"
	GRCH38          = "GRCh38"
)


const (
	defaultConfigFile = "./bob.conf"            // Default location of config file
	defaultPort       = 8080                    // Default port for server
	defaultTimeout    = 20                      // Default timeout for queries, in seconds
)


var (
	// Configuration
	Config BeaconConfig                         

	// Default query parameters
	v02QueryMap = QueryMap {
		CHROMOSOME:      "chromosome",
		START:           "position",
		ALTERNATE_BASES: "allele",
		REFERENCE_BASES: "referenceBases",
		DATASET_IDS:     "dataset",
		ASSEMBLY_ID:     "reference"}
	v03QueryMap = QueryMap {
		CHROMOSOME:      "chromosome",
		START:           "start",
		ALTERNATE_BASES: "alternateBases",
		REFERENCE_BASES: "referenceBases",
		DATASET_IDS:     "datasetIds",
		ASSEMBLY_ID:     "assemblyId"}

	// Default reference assembly names
	v02AssemblyMap = AssemblyMap {    
		GRCH37: "GRCh37",
		GRCH38: "GRCh38"}
	v03AssemblyMap = AssemblyMap {    
		GRCH37: "GRCh37",
		GRCH38: "GRCh38"}

	// Location of the config file
	configFile string

	// Port at which to operate service
	port int

	// Timeout for beacon queries, in seconds
	timeout int
)



// readConfig reads in configuration file with the given file name
func readConfig() {
	if dat, err := ioutil.ReadFile(configFile); err != nil {
		log.Fatal("unable to read configuration file ", configFile)
	} else {
		if err = json.Unmarshal(dat, &Config); err != nil {
			log.Fatal("malformed config file ", configFile)
		}
	}
}



// Apply defaults where fields were *not* specified in the config file
func applyDefaults(beacon *Beacon, qm QueryMap, am AssemblyMap) {
	var k string
	var v string

	if beacon.QueryMap == nil {
		beacon.QueryMap = make(map[string]string)
	}
	
	for k, v = range qm {
		if _, ok := beacon.QueryMap[k]; ok == false {
			beacon.QueryMap[k] = v
		}
	}

	if beacon.AssemblyMap == nil {
		beacon.AssemblyMap = make(map[string]string)
	}
	
	for k, v = range am {
		if _, ok := beacon.AssemblyMap[k]; ok == false {
			beacon.AssemblyMap[k] = v
		}
	}

	if beacon.AdditionalFields == nil {
		beacon.AdditionalFields = make(map[string]string)
	}
}



// Set up beacon data structures by version
func initializeBeacons() {
	for i, _ := range Config.Beacons {
		beacon := &(Config.Beacons[i])
		switch beacon.Version {
		case "0.2":
			applyDefaults(beacon, v02QueryMap, v02AssemblyMap)
		case "0.3":
			applyDefaults(beacon, v03QueryMap, v03AssemblyMap)
		default:
			log.Fatal("Bad version for ", beacon.Name, ". Specify \"0.2\" or \"0.3\".")
		}
	}
}


// Initialize identity provider structures
func initializeIDProviders() {
	for i, _ := range Config.IDProviders {
		idp := &(Config.IDProviders[i])
		if idp.ClientIDEnv != "" {
			idp.ClientID = os.Getenv(idp.ClientIDEnv)
		}
		if idp.ClientSecretEnv != "" {
			idp.ClientSecret = os.Getenv(idp.ClientSecretEnv)
		}
		fmt.Println(idp.ClientID)
		fmt.Println(idp.ClientSecret)
	}
}




// Parse command-line switches; set defaults if not present
func parseSwitches() {
	flag.StringVar(&configFile, "config", defaultConfigFile, "Location of the configuration file")
	flag.IntVar(&port, "port", defaultPort, "Port on which to run server")
	flag.IntVar(&timeout, "timeout", defaultTimeout, "Timeout for beacon queries, in seconds")
	flag.Parse()
}



// Read in configuration and apply defaults
func init() {	
	parseSwitches()
	readConfig()
	initializeBeacons()
	initializeIDProviders()
}



// Query a single beacon
func queryBeacon(beacon Beacon, query *BeaconQuery, ch chan<- BeaconResponse) {
	qs := queryString(&beacon, query)
	uri := fmt.Sprintf("%s?%s", beacon.Endpoint, qs)

	status, body, err := httpGet(uri)	
	fmt.Printf("[%s][%s][%d] %s\n", beacon.Name, friendlyQuery(*query), status, uri)
	resp := parseBeaconResponse(&beacon, status, body, err)

	ch <- *resp
}



// Create a BeaconResponse structure representing the response from the beacon
func parseBeaconResponse(beacon *Beacon, status int, raw []byte, err error) *BeaconResponse {
	response := BeaconResponse{Name: beacon.Name}	

	if err != nil {
		addResponseError(&response, 400, "could not reach beacon")
		return &response		
	}

	if status/100 != 2 {
		addResponseError(&response, status, "beacon error")
		return &response
	}
	
	switch beacon.Version {
	case "0.2":
		parseResponseV02(beacon, status, raw, &response)
	case "0.3":
		parseResponseV03(beacon, status, raw, &response)
	}

	return &response
}



// Add an error condition to the response
func addResponseError(response *BeaconResponse, code int, message string) {
	if response.Error == nil {
		response.Error = make(map[string]string)
	}	
	response.Error["code"] = strconv.Itoa(code)
	response.Error["message"] = message
	response.Status = code
}



// Add a valid result to the response
func addResponseResult(response *BeaconResponse, key string, value string) {
	if response.Responses == nil {
		response.Responses = make(map[string]string)
	}
	response.Responses[key] = value
}



// Parse beacon response for beacon version 0.2
func parseResponseV02(beacon *Beacon, status int, raw []byte, response *BeaconResponse) {
	var v2 struct {Response map[string]string}
	
	if err := json.Unmarshal(raw, &v2); err == nil {
		if v2.Response["error"] == "" {
			response.Status = status
			addResponseResult(response, beacon.Name, v2.Response["exists"])
		} else {
			addResponseError(response, 400, v2.Response["error"])
		}
	} else {
		addResponseError(response, 400, "malformed reply from beacon")
	}
}



// Parse beacon response for beacon version 0.3
func parseResponseV03(beacon *Beacon, status int, raw []byte, response *BeaconResponse) {
	var v3 struct {
		DatasetIds              []string
		DatasetAlleleResponses  []map[string]string
		Error                   map[string]string
		Exists                  string}

	if err := json.Unmarshal(raw, &v3); err == nil {
		if v3.Error == nil {
			response.Status = status
			for _, r := range v3.DatasetAlleleResponses {
				addResponseResult(response, r["id"], r["exists"])
			}
		} else {
			code, _ := strconv.Atoi(v3.Error["errorCode"])
			addResponseError(response, code, v3.Error["message"])
		}
	} else {
		addResponseError(response, 400, "malformed reply from beacon")
	}
}



// Wrapper for HTTP get
func httpGet(uri string) (status int, body []byte, err error) {
	client := &http.Client{}
	var request *http.Request
	
	if request, err = http.NewRequest("GET", uri, nil); err == nil {
		request.Header.Add("Accept", "application/json")
	} else {
		return
	}

	var response *http.Response
	if response, err = client.Do(request); err == nil {
		defer response.Body.Close()
	} else {
		return
	}

	body, err = ioutil.ReadAll(response.Body)
	status = response.StatusCode
	return
}



// Construct the query string
func queryString(beacon *Beacon, query *BeaconQuery) string {
	var k string
	var v string

	ql := make([]string, 0, 20)
	
	for _, d := range beacon.DatasetIds {
		ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap[DATASET_IDS], d))
	}
		
	for k, v = range *query {
		if k == ASSEMBLY_ID {
			ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap[k], beacon.AssemblyMap[v]))
		} else {
			ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap[k], v))
		}
	}

	for k, v = range beacon.AdditionalFields {
		ql = append(ql, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(ql, "&")
}



// Pose a given query to all of the configured beacons and await results
func queryBeaconsSync(query BeaconQuery) []byte {
	num := len(Config.Beacons)
	ch := make(chan BeaconResponse, num)
	responses := make([]BeaconResponse, 0, num)
	
	// Query each beacon
	for _, b := range Config.Beacons {
		go queryBeacon(b, &query, ch)
	}

	// Collect responses, or timeout
	for i := 0; i < num; i++ {
		select {
		case r := <-ch:
			responses = append(responses, r)
		case <- time.After(time.Second * time.Duration(timeout)):
			break
		}
	}

	// Turn into JSON
	j, _ := json.Marshal(responses)	
		
	return j
}



// Turn query into a friendly form
func friendlyQuery(query BeaconQuery) string {
	return fmt.Sprintf("%s:%s %s>%s", query[CHROMOSOME], query[START],
		query[REFERENCE_BASES], query[ALTERNATE_BASES])
}



// Handle HTTP query
func queryHandler(w http.ResponseWriter, r *http.Request) {
	query := extractParams(r, CHROMOSOME, START, REFERENCE_BASES, ALTERNATE_BASES, ASSEMBLY_ID)
	results := queryBeaconsSync(query)
	w.Header().Set("Content-Type", "application/json")
	w.Write(results)
}



// Extract query parameters into a map
func extractParams(r *http.Request, params ...string) (m map[string]string) {
	m = make(map[string]string)
	for _, k := range params {
		if (k != "") && (r.FormValue(k) != "") {
			m[k] = r.FormValue(k)
		}		
	}

	return
}



// Entry point
func main() {
	fmt.Printf("BoB is listening on port %d\n", port)
	http.HandleFunc("/", queryHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
