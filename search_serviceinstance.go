package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// ServiceInstanceSearchResults represents top level attributes of JSON response from Cloud Foundry API
type ServiceInstanceSearchResults struct {
	TotalResults int                              `json:"total_results"`
	TotalPages   int                              `json:"total_pages"`
	Resources    []ServiceInstanceSearchResources `json:"resources"`
}

// ServiceInstanceSearchResources represents resources attribute of JSON response from Cloud Foundry API
type ServiceInstanceSearchResources struct {
	Entity   ServiceInstanceSearchEntity `json:"entity"`
	Metadata Metadata                    `json:"metadata"`
}

// ServiceInstanceSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type ServiceInstanceSearchEntity struct {
	Name            string `json:"name"`
	SpaceGuid       string `json:"space_guid"`
	ServicePlanGuid string `json:"service_plan_guid"`
	Type            string `json:"type"`
}

// GetServiceInstanceData requests all of the Service Instance data from Cloud Foundry
func (c Events) GetServiceInstances(cli plugin.CliConnection) map[string]ServiceInstanceSearchEntity {
	var data = make(map[string]ServiceInstanceSearchEntity)
	services := c.GetServiceInstanceData(cli)

	for _, val := range services.Resources {
		data[val.Metadata.GUID] = val.Entity
	}

	return data
}

// GetServiceInstanceData requests all of the Service Instance data from Cloud Foundry
func (c Events) GetServiceInstanceData(cli plugin.CliConnection) ServiceInstanceSearchResults {
	var res ServiceInstanceSearchResults
	res = c.UnmarshallServiceInstanceSearchResults("/v2/service_instances?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/services_instances?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallServiceInstanceSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallServiceInstanceSearchResults(apiUrl string, cli plugin.CliConnection) ServiceInstanceSearchResults {
	var tRes ServiceInstanceSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
