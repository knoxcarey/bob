package beacon

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)


// Stores information about an individual beacon
type beaconStruct struct {
	Name              string                    // Name we give the beacon internally
	Version           string                    // Beacon API version
	Endpoint          string                    // URL for beacon
	Datasets          []string                  // Datasets to query
	AdditionalFields  map[string]string         // Additional query fields to include
	QueryMap          map[string]string         // Mapping standard names to query fields
}

// Type synonym for query
type BeaconQuery map[string][]string

// Contains the response from the beacon
type BeaconResponse struct {
	Name       string             `json:"name"` 
	Status     int                `json:"status"`
	Responses  map[string]string  `json:"responses,omitempty"` 
	Error      map[string]string  `json:"error,omitempty"`              
}

// Generic interface for beacons
type beacon interface {
	initialize()
	query(query *BeaconQuery, accessToken string, idToken string, ch chan<- BeaconResponse)
}

// List of beacons to be queried
var beacons []beacon

// Map containing types of beacons, keyed by version number string
var beaconType = map[string]reflect.Type{}


// Report number of beacons
func Count() int {
	return len(beacons)
}


// Read a configuration file, and create version-appropriate beacon structure
func AddBeaconFromConfig(file string) {

	// Read the configuration file
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("unable to read configuration file ", file)
	}

	// Unmarshal just enough to check the version
	var js map[string]interface{}
	if err = json.Unmarshal(buffer, &js); err != nil {
		log.Fatal("malformed config file ", file)
	}

	// Create an object of the appropriate version, cast as a generic beacon
	beacon := reflect.New(beaconType[js["version"].(string)]).Interface().(beacon)
	
	// Initialize it, giving it version-specific defaults
	beacon.initialize()

	// Unmarshal the rest of the structure, overriding defaults as necessary
	if err = json.Unmarshal(buffer, &beacon); err != nil {
		log.Fatal("malformed config file ", file)
	}

	// Add to the list of beacons to be queried
	beacons = append(beacons, beacon)
}


// Add an error condition to the response
func addResponseError(response *BeaconResponse, code int, message string) {
	response.Error["code"] = strconv.Itoa(code)
	response.Error["message"] = message
	response.Status = code
}



// Add a valid result to the response
func addResponseResult(response *BeaconResponse, key string, value string) {
	response.Responses[key] = value
}



// Wrapper for HTTP get
func httpGet(uri string, accessToken string, idToken string) (status int, body []byte, err error) {
	client := &http.Client{}
	var request *http.Request
	
	if request, err = http.NewRequest("GET", uri, nil); err == nil {
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Authorization", "Bearer " + accessToken)
		request.Header.Add("IDToken", idToken)
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



// Pose a given query to all of the configured beacons and await results
func QueryBeaconsSync(query BeaconQuery, accessToken string, idToken string, timeout int) []byte {
	num := len(beacons)
	ch := make(chan BeaconResponse, num)
	responses := make([]BeaconResponse, 0, num)

	// Query each beacon
	for _, b := range beacons {
		go b.query(&query, accessToken, idToken, ch)
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


// Query all beacons, writing results back to channel asynchronously
func QueryBeaconsAsync(query BeaconQuery, accessToken string, idToken string, ch chan<- BeaconResponse) {
	for _, b := range beacons {
		go b.query(&query, accessToken, idToken, ch)
	}	
}
