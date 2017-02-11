package beacon

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Type alias for this version
type beaconV3 BeaconType

// Register this version's type
func init() {
	var nilStruct *beaconV3
	beaconType["0.3"] = reflect.TypeOf(nilStruct).Elem()
}

// Initialize the beacon with defaults appropriate for this version
func (beacon *beaconV3) initialize() {
	beacon.QueryMap = make(map[string]string)
	beacon.QueryMap["chromosome"]     = "chromosome"
	beacon.QueryMap["start"]          = "start"
	beacon.QueryMap["alternateBases"] = "alternateBases"
	beacon.QueryMap["referenceBases"] = "referenceBases"
	beacon.QueryMap["datasetIds"]     = "datasetIds"
	beacon.QueryMap["assemblyId"]     = "assemblyId"
	beacon.QueryMap["GRCh37"]         = "GRCh37"
	beacon.QueryMap["GRCh38"]         = "GRCh38"
}


func (beacon *beaconV3) parseResponse(status int, raw []byte, err error) *BeaconResponse {
	response := &BeaconResponse{Name: beacon.Name}	

	if err != nil {
		addResponseError(response, 400, "could not reach beacon")
		return response		
	}

	if status/100 != 2 {
		addResponseError(response, status, "beacon error")
		return response
	}

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

	return response
}


func (beacon *beaconV3) query(query *BeaconQuery, ch chan<- BeaconResponse) {
	qs := beacon.queryString(query)
	uri := fmt.Sprintf("%s?%s", beacon.Endpoint, qs)

	status, body, err := httpGet(uri)	
	resp := beacon.parseResponse(status, body, err)

	ch <- *resp
}


// Construct the query string
func (beacon *beaconV3) queryString(query *BeaconQuery) string {
	ql := make([]string, 0, 20)
	
	for _, d := range beacon.Datasets {
		ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap["datasetIds"], d))
	}
		
	for k, v := range *query {
		if k == "assemblyId" {
			ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap[k], beacon.QueryMap[v[0]]))
		} else {
			ql = append(ql, fmt.Sprintf("%s=%s", beacon.QueryMap[k], v[0]))
		}
	}

	for k2, v2 := range beacon.AdditionalFields {
		ql = append(ql, fmt.Sprintf("%s=%s", k2, v2))
	}

	return strings.Join(ql, "&")
}


