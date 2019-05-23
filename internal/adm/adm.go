package adm

import (
	"github.com/mjolnir42/soma/internal/db"

	"gopkg.in/resty.v1"
)

var (
	client        *resty.Client
	cache         *db.DB
	async         bool
	jobSave       bool
	postProcessor string
)

func ConfigureClient(c *resty.Client) {
	client = c
}

func ConfigureCache(c *db.DB) {
	cache = c
}

func ActivateAsyncWait(b bool) {
	async = b
}

func AutomaticJobSave(b bool) {
	jobSave = b
}

func ConfigureJSONPostProcessor(p string) {
	postProcessor = p
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
