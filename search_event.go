package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"time"

	"github.com/cloudfoundry/cli/plugin"
)

// OutputResults represents the filtered event results for the input args
type OutputResults struct {
	Comment   string                 `json:"comment"`
	Resources []EventSearchResources `json:"resources"`
}

// EventSearchResults represents top level attributes of JSON response from Cloud Foundry API
type EventSearchResults struct {
	TotalResults int                    `json:"total_results"`
	TotalPages   int                    `json:"total_pages"`
	PrevUrl      string                 `json:"prev_url"`
	NextUrl      string                 `json:"next_url"`
	Resources    []EventSearchResources `json:"resources"`
}

// EventSearchResources represents resources attribute of JSON response from Cloud Foundry API
type EventSearchResources struct {
	Entity   EventSearchEntity `json:"entity"`
	Metadata Metadata          `json:"metadata"`
}

// EventSearchEntity represents entity attribute of resources attribute within JSON response from Cloud Foundry API
type EventSearchEntity struct {
	Type      string                    `json:"type"`
	Actor     string                    `json:"actor"`
	ActorType string                    `json:"actor_type"`
	ActorName string                    `json:"actor_name"`
	Actee     string                    `json:"actee"`
	ActeeType string                    `json:"acte_type"`
	ActeeName string                    `json:"actee_name"`
	Timestamp string                    `json:"timestamp"`
	Metadata  EventSearchEntityMetadata `json:"metadata"`
	SpaceGUID string                    `json:"space_guid"`
	OrgGUID   string                    `json:"organization_guid"`
	Space     string                    `json:"space"`
	Org       string                    `json:"org"`
}

type EventSearchEntityMetadata struct {
	Instance        string             `json:"instance,omitempty"`
	Index           int                `json:"index,omitempty"`
	ExitDescription string             `json:"exit_description,omitempty"`
	Reason          string             `json:"reason,omitempty"`
	Request         ESEMetadataRequest `json:"request,omitempty"`
}

type ESEMetadataRequest struct {
	State     string `json:"state"`
	Recursive string `json:"recursive"`
}

// GetEventsData requests all of the application events from Cloud Foundry
func (c Events) GetEventsData(cli plugin.CliConnection, ins Inputs) EventSearchResults {
	var res EventSearchResults

	// var baseUrl = "/v2/events?order-direction=asc&results-per-page=100"
	var baseUrl = "/v2/events?order-direction=desc&results-per-page=100"
	// filter date was passed in. Append to the query.
	var filterurl = "&q=timestamp%3E" + fmt.Sprintf("%s", ins.fromDate.Format("2006-01-02T00:00:00Z"))

	// the pattern is:  "/v2/events?order-direction=desc&results-per-page=100&page=%v&q=timestamp%3E2016-12-10"
	// the first url:
	//	"/v2/events?order-direction=asc&results-per-page=100&page=1&q=timestamp%3E2016-09-20T00:00:00Z"
	// 	"/v2/events?order-direction=desc&results-per-page=100&page=1&q=timestamp%3E2016-12-10"	or
	// 	"/v2/events?order-direction=desc&results-per-page=100&page=1&q=timestamp%3E2016-12-10&q=timestamp%3E2016-12-14"

	var url = fmt.Sprintf("%s&page=%v%s", baseUrl, "1", filterurl)
	res = c.UnmarshallEventSearchResults(url, cli)

	//	fmt.Println("---------------> DEBUG-1 :%s", url)
	//	fmt.Println("---------------> # of entities: ", len(res.Resources), res.TotalPages)

	if res.TotalPages > 1 {
		for i := 2; i <= res.TotalPages; i++ {
			// apiUrl := fmt.Sprintf("/v2/events?order-direction=desc&page=%v&results-per-page=100&q=timestamp%3E2016-12-10", strconv.Itoa(i))
			apiUrl := fmt.Sprintf("%s&page=%v%s", baseUrl, strconv.Itoa(i), filterurl)
			tRes := c.UnmarshallEventSearchResults(apiUrl, cli)
			res.Resources = append(res.Resources, tRes.Resources...)
			//	fmt.Println("---------------> DEBUG-%i :%s", i, apiUrl)
			//	fmt.Println("---------------> # of entities: ", len(res.Resources), res.TotalPages)
		}
	}

	return res
}

func (c Events) UnmarshallEventSearchResults(apiUrl string, cli plugin.CliConnection) EventSearchResults {
	var tRes EventSearchResults
	cmd := []string{"curl", apiUrl}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &tRes)

	return tRes
}

// filter the results for given input criteria.
func (c Events) FilterResults(cli plugin.CliConnection, ins Inputs, orgs map[string]OrgSearchEntity, spaces map[string]SpaceSearchEntity, apps AppSearchResults, events EventSearchResults) OutputResults {
	var results OutputResults

	results.Comment = fmt.Sprintf("Following events were recorded from '%s', to '%s' \n\n", ins.fromDate, ins.toDate)

	for _, val := range events.Resources {

		evTmsp, _ := time.Parse(time.RFC3339, val.Entity.Timestamp)
		// fmt.Println("timestamps: ", evTmsp.Nanosecond(), filterDate.Nanosecond(), )

		if evTmsp.Before(ins.fromDate) {
			// all events are retrieved in descending order.
			// we processed all events that are filterDate onwards
			// reached older events, break out
			// 	fmt.Println("event date: ", evTmsp, "filterDate: ", filterDate )
			break
		}

		if evTmsp.After(ins.toDate) {
			continue
		}

		var outEntity EventSearchResources

		outEntity.Metadata = val.Metadata
		outEntity.Entity.Type = val.Entity.Type
		outEntity.Entity.Actor = val.Entity.Actor
		outEntity.Entity.ActorType = val.Entity.ActorType
		outEntity.Entity.ActorName = val.Entity.ActorName
		outEntity.Entity.Actee = val.Entity.Actee
		outEntity.Entity.ActeeType = val.Entity.ActeeType
		outEntity.Entity.ActeeName = val.Entity.ActeeName
		outEntity.Entity.Timestamp = val.Entity.Timestamp
		outEntity.Entity.Metadata = val.Entity.Metadata
		outEntity.Entity.SpaceGUID = val.Entity.SpaceGUID
		outEntity.Entity.OrgGUID = val.Entity.OrgGUID
		outEntity.Entity.Org = orgs[val.Entity.OrgGUID].Name
		outEntity.Entity.Space = spaces[val.Entity.SpaceGUID].Name
		results.Resources = append(results.Resources, outEntity)
	}
	return results
}
