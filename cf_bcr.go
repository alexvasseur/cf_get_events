// Copyright (c) 2016 ECS Team, Inc. - All Rights Reserved
// https://github.com/ECSTeam/cloudfoundry-top-plugin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/olekukonko/tablewriter"
	"github.com/simonleung8/flags"
)

// Events represents Buildpack Usage CLI interface
type Events struct{}

// Metadata is the data retrived from the response json
type Metadata struct {
	GUID string `json:"guid"`
}

// Inputs represent the parsed input args
type Inputs struct {
	fromDate time.Time
	toDate   time.Time
	isCsv    bool
	isJson   bool
	AI       bool
	SI       bool
	monthly  bool
}

type Total struct {
	org            int
	space          int
	app            int
	appUser        int
	appStarted     int
	appUserStarted int
	AI             int
	AIStarted      int
	AIUser         int
	AIUserStarted  int
	mem            int
	memStarted     int
	memUser        int
	memUserStarted int
	si             int
	siMySQL        int
	siRabbitMQ     int
	siRedis        int
	siOther        int
}

// GetMetadata provides the Cloud Foundry CLI with metadata to provide user about how to use `bcr` command
func (c *Events) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "bcr",
		Version: plugin.VersionType{
			Major: 2,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "bcr",
				HelpText: "Get Apps and Services consumption details",
				UsageDetails: plugin.Usage{
					Usage: UsageText(),
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(Events))
}

// Run is what is executed by the Cloud Foundry CLI when the bcr command is specified
func (c Events) Run(cli plugin.CliConnection, args []string) {
	var ins Inputs

	switch args[0] {
	case "bcr":
		if len(args) >= 2 {
			ins = c.buildClientOptions(args)
		} else {
			Usage(1)
		}
	default:
		Usage(0)
	}

	if ins.monthly {
		month := c.GetMonthlyUsage(cli)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Year", "Month", "AI avg", "AI max", "Task concurrent", "Task total runs"})
		for _, m := range month {
			table.Append([]string{
				strconv.Itoa(m.Year), strconv.Itoa(m.Month),
				fmt.Sprintf("%.0f", m.Avg), strconv.Itoa(m.Max),
				strconv.Itoa(m.TaskMaxConcurrent), strconv.Itoa(m.TaskTotalRun)})
		}
		table.Render()
	}

	orgs := c.GetOrgs(cli)
	spaces := c.GetSpaces(cli)
	var total Total
	total.org = len(orgs)
	total.space = len(spaces)

	var services map[string]ServiceSearchEntity
	var plans map[string]ServicePlanSearchEntity
	var serviceInstances map[string]ServiceInstanceSearchEntity
	var apps AppSearchResults

	// Data loading
	if ins.SI {
		services = c.GetServices(cli)
		plans = c.GetServicePlans(cli)
		serviceInstances = c.GetServiceInstances(cli)
	}
	if ins.AI {
		apps = c.GetAppData(cli)
	}

	// services instances -- DEBUG only
	if ins.SI && false {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Org", "Space", "Service Instance"})
		//TODO loop on org, space id
		for _, si := range serviceInstances {
			table.Append([]string{orgs[spaces[si.SpaceGuid].OrgGUID].Name, spaces[si.SpaceGuid].Name, si.Name})
		}
		table.Render()

	}

	// sort orgs by org Name
	i := 0
	sortedOrgs := make([]string, len(orgs))
	for k := range orgs {
		sortedOrgs[i] = k
		i++
	}
	sort.Slice(sortedOrgs, func(i, j int) bool {
		switch strings.Compare(strings.ToLower(orgs[sortedOrgs[i]].Name), strings.ToLower(orgs[sortedOrgs[j]].Name)) {
		case -1:
			return true
		case 1:
			return false
		}
		return true
	})
	//TODO sort space by Name and use it below
	//TODO for some reasons space is at least grouped?

	// *** SI table
	if ins.SI {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Org", "Space", "SI", "Pivotal MySQL", "Pivotal RabbitMQ", "Pivotal Redis", "Other Services"})
		//TODO - BROKERAGE, other?
		for _, oguid := range sortedOrgs {
			for sguid, space := range spaces {
				if space.OrgGUID == oguid {

					siSpace := 0
					siList := make(map[string]int)

					for _, si := range serviceInstances {
						if si.SpaceGuid == sguid && si.Type == "managed_service_instance" {
							siSpace++
							siList[services[plans[si.ServicePlanGuid].ServiceGuid].Label /*+":"+plans[si.ServicePlanGuid].Name*/] += 1
						}
					}
					flat := []string{}
					siMySQL := 0
					siRabbitMQ := 0
					siRedis := 0
					siOther := 0
					for n, c := range siList {
						total.si += c
						switch n {
						case "p-mysql":
							siMySQL += c
							total.siMySQL += c
						case "p.mysql":
							siMySQL += c
							total.siMySQL += c
						case "p-rabbitmq":
							siRabbitMQ += c
							total.siRabbitMQ += c
						case "p.rabbitmq":
							siRabbitMQ += c
							total.siRabbitMQ += c
						case "p-redis":
							siRedis += c
							total.siRedis += c
						case "p.redis":
							siRedis += c
							total.siRedis += c
						default:
							siOther += c
							total.siOther += c
							flat = append(flat, n+":"+strconv.Itoa(c))
						}
						//flat = append(flat, n+":"+strconv.Itoa(c))
					}
					siMySQLstring := ""
					if siMySQL > 0 {
						siMySQLstring = strconv.Itoa(siMySQL)
					}
					siRabbitMQstring := ""
					if siRabbitMQ > 0 {
						siRabbitMQstring = strconv.Itoa(siRabbitMQ)
					}
					siRedisstring := ""
					if siRedis > 0 {
						siRedisstring = strconv.Itoa(siRedis)
					}
					table.Append([]string{orgs[oguid].Name, spaces[sguid].Name, strconv.Itoa(siSpace),
						siMySQLstring, siRabbitMQstring, siRedisstring,
						strings.Join(flat, ",")})

				}
			}
		}
		//TODO total SI and Pivotal SI
		table.SetFooter([]string{"-", "-", strconv.Itoa(total.si), strconv.Itoa(total.siMySQL), strconv.Itoa(total.siRabbitMQ), strconv.Itoa(total.siRedis), strconv.Itoa(total.siOther) + " (Pivotal: " + strconv.Itoa(total.si-total.siOther) + ")"})
		table.Render()
	}

	// *** SI summary
	// sort service by service Label
	i = 0
	sortedServices := make([]string, len(services))
	for k := range services {
		sortedServices[i] = k
		i++
	}
	sort.Slice(sortedServices, func(i, j int) bool {
		switch strings.Compare(strings.ToLower(services[sortedServices[i]].Label), strings.ToLower(services[sortedServices[j]].Label)) {
		case -1:
			return true
		case 1:
			return false
		}
		return true
	})
	if ins.SI {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service", "Plan", "Service Instances"})
		for _, guid := range sortedServices {
			for planGuid, plan := range plans {
				if plan.ServiceGuid == guid {
					var count = 0
					for _, si := range serviceInstances {
						if si.ServicePlanGuid == planGuid {
							count++
						}
					}
					table.Append([]string{services[plan.ServiceGuid].Label, plan.Name, strconv.Itoa(count)})
				}
			}
		}
		table.Render()
	}

	// *** APP table
	if ins.AI {
		total.app = len(apps.Resources)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Org", "Space", "App", "AI", "Memory", "State", "Memory usage"})

		// order by Orgs, then by Space, then by State
		for _, oguid := range sortedOrgs {
			for sguid, space := range spaces {
				if space.OrgGUID == oguid {

					// count non system only
					if orgs[oguid].Name != "system" { //&& orgs[oguid] != "p-spring-cloud-services" {
						for _, val := range apps.Resources {
							if val.Entity.SpaceGUID == sguid {
								total.appUser++
								total.AIUser += val.Entity.Instances
								total.memUser += val.Entity.Instances * val.Entity.Memory
								if val.Entity.State == "STARTED" {
									total.AIUserStarted += val.Entity.Instances
									total.appUserStarted++
									total.memUserStarted += val.Entity.Instances * val.Entity.Memory
								}
							}
						}
					}

					// STARTED first
					for _, val := range apps.Resources {
						if val.Entity.SpaceGUID == sguid && val.Entity.State == "STARTED" {

							total.AI += val.Entity.Instances
							total.mem += val.Entity.Instances * val.Entity.Memory

							total.AIStarted += val.Entity.Instances
							total.memStarted += val.Entity.Instances * val.Entity.Memory
							total.appStarted++

							memUsage := val.Entity.Instances * val.Entity.Memory

							table.Append([]string{orgs[spaces[val.Entity.SpaceGUID].OrgGUID].Name, spaces[val.Entity.SpaceGUID].Name, val.Entity.Name,
								strconv.Itoa(val.Entity.Instances), strconv.Itoa(val.Entity.Memory), val.Entity.State, strconv.Itoa(memUsage)})
							//fmt.Printf("%s,%s,%s,%d,%d,%s\n",
							//	orgs[spaces[val.Entity.SpaceGUID].OrgGUID], spaces[val.Entity.SpaceGUID].Name, val.Entity.Name,
							//	val.Entity.Instances, val.Entity.Memory, val.Entity.State)
						}
					}
					// any other state then
					for _, val := range apps.Resources {
						if val.Entity.SpaceGUID == sguid && val.Entity.State != "STARTED" {

							total.AI += val.Entity.Instances
							total.mem += val.Entity.Instances * val.Entity.Memory

							table.Append([]string{orgs[spaces[val.Entity.SpaceGUID].OrgGUID].Name, spaces[val.Entity.SpaceGUID].Name, val.Entity.Name,
								strconv.Itoa(val.Entity.Instances), strconv.Itoa(val.Entity.Memory), val.Entity.State, ""})
							//fmt.Printf("%s,%s,%s,%d,%d,%s\n",
							//	orgs[spaces[val.Entity.SpaceGUID].OrgGUID], spaces[val.Entity.SpaceGUID].Name, val.Entity.Name,
							//	val.Entity.Instances, val.Entity.Memory, val.Entity.State)
						}
					}
				}
			}
		}

		table.SetFooter([]string{strconv.Itoa(total.org), strconv.Itoa(total.space), strconv.Itoa(total.app), strconv.Itoa(total.AI), strconv.Itoa(total.mem), strconv.Itoa(total.appStarted) + " (started)", strconv.Itoa(total.memStarted)})
		table.SetFooter([]string{"-", "-", "-", "-", "-", "-"})
		table.Render()

		// org mem usage and AI running per org
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Org", "Memory Limit", "Memory Usage", "Usage %", "AI (started)"})
		var orgsSummary = c.GetOrgsSummary(cli)
		for _, oguid := range sortedOrgs {
			val := orgsSummary[oguid]
			aicount := 0
			for spaceGuid, space := range spaces {
				if space.OrgGUID == oguid {
					for _, app := range apps.Resources {
						if app.Entity.SpaceGUID == spaceGuid && app.Entity.State == "STARTED" {
							aicount += app.Entity.Instances
						}
					}
				}
			}

			table.Append([]string{val.Name, strconv.Itoa(val.MemoryLimitOrgQuota), strconv.Itoa(val.Memory), strconv.Itoa(val.MemoryUsage), strconv.Itoa(aicount)})
		}
		table.Render()

		// summary table
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Category", "App", "AI", "Memory"})
		table.Append([]string{"Total", strconv.Itoa(total.app), strconv.Itoa(total.AI), strconv.Itoa(total.mem)})
		table.Append([]string{"Total (excl system)", strconv.Itoa(total.appUser), strconv.Itoa(total.AIUser), strconv.Itoa(total.memUser)})
		table.Append([]string{"STARTED", strconv.Itoa(total.appStarted), strconv.Itoa(total.AIStarted), strconv.Itoa(total.memStarted)})
		table.Append([]string{"STARTED (excl system)", strconv.Itoa(total.appUserStarted), strconv.Itoa(total.AIUserStarted), strconv.Itoa(total.memUserStarted)})
		table.Render()

	}

	if false {
		events := c.GetEventsData(cli, ins)
		c.FilterResults(cli, ins, orgs, spaces, apps, events)
		//results := c.FilterResults(cli, ins, orgs, spaces, apps, events)
		/*
			if ins.isCsv {
				c.EventsInCSVFormat(results)
			} else {
				c.EventsInJsonFormat(results)
			}
		*/
	}
}

func Usage(code int) {
	fmt.Println("\nUsage: ", UsageText())
	os.Exit(code)
}

func UsageText() string {
	/*	usage := "cf get-events [options]" +
		"\n    where options include: " +
		"\n       --today                  : get all events for today (till now)" +
		"\n       --yesterday              : get events for yesterday only" +
		"\n       --yesterday-on           : get events from yesterday onwards (till now)" +
		"\n       --all                    : get all events (defaults to last 90 days)" +
		"\n       --json                   : list output in json format (default is csv)\n" +
		"\n       --from <yyyymmdd>        : get events from given date onwards (till now)" +
		"\n       --from <yyyymmddhhmmss>  : get events from given date and time onwards (till now)" +
		"\n       --to <yyyymmdd>          : get events till given date" +
		"\n       --to <yyyymmddhhmmss>    : get events till given date and time\n" +
		"\n       --from <yyyymmdd> --to <yyyymmdd>" +
		"\n       --from <yyyymmddhhmmss> --to <yyyymmddhhmmss>"
	*/
	usage := "cf bcr [options]" +
		"\n	--ai" +
		"\n	--si" +
		"\n	--monthly"
	return usage
}

func GetStartOfDay(today time.Time) time.Time {
	var now = fmt.Sprintf("%s", today.Format("2006-01-02"))
	t, _ := time.Parse(time.RFC3339, now+"T00:00:00Z")
	return t
}

func GetEndOfDay(today time.Time) time.Time {
	var now = fmt.Sprintf("%s", today.Format("2006-01-02"))
	t, _ := time.Parse(time.RFC3339, now+"T23:59:59Z")
	return t
}

// sanitize data by replacing \r, and \n with ';'
func sanitize(data string) string {
	var re = regexp.MustCompile(`\r?\n`)
	var str = re.ReplaceAllString(data, ";")
	str = strings.Replace(str, ";;", ";", 1)
	return str
}

// read arguments passed for the plugin
func (c *Events) buildClientOptions(args []string) Inputs {
	fc := flags.New()
	/*
		fc.NewBoolFlag("all", "all", " get all events (defaults to last 90 days)")
		fc.NewBoolFlag("today", "today", "get all events for today (till now)")
		fc.NewBoolFlag("yesterday", "yest", "get events from yesterday only")
		fc.NewBoolFlag("yesterday-on", "yon", "get events for yesterday onwards (till now)")
		fc.NewStringFlag("from", "fr", "get events from given date [+ time] onwards (till now)")
		fc.NewStringFlag("to", "to", "get events till given date [+ time]")
		fc.NewBoolFlag("json", "js", "list output in json format (default is csv)")
	*/

	// for AI SI
	fc.NewBoolFlag("ai", "ai", "Application instances")
	fc.NewBoolFlag("si", "si", "Service instances")
	fc.NewBoolFlag("monthly", "monthly", "Monthly usage report, last 7 months")

	err := fc.Parse(args[1:]...)

	if err != nil {
		fmt.Println("\n Receive error reading arguments ... ", err)
		Usage(1)
	}

	today := time.Now()
	var ins Inputs
	ins.isCsv = true
	ins.isJson = false
	ins.fromDate = GetStartOfDay(today)
	ins.toDate = time.Now()

	if fc.IsSet("ai") {
		ins.AI = true
	}
	if fc.IsSet("si") {
		ins.SI = true
	}
	if fc.IsSet("monthly") {
		ins.monthly = true
	}
	/*
		if fc.IsSet("all") {
			nintyDays := time.Hour * -(24 * 90)
			ins.fromDate = today.Add(nintyDays) // today - 90  days
		}
		if fc.IsSet("today") {
			ins.fromDate = GetStartOfDay(today)
		}
		if fc.IsSet("yesterday") {
			oneDay := time.Hour * -24
			ins.fromDate = GetStartOfDay(today.Add(oneDay)) // today - 1 day
			ins.toDate = GetEndOfDay(ins.fromDate)
		}
		if fc.IsSet("yesterday-on") {
			oneDay := time.Hour * -24
			ins.fromDate = GetStartOfDay(today.Add(oneDay)) // today - 1 day
		}
		if fc.IsSet("from") {
			var value = fc.String("from")
			var layout string

			switch len(value) {
			case 8:
				layout = "20060102" // yyyymmdd
			case 14:
				layout = "20060102150405" // yyyymmddhhmmss
			default:
				fmt.Println("Error: Failed to parse `from` date - ", value)
				fmt.Println(err)
				Usage(1)
			}
			t, err := time.Parse(layout, value)
			// fmt.Println("-------> (1) filter date - ", t, filterDate, err)
			if err != nil {
				fmt.Println("Error: Failed to parse `from` date - ", value)
				fmt.Println(err)
				Usage(1)
			} else {
				ins.fromDate = t
			}
		}
		if fc.IsSet("to") {
			var value = fc.String("to")
			const layout = "20060102150405" // yyyymmdd

			switch len(value) {
			case 8:
				value = value + "235959"
			case 14:
			default:
				fmt.Println("Error: Failed to parse `from` date - ", value)
				fmt.Println(err)
				Usage(1)
			}
			t, err := time.Parse(layout, value)
			// fmt.Println("-------> (1) filter date - ", t, filterDate, err)
			if err != nil {
				fmt.Println("Error: Failed to parse given date - ", value)
				fmt.Println(err)
				Usage(1)
			} else {
				// filterDate = fmt.Sprintf("%s", t.Format("2006-01-02"))
				ins.toDate = t
			}
		}

		if fc.IsSet("json") {
			ins.isJson = true
			ins.isCsv = false
		}

		// fmt.Println("-------> (1) ins - ", ins.fromDate, ins.toDate)
	*/
	return ins
}

// prints the results as a csv text to console
func (c Events) EventsInCSVFormat(results OutputResults) {
	fmt.Println("")
	fmt.Printf(results.Comment)

	//  "20161212", "dr", "lab", "app", "pcf-status", "pcf-status",  "app.crash", "crashed", "2 error(s) occurred:\n\n* 2 error(s) occurred:\n\n* Exited with status 255 (out of memory)\n* cancelled\n* 1 error(s) occurred:\n\n* cancelled"
	//  "2016-12-09T21:44:46Z", "demo", "sandbox", "app", "test-nodejs", "admin", "app.update", "stopped", ""

	fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n", "DATE", "ORG", "SPACE", "ACTEE-TYPE", "ACTEE-NAME", "ACTOR", "EVENT TYPE", "DETAILS")
	for _, val := range results.Resources {
		var mdata = sanitize(fmt.Sprintf("%+v", val.Entity.Metadata))
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			val.Entity.Timestamp, val.Entity.Org, val.Entity.Space,
			val.Entity.ActeeType, val.Entity.ActeeName, val.Entity.ActorName, val.Entity.Type, mdata)
	}

}

// prints the results as a json text to console
func (c Events) EventsInJsonFormat(results OutputResults) {
	var out bytes.Buffer
	b, _ := json.Marshal(results)
	err := json.Indent(&out, b, "", "\t")
	if err != nil {
		fmt.Println(" Recevied error formatting json output.")
	} else {
		fmt.Println(out.String())
	}
}
