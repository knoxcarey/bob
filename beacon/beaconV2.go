package beacon

import (
	"encoding/json"
	"fmt"
	"strings"
)

type beaconV2 BeaconBase

func (beacon *beaconV2) initialize() {
	beacon.QueryMap = make(map[string]string)
	beacon.QueryMap["chromosome"]     = "chromosome"
	beacon.QueryMap["start"]          = "position"
	beacon.QueryMap["alternateBases"] = "allele"
	beacon.QueryMap["referenceBases"] = "referenceBases"
	beacon.QueryMap["datasetIds"]     = "dataset"
	beacon.QueryMap["assemblyId"]     = "reference"
	beacon.QueryMap["GRCh37"]         = "GRCh37"
	beacon.QueryMap["GRCh38"]         = "GRCh38"
}

func (beacon *beaconV2) parseResponse(status int, raw []byte, err error) *BeaconResponse {
	response := &BeaconResponse{Name: beacon.Name}	

	if err != nil {
		addResponseError(response, 400, "could not reach beacon")
		return response		
	}

	if status/100 != 2 {
		addResponseError(response, status, "beacon error")
		return response
	}

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

	return response
}


func (beacon *beaconV2) queryBeacon(query *BeaconQuery, ch chan<- BeaconResponse) {
	qs := beacon.queryString(query)
	uri := fmt.Sprintf("%s?%s", beacon.Endpoint, qs)

	status, body, err := httpGet(uri)	
	resp := beacon.parseResponse(status, body, err)

	ch <- *resp
}


// Construct the query string
func (beacon *beaconV2) queryString(query *BeaconQuery) string {
	ql := make([]string, 0, 20)
	
	// Add datasets
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

	for k, v := range beacon.AdditionalFields {
		ql = append(ql, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(ql, "&")
}