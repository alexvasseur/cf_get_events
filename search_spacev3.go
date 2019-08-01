package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

// SpaceSearchResults represents top level attributes of JSON response from Cloud Foundry API
type SpaceSearchResultsv3 struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
		TotalPages   int `json:"total_pages"`
	} `json:"pagination"`
	Resources []SpaceSearchResourcesv3 `json:"resources"`
}

// SpaceSearchResources represents resources attribute of JSON response from Cloud Foundry API
type SpaceSearchResourcesv3 struct {
	Name     string `json:"name"`
	GUID     string `json:"guid"`
	Metadata struct {
		Labels map[string]string `json:"labels"`
	} `json:"metadata"`
	Relationships struct {
		Organization struct {
			Data struct {
				OrgGUID string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
	} `json:"relationships"`
}

func (ssr SpaceSearchResourcesv3) OrgGUID() string {
	return ssr.Relationships.Organization.Data.OrgGUID
}

// GetSearchSpaceData requests all of the Application data from Cloud Foundry
func (c Events) GetSearchSpacesv3(label_selector string, cli plugin.CliConnection) map[string]SpaceSearchResourcesv3 {
	if label_selector == "" {
		return c.GetSpacesv3(cli)
	} else {
		return c.SearchSpacesv3(label_selector, cli)
	}
}

// GetSpaceData requests all of the Application data from Cloud Foundry
func (c Events) GetSpacesv3(cli plugin.CliConnection) map[string]SpaceSearchResourcesv3 {
	var data = make(map[string]SpaceSearchResourcesv3)
	spaces := c.GetSpaceDatav3(cli)

	for _, val := range spaces.Resources {
		data[val.GUID] = val
	}

	return data
}

// GetSpaceData requests all of the Spaces data from Cloud Foundry
func (c Events) GetSpaceDatav3(cli plugin.CliConnection) SpaceSearchResultsv3 {
	var res SpaceSearchResultsv3
	res = c.UnmarshallSpaceSearchResultsv3("/v3/spaces?order_by=name&per_page=100", cli)

	if res.Pagination.TotalPages > 1 {
		for i := 2; i <= res.Pagination.TotalPages; i++ {
			apiUrl := fmt.Sprintf("/v3/spaces?order_by=name&page=%v&per_page=100", strconv.Itoa(i))
			tRes := c.UnmarshallSpaceSearchResultsv3(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

// GetSpaceData requests all of the Application data from Cloud Foundry
func (c Events) SearchSpacesv3(label_selector string, cli plugin.CliConnection) map[string]SpaceSearchResourcesv3 {
	var data = make(map[string]SpaceSearchResourcesv3)
	spaces := c.SearchSpaceDatav3(label_selector, cli)

	for _, val := range spaces.Resources {
		data[val.GUID] = val
	}

	return data
}

// GetSpaceData requests all of the Spaces data from Cloud Foundry
func (c Events) SearchSpaceDatav3(label_selector string, cli plugin.CliConnection) SpaceSearchResultsv3 {
	var res SpaceSearchResultsv3
	query := fmt.Sprintf("/v3/spaces?order_by=name&per_page=100&label_selector=%s", url.QueryEscape(label_selector))
	res = c.UnmarshallSpaceSearchResultsv3(query, cli)

	if res.Pagination.TotalPages > 1 {
		for i := 2; i <= res.Pagination.TotalPages; i++ {
			apiUrl := fmt.Sprintf("%s&page=%v", query, strconv.Itoa(i))
			tRes := c.UnmarshallSpaceSearchResultsv3(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
		}
	}

	return res
}

func (c Events) UnmarshallSpaceSearchResultsv3(apiUrl string, cli plugin.CliConnection) SpaceSearchResultsv3 {
	var tRes SpaceSearchResultsv3
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}

func (c Events) WriteSpaceLabel(guid string, labelKey string, labelValue string, cli plugin.CliConnection) string {
	apiUrl := fmt.Sprintf("/v3/spaces/%s", guid)
	json := fmt.Sprintf("'{    \"metadata\": { \"labels\": { \"%s\": \"%s\" } }  }'", labelKey, labelValue)
	//special delete
	if labelValue == "" {
		json = fmt.Sprintf("'{    \"metadata\": { \"labels\": { \"%s\":null } }  }'", labelKey)
	}
	cmd := []string{"curl", apiUrl, "-X", "PATCH", "-d", json}
	cli.CliCommandWithoutTerminalOutput(cmd...)

	return guid
}

func (c Events) ReadSpaceLabels(guid string, cli plugin.CliConnection) map[string]string {
	apiUrl := fmt.Sprintf("/v3/spaces/%s", guid)
	var tRes SpaceSearchResourcesv3
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes.Metadata.Labels
}
