package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

// AppSearchResults represents top level attributes of JSON response from Cloud Foundry API
type AppSearchResults struct {
	TotalResults int                  `json:"total_results"`
	TotalPages   int                  `json:"total_pages"`
	PrevUrl      string               `json:"prev_url"`
	NextUrl      string               `json:"next_url"`
	Resources    []AppSearchResources `json:"resources"`
}

// AppSearchResources represents resources attribute of JSON response from Cloud Foundry API
type AppSearchResources struct {
	Entity   AppSearchEntity `json:"entity"`
	Metadata Metadata        `json:"metadata"`
}

// AppSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type AppSearchEntity struct {
	Name              string `json:"name"`
	Buildpack         string `json:"buildpack"`
	DetectedBuildpack string `json:"detected_buildpack"`
	SpaceGUID         string `json:"space_guid"`
	Instances         int    `json:"instances"`
	State             string `json:"state"` // STARTED, STOPPED, ??TODO
	Memory            int    `json:"memory"`
	DiskQuota         int    `json:"disk_quota"`
}

// util to remove i in the array
func (c Events) remove(slice []AppSearchResources, i int) []AppSearchResources {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

//// GetAppData requests all of the Application data from Cloud Foundry
//func (c Events) GetApps(cli plugin.CliConnection) map[string]AppSearchResults {
//	var data map[string]AppSearchResults
//	data = make(map[string]AppSearchResults)
//	spaces := c.GetAppData(cli)
//
//	for _, val := range spaces.Resources {
//		data[val.Metadata.GUID] = val.Metadata
//	}
//
//	return data
//}

// GetAppData requests all of the Application data from Cloud Foundry
func (c Events) GetAppData(cli plugin.CliConnection) AppSearchResults {
	var res AppSearchResults
	res = c.UnmarshallAppSearchResults("/v2/apps?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/apps?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallAppSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallAppSearchResults(apiUrl string, cli plugin.CliConnection) AppSearchResults {
	var tRes AppSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
