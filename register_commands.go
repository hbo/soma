package main

import (
	"github.com/codegangsta/cli"
)

func registerCommands(app cli.App) *cli.App {

	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Initialize local client files",
			Action: cmdClientInit,
		}, // end init
		{
			Name:   "views",
			Usage:  "SUBCOMMANDS for views",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Register a new view",
					Action: cmdViewsAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove an existing view",
					Action: cmdViewsRemove,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing view",
					Action: cmdViewsRename,
				},
				{
					Name:   "list",
					Usage:  "List all registered views",
					Action: cmdViewsList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a specific view",
					Action: cmdViewsShow,
				},
			},
		}, // end views
		{
			Name:   "environments",
			Usage:  "SUBCOMMANDS for environments",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Register a new view",
					Action: cmdEnvironmentsAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove an existing unused environment",
					Action: cmdEnvironmentsRemove,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing environment",
					Action: cmdEnvironmentsRename,
				},
				{
					Name:   "list",
					Usage:  "List all available environments",
					Action: cmdEnvironmentsList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a specific environment",
					Action: cmdEnvironmentsShow,
				},
			},
		}, // end environments
		{
			Name:   "types",
			Usage:  "SUBCOMMANDS for object types",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Add a new object type",
					Action: cmdObjectTypesAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove an existing object type",
					Action: cmdObjectTypesRemove,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing object type",
					Action: cmdObjectTypesRename,
				},
				{
					Name:   "list",
					Usage:  "List all object types",
					Action: cmdObjectTypesList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a specific object type",
					Action: cmdObjectTypesShow,
				},
			},
		}, // end types
		{
			Name:   "states",
			Usage:  "SUBCOMMANDS for states",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Add a new object state",
					Action: cmdObjectStatesAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove an existing object state",
					Action: cmdObjectStatesRemove,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing object state",
					Action: cmdObjectStatesRename,
				},
				{
					Name:   "list",
					Usage:  "List all object states",
					Action: cmdObjectStatesList,
				},
				{
					Name:   "show",
					Usage:  "Show information about an object states",
					Action: cmdObjectStatesShow,
				},
			},
		}, // end states
		{
			Name:   "datacenters",
			Usage:  "SUBCOMMANDS for datacenters",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Register a new datacenter",
					Action: cmdDatacentersAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove an existing datacenter",
					Action: cmdDatacentersRemove,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing datacenter",
					Action: cmdDatacentersRename,
				},
				{
					Name:   "list",
					Usage:  "List all datacenters",
					Action: cmdDatacentersList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a specific datacenter",
					Action: cmdDatacentersShow,
				},
				{
					Name:   "groupadd",
					Usage:  "Add a datacenter to a datacenter group",
					Action: cmdDatacentersAddToGroup,
				},
				{
					Name:   "groupdel",
					Usage:  "Remove a datacenter from a datacenter group",
					Action: cmdDatacentersRemoveFromGroup,
				},
				{
					Name:   "grouplist",
					Usage:  "List all datacenter groups",
					Action: cmdDatacentersListGroups,
				},
				{
					Name:   "groupshow",
					Usage:  "Show information about a datacenter group",
					Action: cmdDatacentersShowGroup,
				},
			},
		}, // end datacenters
		{
			Name:   "servers",
			Usage:  "SUBCOMMANDS for servers",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:        "create",
					Usage:       "Create a new physical server",
					Description: help.CmdServerCreate,
					Action:      cmdServerCreate,
				},
				{
					Name:   "delete",
					Usage:  "Mark an existing physical server as deleted",
					Action: cmdServerMarkAsDeleted,
				},
				{
					Name:   "purge",
					Usage:  "Remove all unreferenced servers marked as deleted",
					Action: cmdServerPurgeDeleted,
				},
				{
					Name:   "update",
					Usage:  "Full update of server attributes (replace, not merge)",
					Action: cmdServerUpdate,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing server",
					Action: cmdServerRename,
				},
				{
					Name:   "online",
					Usage:  "Set an existing server to online",
					Action: cmdServerOnline,
				},
				{
					Name:   "offline",
					Usage:  "Set an existing server to offline",
					Action: cmdServerOffline,
				},
				{
					Name:   "move",
					Usage:  "Change a server's registered location",
					Action: cmdServerMove,
				},
				{
					Name:   "list",
					Usage:  "List all servers, see full description for possible filters",
					Action: cmdServerList,
				},
				{
					Name:   "show",
					Usage:  "Show details about a specific server",
					Action: cmdServerShow,
				},
				{
					Name:   "sync",
					Usage:  "Request a data sync for a server",
					Action: cmdServerSyncRequest,
				},
			},
		}, // end servers
		{
			Name:   "permissions",
			Usage:  "SUBCOMMANDS for permissions",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:  "type",
					Usage: "SUBCOMMANDS for permission types",
					Subcommands: []cli.Command{
						{
							Name:   "add",
							Usage:  "Register a new permission type",
							Action: cmdPermissionTypeAdd,
						},
						{
							Name:   "remove",
							Usage:  "Remove an existing permission type",
							Action: cmdPermissionTypeDel,
						},
						{
							Name:   "rename",
							Usage:  "Rename an existing permission type",
							Action: cmdPermissionTypeRename,
						},
						{
							Name:   "list",
							Usage:  "List all permission types",
							Action: cmdPermissionTypeList,
						},
						{
							Name:   "show",
							Usage:  "Show details for a permission type",
							Action: cmdPermissionTypeShow,
						},
					}, // end permissions type
				},
				{
					Name:   "add",
					Usage:  "Register a new permission",
					Action: cmdPermissionAdd,
				},
				{
					Name:   "remove",
					Usage:  "Remove a permission",
					Action: cmdPermissionDel,
				},
				{
					Name:   "list",
					Usage:  "List all permissions",
					Action: cmdPermissionList,
				},
				{
					Name:  "show",
					Usage: "SUBCOMMANDS for permission show",
					Subcommands: []cli.Command{
						{
							Name:   "user",
							Usage:  "Show permissions of a user",
							Action: cmdPermissionShowUser,
						},
						{
							Name:   "team",
							Usage:  "Show permissions of a team",
							Action: cmdPermissionShowTeam,
						},
						{
							Name:   "tool",
							Usage:  "Show permissions of a tool account",
							Action: cmdPermissionShowTool,
						},
						{
							Name:   "permission",
							Usage:  "Show details about a permission",
							Action: cmdPermissionShowPermission,
						},
					},
				}, // end permissions show
				{
					Name:   "audit",
					Usage:  "Show all limited permissions associated with a repository",
					Action: cmdPermissionAudit,
				},
				{
					Name:  "grant",
					Usage: "SUBCOMMANDS for permission grant",
					Subcommands: []cli.Command{
						{
							Name:   "enable",
							Usage:  "Enable a useraccount to receive GRANT permissions",
							Action: cmdPermissionGrantEnable,
						},
						{
							Name:   "global",
							Usage:  "Grant a global permission",
							Action: cmdPermissionGrantGlobal,
						},
						{
							Name:   "limited",
							Usage:  "Grant a limited permission",
							Action: cmdPermissionGrantLimited,
						},
						{
							Name:   "system",
							Usage:  "Grant a system permission",
							Action: cmdPermissionGrantSystem,
						},
					},
				}, // end permissions grant
			},
		}, // end permissions
		{
			Name:   "teams",
			Usage:  "SUBCOMMANDS for teams",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Register a new team",
					Action: cmdTeamAdd,
				},
				{
					Name:   "remove",
					Usage:  "Delete an existing team",
					Action: cmdTeamDel,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing team",
					Action: cmdTeamRename,
				},
				{
					Name:   "migrate",
					Usage:  "Migrate users between teams",
					Action: cmdTeamMigrate,
				},
				{
					Name:   "list",
					Usage:  "List all teams",
					Action: cmdTeamList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a team",
					Action: cmdTeamShow,
				},
			},
		}, // end teams
		{
			Name:   "oncall",
			Usage:  "SUBCOMMANDS for oncall duty teams",
			Before: runtimePreCmd,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Register a new oncall duty team",
					Action: cmdOnCallAdd,
				},
				{
					Name:   "remove",
					Usage:  "Delete an existing oncall duty team",
					Action: cmdOnCallDel,
				},
				{
					Name:   "rename",
					Usage:  "Rename an existing oncall duty team",
					Action: cmdOnCallRename,
				},
				{
					Name:   "update",
					Usage:  "Update phone number of an existing oncall duty team",
					Action: cmdOnCallUpdate,
				},
				{
					Name:   "list",
					Usage:  "List all registered oncall duty teams",
					Action: cmdOnCallList,
				},
				{
					Name:   "show",
					Usage:  "Show information about a specific oncall duty team",
					Action: cmdOnCallShow,
				},
				{
					Name:  "member",
					Usage: "SUBCOMMANDS to manipulate oncall duty members",
					Subcommands: []cli.Command{
						{
							Name:   "add",
							Usage:  "Add a user to an oncall duty team",
							Action: cmdOnCallMemberAdd,
						},
						{
							Name:   "remove",
							Usage:  "Remove a member from an oncall duty team",
							Action: cmdOnCallMemberDel,
						},
						{
							Name:   "list",
							Usage:  "List the users of an oncall duty team",
							Action: cmdOnCallMemberList,
						},
					},
				},
			},
		}, // end oncall
	}
	return &app
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
