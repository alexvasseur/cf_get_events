package main

import (
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

type OrgSummaryFromQuota struct {
	Mem         int
	MemOrgQuota int
}

type OrgSummary struct {
	Name                string                  `json:"name"`
	Resources           []OrgSummarySpaceEntity `json:"spaces"`
	Memory              int
	MemoryLimitOrgQuota int
	MemoryUsage         int
}

type OrgSummarySpaceEntity struct {
	Name         string `json:"name"`
	ServiceCount int    `json:"service_count"`
	Memory       int    `json:"mem_dev_total"`
}

type QuotaDefinition struct {
	Entity QuotaEntity `json:"entity"`
}

type QuotaEntity struct {
	MemoryLimit int `json:"memory_limit"`
}

func (c Events) GetOrgsSummary(cli plugin.CliConnection) map[string]OrgSummary {
	var data = make(map[string]OrgSummary)
	orgs := c.GetOrgData(cli)

	var quotaCache = make(map[string]QuotaDefinition)

	for _, val := range orgs.Resources {
		// lookup usage in /v2/organizations/<org guid>/summary
		var orgSummary = c.GetOrgSummaryData(cli, val.Metadata.GUID)

		// compute Mem use total from all org' spaces
		for _, space := range orgSummary.Resources {
			orgSummary.Memory += space.Memory
		}

		// lookup quota def
		orgQuota, exist := quotaCache[val.Entity.QuotaGuid]
		if !exist {
			orgQuota = c.GetQuotaData(cli, val.Entity.QuotaGuid)
			quotaCache[val.Entity.QuotaGuid] = orgQuota
		}
		orgSummary.MemoryLimitOrgQuota = orgQuota.Entity.MemoryLimit
		// special case for quota <=0 - see https://github.com/avasseur-pivotal/cf_get_events/issues/7
		if orgSummary.MemoryLimitOrgQuota > 0 {
			orgSummary.MemoryUsage = (int)(orgSummary.Memory * 100 / orgSummary.MemoryLimitOrgQuota)
		} else {
			orgSummary.MemoryUsage = 0
		}
		data[val.Metadata.GUID] = orgSummary
	}

	return data
}

func (c Events) GetOrgSummaryData(cli plugin.CliConnection, orgGuid string) OrgSummary {
	var res OrgSummary
	cmd := []string{"curl", "/v2/organizations/" + orgGuid + "/summary"}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &res)
	return res
}

func (c Events) GetQuotaData(cli plugin.CliConnection, quotaGuid string) QuotaDefinition {
	var res QuotaDefinition
	cmd := []string{"curl", "/v2/quota_definitions/" + quotaGuid}
	output, _ := cli.CliCommandWithoutTerminalOutput(cmd...)
	json.Unmarshal([]byte(strings.Join(output, "")), &res)
	return res
}
