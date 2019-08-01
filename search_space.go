package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

// SpaceSearchResults represents top level attributes of JSON response from Cloud Foundry API
type SpaceSearchResults struct {
	TotalResults int                    `json:"total_results"`
	TotalPages   int                    `json:"total_pages"`
	Resources    []SpaceSearchResources `json:"resources"`
}

// SpaceSearchResources represents resources attribute of JSON response from Cloud Foundry API
type SpaceSearchResources struct {
	Entity   SpaceSearchEntity `json:"entity"`
	Metadata Metadata          `json:"metadata"`
}

// SpaceSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type SpaceSearchEntity struct {
	Name    string `json:"name"`
	OrgGUID string `json:"organization_guid"`
}

// GetSpaceData requests all of the Application data from Cloud Foundry
func (c Events) GetSpaces(cli plugin.CliConnection) map[string]SpaceSearchEntity {
	var data = make(map[string]SpaceSearchEntity)
	spaces := c.GetSpaceData(cli)

	for _, val := range spaces.Resources {
		data[val.Metadata.GUID] = val.Entity
	}

	return data
}

// GetSpaceData requests all of the Spaces data from Cloud Foundry
func (c Events) GetSpaceData(cli plugin.CliConnection) SpaceSearchResults {
	var res SpaceSearchResults
	res = c.UnmarshallSpaceSearchResults("/v2/spaces?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/spaces?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallSpaceSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallSpaceSearchResults(apiUrl string, cli plugin.CliConnection) SpaceSearchResults {
	var tRes SpaceSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
