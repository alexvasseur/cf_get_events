package main

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/olekukonko/tablewriter"
)

func (c Events) GetLabelSpace(cli plugin.CliConnection) {
	var hasspace bool
	hasspace, _ = cli.HasSpace()
	if hasspace {
		var space plugin_models.Space
		space, _ = cli.GetCurrentSpace()
		labels := c.ReadSpaceLabels(space.Guid, cli)
		for k, v := range labels {
			fmt.Printf("%s=%s\n", k, v)
		}
	} else {
		fmt.Printf("no space targetted")
	}
}

func (c Events) WriteLabelSpace(kv string, cli plugin.CliConnection) {
	var hasspace bool
	hasspace, _ = cli.HasSpace()
	if hasspace {
		var space plugin_models.Space
		space, _ = cli.GetCurrentSpace()
		kv := strings.Split(kv, "=")
		if len(kv) == 2 {
			c.WriteSpaceLabel(space.Guid, kv[0], kv[1], cli)
		} else {
			fmt.Printf("misformatted label (must be domain/key=value format)")
		}
	} else {
		fmt.Printf("no space targetted")
	}
}

func (c Events) DeleteLabelSpace(k string, cli plugin.CliConnection) {
	var hasspace bool
	hasspace, _ = cli.HasSpace()
	if hasspace {
		var space plugin_models.Space
		space, _ = cli.GetCurrentSpace()
		c.WriteSpaceLabel(space.Guid, k, "", cli)
	} else {
		fmt.Printf("no space targetted")
	}
}

func (c Events) SearchLabelSpace(label_selector string, cli plugin.CliConnection) {
	spacesv3search := c.SearchSpacesv3(label_selector, cli)
	orgs := c.GetOrgs(cli)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Org", "Space", "Org GUID", "Space GUID"})
	for _, space := range spacesv3search {
		table.Append([]string{orgs[space.Relationships.Organization.Data.OrgGUID].Name, space.Name, space.Relationships.Organization.Data.OrgGUID, space.GUID})
	}
	table.Render()
}
