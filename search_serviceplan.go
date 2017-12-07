package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// ServicePlanSearchResults represents top level attributes of JSON response from Cloud Foundry API
type ServicePlanSearchResults struct {
	TotalResults int                          `json:"total_results"`
	TotalPages   int                          `json:"total_pages"`
	Resources    []ServicePlanSearchResources `json:"resources"`
}

// ServicePlanSearchResources represents resources attribute of JSON response from Cloud Foundry API
type ServicePlanSearchResources struct {
	Entity   ServicePlanSearchEntity `json:"entity"`
	Metadata Metadata                `json:"metadata"`
}

// ServicePlanSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type ServicePlanSearchEntity struct {
	Name        string `json:"name"`
	ServiceGuid string `json:"service_guid"`
}

// GetServicePlanData requests all of the Service data from Cloud Foundry
func (c Events) GetServicePlans(cli plugin.CliConnection) map[string]ServicePlanSearchEntity {
	var data = make(map[string]ServicePlanSearchEntity)
	services := c.GetServicePlanData(cli)

	for _, val := range services.Resources {
		data[val.Metadata.GUID] = val.Entity
	}

	return data
}

// GetServicePlanData requests all of the Service data from Cloud Foundry
func (c Events) GetServicePlanData(cli plugin.CliConnection) ServicePlanSearchResults {
	var res ServicePlanSearchResults
	res = c.UnmarshallServicePlanSearchResults("/v2/service_plans?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/services_plans?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallServicePlanSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallServicePlanSearchResults(apiUrl string, cli plugin.CliConnection) ServicePlanSearchResults {
	var tRes ServicePlanSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
