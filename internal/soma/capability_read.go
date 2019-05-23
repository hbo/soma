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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/handler"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// CapabilityRead handles read requests for capabilities
type CapabilityRead struct {
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

// newCapabilityRead return a new CapabilityRead handler with input buffer of length
func newCapabilityRead(length int) (string, *CapabilityRead) {
	r := &CapabilityRead{}
	r.handlerName = generateHandlerName() + `_r`
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return r.handlerName, r
}

// Register initializes resources provided by the Soma app
func (r *CapabilityRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (r *CapabilityRead) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionList,
		msg.ActionShow,
		msg.ActionSearch,
	} {
		hmap.Request(msg.SectionCapability, action, r.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (r *CapabilityRead) Intake() chan msg.Request {
	return r.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (r *CapabilityRead) PriorityIntake() chan msg.Request {
	return r.Intake()
}

// Run is the event loop for CapabilityRead
func (r *CapabilityRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.ListAllCapabilities: &r.stmtList,
		stmt.ShowCapability:      &r.stmtShow,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`capability`, err, stmt.Name(statement))
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
func (r *CapabilityRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	logRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionSearch:
		// XXX BUG r.search(q, &result)
		r.list(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all capabilities
func (r *CapabilityRead) list(q *msg.Request, mr *msg.Result) {
	var (
		id, monitoring, metric, view, monName string
		rows                                  *sql.Rows
		err                                   error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&id,
			&monitoring,
			&metric,
			&view,
			&monName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Capability = append(mr.Capability, proto.Capability{
			ID:           id,
			MonitoringID: monitoring,
			Metric:       metric,
			View:         view,
			Name: fmt.Sprintf("%s.%s.%s", monName, view,
				metric),
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details of a specific capability
func (r *CapabilityRead) show(q *msg.Request, mr *msg.Result) {
	var (
		id, monitoring, metric, view, monName string
		thresholds                            int
		err                                   error
	)

	if err = r.stmtShow.QueryRow(
		q.Capability.ID,
	).Scan(
		&id,
		&monitoring,
		&metric,
		&view,
		&thresholds,
		&monName,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	mr.Capability = append(mr.Capability, proto.Capability{
		ID:           id,
		MonitoringID: monitoring,
		Metric:       metric,
		View:         view,
		Thresholds:   uint64(thresholds),
		Name:         fmt.Sprintf("%s.%s.%s", monName, view, metric),
	})
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *CapabilityRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
