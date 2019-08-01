package main

import (
	"encoding/json"
	"fmt"
	//	"sort"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

// OrgSearchResults represents top level attributes of JSON response from Cloud Foundry API
type OrgSearchResults struct {
	TotalResults int                  `json:"total_results"`
	TotalPages   int                  `json:"total_pages"`
	Resources    []OrgSearchResources `json:"resources"`
}

// OrgSearchResources represents resources attribute of JSON response from Cloud Foundry API
type OrgSearchResources struct {
	Entity   OrgSearchEntity `json:"entity"`
	Metadata Metadata        `json:"metadata"`
}

// OrgSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type OrgSearchEntity struct {
	Name      string `json:"name"`
	QuotaGuid string `json:"quota_definition_guid"`
}

func (c Events) GetOrgs(cli plugin.CliConnection) map[string]OrgSearchEntity {
	var data = make(map[string]OrgSearchEntity)
	orgs := c.GetOrgData(cli)

	for _, val := range orgs.Resources {
		data[val.Metadata.GUID] = val.Entity
	}

	return data
}

// GetOrgData requests all of the Application data from Cloud Foundry
func (c Events) GetOrgData(cli plugin.CliConnection) OrgSearchResults {
	var res OrgSearchResults
	res = c.UnmarshallOrgSearchResults("/v2/organizations?order-direction=asc&results-per-page=100", cli)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v2/organizations?order-direction=asc&page=%v&results-per-page=100", strconv.Itoa(i))
			tRes := c.UnmarshallOrgSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	//sort by org name - TODO
	/*
		sort.Slice(&res.Resources, func(i, j int) bool {
			switch strings.Compare(strings.ToLower(res.Resources[i].Entity.Name), strings.ToLower(res.Resources[j].Entity.Name)) {
			case -1:
				return true
			case 1:
				return false
			}
			return true
		})
	*/
	return res
}

func (c Events) UnmarshallOrgSearchResults(apiUrl string, cli plugin.CliConnection) OrgSearchResults {
	var tRes OrgSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}
