package cmpl

import "github.com/codegangsta/cli"

func OncallAdd(c *cli.Context) {
	Generic(c, []string{`phone`})
}

func OncallUpdate(c *cli.Context) {
	Generic(c, []string{`phone`, `name`})
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
