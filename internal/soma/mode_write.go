/*-
 * Copyright (c) 2016-2018, Jörg Pernfuß
 * Copyright (c) 2016, 1&1 Internet SE
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma // import "github.com/mjolnir42/soma/internal/soma"

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/handler"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
)

// ModeWrite handles write requests for modes
type ModeWrite struct {
	Input       chan msg.Request
	Shutdown    chan struct{}
	handlerName string
	conn        *sql.DB
	stmtAdd     *sql.Stmt
	stmtRemove  *sql.Stmt
	appLog      *logrus.Logger
	reqLog      *logrus.Logger
	errLog      *logrus.Logger
}

// newModeWrite return a new ModeWrite handler with input buffer of
// length
func newModeWrite(length int) (string, *ModeWrite) {
	w := &ModeWrite{}
	w.handlerName = generateHandlerName() + `_w`
	w.Input = make(chan msg.Request, length)
	w.Shutdown = make(chan struct{})
	return w.handlerName, w
}

// Register initializes resources provided by the Soma app
func (w *ModeWrite) Register(c *sql.DB, l ...*logrus.Logger) {
	w.conn = c
	w.appLog = l[0]
	w.reqLog = l[1]
	w.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (w *ModeWrite) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionAdd,
		msg.ActionRemove,
	} {
		hmap.Request(msg.SectionMode, action, w.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (w *ModeWrite) Intake() chan msg.Request {
	return w.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (w *ModeWrite) PriorityIntake() chan msg.Request {
	return w.Intake()
}

// Run is the event loop for ModeWrite
func (w *ModeWrite) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.ModeAdd: &w.stmtAdd,
		stmt.ModeDel: &w.stmtRemove,
	} {
		if *prepStmt, err = w.conn.Prepare(statement); err != nil {
			w.errLog.Fatal(`mode`, err, stmt.Name(statement))
		}
		defer (*prepStmt).Close()
	}

runloop:
	for {
		select {
		case <-w.Shutdown:
			break runloop
		case req := <-w.Input:
			w.process(&req)
		}
	}
}

// process is the request dispatcher
func (w *ModeWrite) process(q *msg.Request) {
	result := msg.FromRequest(q)
	logRequest(w.reqLog, q)

	switch q.Action {
	case msg.ActionAdd:
		w.add(q, &result)
	case msg.ActionRemove:
		w.remove(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// add inserts a new mode
func (w *ModeWrite) add(q *msg.Request, mr *msg.Result) {
	var (
		res sql.Result
		err error
	)

	if res, err = w.stmtAdd.Exec(
		q.Mode.Mode,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if mr.RowCnt(res.RowsAffected()) {
		mr.Mode = append(mr.Mode, q.Mode)
	}
}

// remove deletes a mode
func (w *ModeWrite) remove(q *msg.Request, mr *msg.Result) {
	var (
		res sql.Result
		err error
	)

	if res, err = w.stmtRemove.Exec(
		q.Mode.Mode,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if mr.RowCnt(res.RowsAffected()) {
		mr.Mode = append(mr.Mode, q.Mode)
	}
}

// ShutdownNow signals the handler to shut down
func (w *ModeWrite) ShutdownNow() {
	close(w.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
