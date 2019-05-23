/*-
 * Copyright (c) 2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package soma implements the application handlers of the SOMA
// service.
package soma

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/config"
	"github.com/mjolnir42/soma/internal/handler"
)

// Soma application struct
type Soma struct {
	handlerMap   *handler.Map
	logMap       *LogHandleMap
	dbConnection *sql.DB
	conf         *config.Config
	appLog       *logrus.Logger
	reqLog       *logrus.Logger
	errLog       *logrus.Logger
	auditLog     *logrus.Logger
}

// New returns a new SOMA application
func New(
	appHandlerMap *handler.Map,
	logHandleMap *LogHandleMap,
	dbConnection *sql.DB,
	conf *config.Config,
	appLog, reqLog, errLog, auditLog *logrus.Logger,
) *Soma {
	s := Soma{}
	s.handlerMap = appHandlerMap
	s.logMap = logHandleMap
	s.dbConnection = dbConnection
	s.conf = conf
	s.appLog = appLog
	s.reqLog = reqLog
	s.errLog = errLog
	s.auditLog = auditLog
	return &s
}

// exportLogger returns references to the instances loggers
func (s *Soma) exportLogger() []*logrus.Logger {
	return []*logrus.Logger{s.appLog, s.reqLog, s.errLog}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
