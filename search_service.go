package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

// ServiceSearchResults represents top level attributes of JSON response from Cloud Foundry API
type ServiceSearchResults struct {
	TotalResults int                      `json:"total_results"`
	TotalPages   int                      `json:"total_pages"`
	Resources    []ServiceSearchResources `json:"resources"`
}

// ServiceSearchResources represents resources attribute of JSON response from Cloud Foundry API
type ServiceSearchResources struct {
	Entity   ServiceSearchEntity `json:"entity"`
	Metadata Metadata            `json:"metadata"`
}

// ServiceSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type ServiceSearchEntity struct {
	Label string `json:"label"`
}

// GetServiceData requests all of the Service data from Cloud Foundry
func (c Events) GetServices(cli plugin.CliConnection) map[string]ServiceSearchEntity {
	var data = make(map[string]ServiceSearchEntity)
	services := c.GetServiceData(cli)

	for _, val := range services.Resources {
		data[val.Metadata.GUID] = val.Entity //((ServiceSearchEntity)(val.Entity)) //   Label
	}

	return data
}

// GetServiceData requests all of the Service data from Cloud Foundry
func (c Events) GetServiceData(cli plugin.CliConnection) ServiceSearchResults {
	var res ServiceSearchResults
	res = c.UnmarshallServiceSearchResults("/v2/services?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/services?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallServiceSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallServiceSearchResults(apiUrl string, cli plugin.CliConnection) ServiceSearchResults {
	var tRes ServiceSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
