// fetch_usage.go
package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

type CFInfo struct {
}

type MonthlyUsage struct {
	MonthUsage []MonthUsage `json:"monthly_reports"`
}

type TaskMonthlyUsage struct {
	TaskMonthUsage []TaskMonthUsage `json:"monthly_reports"`
}

type MonthUsage struct {
	Year              int     `json:"year"`
	Month             int     `json:"month"`
	Avg               float32 `json:"average_app_instances"`
	Max               int     `json:"maximum_app_instances"`
	TaskMaxConcurrent int     // we will merge from Tasks
	TaskTotalRun      int     // we will merge from Tasks
}

type TaskMonthUsage struct {
	Year          int `json:"year"`
	Month         int `json:"month"`
	MaxConcurrent int `json:"maximum_concurrent_tasks"`
	TotalRun      int `json:"total_task_runs"`
}

func (c Events) GetMonthlyUsage(cli plugin.CliConnection) []MonthUsage {
	var res MonthlyUsage
	var res2 TaskMonthlyUsage
	OLDEST := 7
	months := make([]MonthUsage, OLDEST) // retrieve only the 7 latest months

	// find system api and replace with app-usage PCF system app
	api, _ := cli.ApiEndpoint()
	appUsage := strings.Replace(api, "https://api.", "https://app-usage.", 1)

	// prepare OAuth token
	token, _ := cli.AccessToken()

	// ignore TLS cert and use a 20s timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{
		Transport: tr,
		Timeout:   time.Second * 20,
	}

	// AI usage
	req, _ := http.NewRequest("GET", appUsage+"/system_report/app_usages", nil)
	req.Header.Add("authorization", token)
	response, _ := client.Do(req)
	if response.StatusCode == 200 {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, &res)
	}

	// Task usage
	req, _ = http.NewRequest("GET", appUsage+"/system_report/task_usages", nil)
	req.Header.Add("authorization", token)
	response, _ = client.Do(req)
	if response.StatusCode == 200 {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, &res2)
	}

	// error handling should be done here if res and res2 are not populated

	// Keep most recent 7 - the API returns oldest to most recent
	for k := range res.MonthUsage {
		kk := len(res.MonthUsage) - 1 - k
		if k > OLDEST-1 {
			break
		}

		months[k] = res.MonthUsage[kk]

		// search in Task the corresponding year / month
		for _, t := range res2.TaskMonthUsage {
			if months[k].Year == t.Year && months[k].Month == t.Month {
				months[k].TaskMaxConcurrent = t.MaxConcurrent
				months[k].TaskTotalRun = t.TotalRun
			}
		}
	}

	return months
}
