/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016-2018, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/mjolnir42/soma/internal/adm"
)

func registerInstances(app cli.App) *cli.App {
	app.Commands = append(app.Commands,
		[]cli.Command{
			{
				Name:  `instances`,
				Usage: `SUBCOMMANDS for check instances`,
				Subcommands: []cli.Command{
					{
						Name:   `cascade-delete`,
						Usage:  `Delete check configuration that created the instance`,
						Action: runtime(cmdInstanceCascade),
					},
				},
			},
		}...,
	)
	return &app
}

func cmdInstanceCascade(c *cli.Context) error {
	var (
		err             error
		checkID, repoID string
	)
	if err = adm.VerifySingleArgument(c); err != nil {
		return err
	}
	if !adm.IsUUID(c.Args().First()) {
		return fmt.Errorf("Argument is not a UUID: %s",
			c.Args().First())
	}
	if checkID, repoID, err = adm.LookupCheckConfigID(
		``, ``, c.Args().First()); err != nil {
		return err
	}

	path := fmt.Sprintf("/checks/%s/%s", repoID, checkID)
	return adm.Perform(`delete`, path, `command`, nil, c)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
