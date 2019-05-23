/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 * Copyright (c) 2016, 1&1 Internet SE
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/handler"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// ValidityRead handles read requests for validity definitions
type ValidityRead struct {
	Input       chan msg.Request
	Shutdown    chan struct{}
	handlerName string
	conn        *sql.DB
	stmtList    *sql.Stmt
	stmtShow    *sql.Stmt
	appLog      *logrus.Logger
	reqLog      *logrus.Logger
	errLog      *logrus.Logger
}

// newValidityRead returns a new ValidityRead handler with input buffer
// of length
func newValidityRead(length int) (string, *ValidityRead) {
	r := &ValidityRead{}
	r.handlerName = generateHandlerName() + `_r`
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return r.handlerName, r
}

// Register initializes resources provided by the Soma app
func (r *ValidityRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (r *ValidityRead) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionList,
		msg.ActionShow,
	} {
		hmap.Request(msg.SectionValidity, action, r.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (r *ValidityRead) Intake() chan msg.Request {
	return r.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (r *ValidityRead) PriorityIntake() chan msg.Request {
	return r.Intake()
}

// Run is the event loop for ValidityRead
func (r *ValidityRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.ValidityList: &r.stmtList,
		stmt.ValidityShow: &r.stmtShow,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`validity`, err, stmt.Name(statement))
		}
		defer (*prepStmt).Close()
	}

runloop:
	for {
		select {
		case <-r.Shutdown:
			break runloop
		case req := <-r.Input:
			go func() {
				r.process(&req)
			}()
		}
	}
}

// process is the request dispatcher
func (r *ValidityRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	logRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all validity definitions
func (r *ValidityRead) list(q *msg.Request, mr *msg.Result) {
	var (
		systemProperty, entity string
		rows                   *sql.Rows
		err                    error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&systemProperty,
			&entity,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Validity = append(mr.Validity, proto.Validity{
			SystemProperty: systemProperty,
			Entity:         entity,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns all validity definitions for a specific system property
func (r *ValidityRead) show(q *msg.Request, mr *msg.Result) {
	var (
		systemProperty, entity string
		isInherited            bool
		rows                   *sql.Rows
		err                    error
	)

	if rows, err = r.stmtShow.Query(
		q.Validity.SystemProperty,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	data := make(map[string]map[string]bool)
	for rows.Next() {
		if err = rows.Scan(
			&systemProperty,
			&entity,
			&isInherited,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		if data[entity] == nil {
			data[entity] = make(map[string]bool)
		}
		if isInherited {
			data[entity][`inherited`] = true
		} else {
			data[entity][`direct`] = true
		}
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for entObj := range data {
		mr.Validity = append(mr.Validity, proto.Validity{
			SystemProperty: q.Validity.SystemProperty,
			Entity:         entObj,
			Direct:         data[entObj][`direct`],
			Inherited:      data[entObj][`inherited`],
		})
	}
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *ValidityRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
