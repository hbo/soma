package main

import "github.com/codegangsta/cli"

func registerCommands(app cli.App) *cli.App {
	app.Commands = []cli.Command{
		{
			Name:    "initialize",
			Aliases: []string{"init"},
			Usage:   "initialize an empty database",
			Before:  configSetup,
			After:   dbClose,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose, v",
					Usage: "print full query on execute",
				},
			},
			Action: func(c *cli.Context) {
				done := make(chan bool, 1)
				printOnly := c.GlobalBool("no-execute")
				verbose := c.Bool("verbose")
				commandInitialize(done, printOnly, verbose, app.Version)
				<-done
			},
		},
		{
			Name:    "wipe",
			Aliases: []string{"rm"},
			Usage:   "wipe database contents",
			Description: `Completely cleans the database. Very destructive. Removes all:
     * row data
     * indices
     * tables
     * schemas`,
			Before: configSetup,
			After:  dbClose,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Do not prompt for confirmation",
				},
			},
			Action: func(c *cli.Context) {
				done := make(chan bool, 1)
				forced := c.Bool("force")
				printOnly := c.GlobalBool("no-execute")
				commandWipe(done, forced, printOnly)
				<-done
			},
		},
		{
			Name:    "upgrade",
			Aliases: []string{"update"},
			Usage:   "update database schema version",
			Before:  configSetup,
			After:   dbClose,
			Action: func(c *cli.Context) {
				done := make(chan bool, 1)
				printOnly := c.GlobalBool("no-execute")
				commandUpgradeSchema(done, 0, app.Version, printOnly)
				<-done
			},
		},
		{
			Name:  "cleanup",
			Usage: "Clean database from various objects",
			Subcommands: []cli.Command{
				{
					Name:   "checks",
					Usage:  "Clean out all checks from the database",
					Before: configSetup,
					After:  dbClose,
					Action: func(c *cli.Context) {
						done := make(chan bool, 1)
						printOnly := c.GlobalBool("no-execute")
						commandCleanupChecks(done, printOnly)
						<-done
					},
				},
			},
		},
	}
	return &app
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
