/*-
 * Copyright (c) 2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package rest implements the REST routes to access SOMA.
package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"crypto/tls"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/config"
	"github.com/mjolnir42/soma/internal/handler"
	"github.com/mjolnir42/soma/internal/msg"
	metrics "github.com/rcrowley/go-metrics"
)

// ShutdownInProgress indicates a pending service shutdown
var ShutdownInProgress bool

// Metrics is the map of runtime metric registries
var Metrics = make(map[string]metrics.Registry)

// Rest holds the required state for the REST interface
type Rest struct {
	isAuthorized func(*msg.Request) bool
	handlerMap   *handler.Map
	conf         *config.Config
	reqLog       *logrus.Logger
	errLog       *logrus.Logger
	restricted   bool
}

// New returns a new REST interface
func New(
	authorizationFunction func(*msg.Request) bool,
	appHandlerMap *handler.Map,
	conf *config.Config,
	reqLog, errLog *logrus.Logger,
) *Rest {
	x := Rest{}
	x.isAuthorized = authorizationFunction
	x.restricted = false
	x.handlerMap = appHandlerMap
	x.reqLog = reqLog
	x.errLog = errLog
	x.conf = conf
	return &x
}

// Run is the event server for Rest
func (x *Rest) Run() {
	router := x.setupRouter()

	// TODO switch to new abortable interface
	if x.conf.Daemon.TLS {
		server := &http.Server{Addr: x.conf.Daemon.URL.Host, Handler: router}
		server.TLSConfig = &tls.Config{MaxVersion: tls.VersionTLS13, MinVersion: tls.VersionTLS10}
		err := server.ListenAndServeTLS(x.conf.Daemon.Cert, x.conf.Daemon.Key)
		x.errLog.Fatal(err)
	} else {
		x.errLog.Fatal(http.ListenAndServe(x.conf.Daemon.URL.Host, router))
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
