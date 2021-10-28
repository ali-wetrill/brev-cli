// Package ls lists workspaces in the current org
package ls

import (
	"strings"

	"github.com/brevdev/brev-cli/pkg/brev_api"
	"github.com/brevdev/brev-cli/pkg/cmdcontext"
	"github.com/brevdev/brev-cli/pkg/terminal"

	"github.com/spf13/cobra"
)

// Helper functions
func getMe() brev_api.User {
	client, _ := brev_api.NewClient()
	user, _ := client.GetMe()
	return *user
}


func getOrgs() ([]brev_api.Organization, error) {
	client, err := brev_api.NewClient()
	if err != nil {
		return nil, err
	}
	orgs, err := client.GetOrgs()
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func getWorkspaces(orgID string) ([]brev_api.Workspace, error) {
	client, err := brev_api.NewClient()
	if err != nil {
		return nil, err
	}
	wss, err := client.GetWorkspaces(orgID)
	if err != nil {
		return nil, err
	}

	return wss, nil
}

func NewCmdLs(t *terminal.Terminal) *cobra.Command {
	cmd := &cobra.Command{
		Annotations: map[string]string{"context": ""},
		Use:         "ls",
		Short:       "List workspaces",
		Long:        "List workspaces within your active org",
		Example:     `brev ls`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := cmdcontext.InvokeParentPersistentPreRun(cmd, args)
			if err != nil {
				return err
			}

			return nil
		},
		Args:      cobra.MinimumNArgs(0),
		ValidArgs: []string{"orgs", "workspaces"},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ls(t, args)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func ls(t *terminal.Terminal, args []string) error {
	if len(args) > 0 && (args[0] == "orgs" || args[0] == "organizations") {
		orgs, err := getOrgs()
		if err != nil {
			return err
		}
		if len(orgs) == 0 {
			t.Vprint(t.Yellow("You don't have any orgs. Create one!"))
			return nil
		}

		t.Vprint(t.Yellow("Your organizations:"))
		for _, v := range orgs {
			t.Vprint("\t" + v.Name + t.Yellow(" id:"+v.ID))
		}
		return nil

	}
	activeorg, err := brev_api.GetActiveOrgContext()
	if err != nil {
		// TODO: if no active org, print everything
		orgs, err := getOrgs()
		if err != nil {
			return err
		}
		if len(orgs) == 0 {
			t.Vprint(t.Yellow("You don't have any orgs or workspaces. Create an org to get started!"))
			return nil
		}
		for _, o := range orgs {
			wss, err := getWorkspaces(o.ID)
			if err != nil {
				return err
			}
			_,_,err = printWorkspaceTable(t, wss, o)
			if err != nil {
				return err
			}
		}
		t.Vprint(t.Yellow("\nYou don't have any active org set. Run 'brev set <orgname>' to set one."))

		return nil;
	}

	wss, err := getWorkspaces(activeorg.ID)
	if err != nil {
		return err
	}
	
	joined, _, err := printWorkspaceTable(t, wss, *activeorg)
	if len(joined) > 0 {
		t.Vprint(t.Green("\nConnect to one with 'brev link <name/id>\n"))
	}

	if err != nil {
		return err
	}

	return nil
}

func truncateString(s string, delimterCount int) string {
	if len(s) <= delimterCount {
		return s
	} else {
		return s[:delimterCount]
	}
}

func printWorkspaceTable(t *terminal.Terminal, workspaces []brev_api.Workspace, activeorg brev_api.Organization)  ([]brev_api.Workspace, []brev_api.Workspace, error) {
	me := getMe()
	
	var unjoinedWorkspaces []brev_api.Workspace
	var joinedWorkspaces []brev_api.Workspace
	
	for _, v := range workspaces {
		if v.CreatedByUserID == me.Id {
			joinedWorkspaces = append(joinedWorkspaces, v)
		} else {
			unjoinedWorkspaces = append(unjoinedWorkspaces, v)
		}
	}
	
	DELIMETER := 40
	if len(joinedWorkspaces) > 0 {
		t.Vprintf("You have %d workspaces in Org "+t.Yellow(activeorg.Name)+"\n", len(joinedWorkspaces))
		t.Vprint("\nNAME"+ strings.Repeat(" ", DELIMETER+1-len("NAME")) +"ID"+ strings.Repeat(" ", len(joinedWorkspaces[0].ID)+5-len("ID")) +"URL")
		for _, v := range joinedWorkspaces {
			t.Vprint(truncateString(v.Name, DELIMETER)+ strings.Repeat(" ", DELIMETER-len(truncateString(v.Name, DELIMETER)))+ " " +v.ID+ strings.Repeat(" ", 5) +v.DNS)
		}
	}

	return joinedWorkspaces, unjoinedWorkspaces, nil
}
