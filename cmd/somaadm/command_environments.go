package main

import (
	"fmt"

	"github.com/1and1/soma/lib/cmpl"
	"github.com/1and1/soma/lib/proto"
	"github.com/codegangsta/cli"
)

func registerEnvironments(app cli.App) *cli.App {
	app.Commands = append(app.Commands,
		[]cli.Command{
			// environments
			{
				Name:  "environments",
				Usage: "SUBCOMMANDS for environments",
				Subcommands: []cli.Command{
					{
						Name:   "add",
						Usage:  "Register a new view",
						Action: runtime(cmdEnvironmentsAdd),
					},
					{
						Name:   "remove",
						Usage:  "Remove an existing unused environment",
						Action: runtime(cmdEnvironmentsRemove),
					},
					{
						Name:         "rename",
						Usage:        "Rename an existing environment",
						Action:       runtime(cmdEnvironmentsRename),
						BashComplete: cmpl.To,
					},
					{
						Name:   "list",
						Usage:  "List all available environments",
						Action: runtime(cmdEnvironmentsList),
					},
					{
						Name:   "show",
						Usage:  "Show information about a specific environment",
						Action: runtime(cmdEnvironmentsShow),
					},
				},
			}, // end environments
		}...,
	)
	return &app
}

func cmdEnvironmentsAdd(c *cli.Context) error {
	utl.ValidateCliArgumentCount(c, 1)

	req := proto.NewEnvironmentRequest()
	req.Environment.Name = c.Args().First()

	resp := utl.PostRequestWithBody(Client, req, "/environments/")
	fmt.Println(resp)
	return nil
}

func cmdEnvironmentsRemove(c *cli.Context) error {
	utl.ValidateCliArgumentCount(c, 1)

	path := fmt.Sprintf("/environments/%s", c.Args().First())

	resp := utl.DeleteRequest(Client, path)
	fmt.Println(resp)
	return nil
}

func cmdEnvironmentsRename(c *cli.Context) error {
	utl.ValidateCliArgumentCount(c, 3)
	key := []string{`to`}

	opts := utl.ParseVariadicArguments(key, key, key, c.Args().Tail())

	req := proto.NewEnvironmentRequest()
	req.Environment.Name = opts[`to`][0]

	path := fmt.Sprintf("/environments/%s", c.Args().First())

	resp := utl.PutRequestWithBody(Client, req, path)
	fmt.Println(resp)
	return nil
}

func cmdEnvironmentsList(c *cli.Context) error {
	utl.ValidateCliArgumentCount(c, 0)
	resp := utl.GetRequest(Client, "/environments/")
	fmt.Println(resp)
	return nil
}

func cmdEnvironmentsShow(c *cli.Context) error {
	utl.ValidateCliArgumentCount(c, 1)

	path := fmt.Sprintf("/environments/%s", c.Args().First())

	resp := utl.GetRequest(Client, path)
	fmt.Println(resp)
	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix