package main

import (
	"github.com/codegangsta/cli"
)

func dbClose(c *cli.Context) error {
	printOnly := c.GlobalBool("no-execute")
	if !printOnly {
		defer db.Close()
	}

	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
