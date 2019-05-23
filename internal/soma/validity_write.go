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
)

// ValidityWrite handles write requests for validity definitions
type ValidityWrite struct {
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

// newValidityWrite returns a new ValidityWrite handler with input
// buffer of length
func newValidityWrite(length int) (string, *ValidityWrite) {
	w := &ValidityWrite{}
	w.handlerName = generateHandlerName() + `_w`
	w.Input = make(chan msg.Request, length)
	w.Shutdown = make(chan struct{})
	return w.handlerName, w
}

// Register initializes resources provided by the Soma app
func (w *ValidityWrite) Register(c *sql.DB, l ...*logrus.Logger) {
	w.conn = c
	w.appLog = l[0]
	w.reqLog = l[1]
	w.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (w *ValidityWrite) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionAdd,
		msg.ActionRemove,
	} {
		hmap.Request(msg.SectionValidity, action, w.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (w *ValidityWrite) Intake() chan msg.Request {
	return w.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (w *ValidityWrite) PriorityIntake() chan msg.Request {
	return w.Intake()
}

// Run is the event loop for ValidityWrite
func (w *ValidityWrite) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.ValidityAdd: &w.stmtAdd,
		stmt.ValidityDel: &w.stmtRemove,
	} {
		if *prepStmt, err = w.conn.Prepare(statement); err != nil {
			w.errLog.Fatal(`validity`, err, stmt.Name(statement))
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
func (w *ValidityWrite) process(q *msg.Request) {
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

// add inserts a new validity definition
func (w *ValidityWrite) add(q *msg.Request, mr *msg.Result) {
	var (
		err error
		res sql.Result
		tx  *sql.Tx
	)

	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	txStmtAdd := tx.Stmt(w.stmtAdd)

	// every record inside the table is a validity that either
	// represents a direct validity (inherited == false) or an
	// inherited validity (inherited == true).
	// Invalidity is the absence of data causing a foreign key
	// constraint violation.

	// insert direct validity
	if q.Validity.Direct {
		if res, err = txStmtAdd.Exec(
			q.Validity.SystemProperty,
			q.Validity.Entity,
			false,
		); err != nil {
			tx.Rollback()
			mr.ServerError(err, q.Section)
			return
		}
		if !mr.RowCnt(res.RowsAffected()) {
			tx.Rollback()
			return
		}
	}
	// insert inherited validity
	if q.Validity.Inherited {
		if res, err = txStmtAdd.Exec(
			q.Validity.SystemProperty,
			q.Validity.Entity,
			true,
		); err != nil {
			tx.Rollback()
			mr.ServerError(err, q.Section)
			return
		}
		if !mr.RowCnt(res.RowsAffected()) {
			tx.Rollback()
			return
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		mr.ServerError(err, q.Section)
		return
	}
	mr.Validity = append(mr.Validity, q.Validity)
	mr.OK()
}

// remove deletes a validity definition
func (w *ValidityWrite) remove(q *msg.Request, mr *msg.Result) {
	var (
		err error
		res sql.Result
	)

	if res, err = w.stmtRemove.Exec(
		q.Validity.SystemProperty,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if mr.RowCntMany(res.RowsAffected()) {
		mr.Validity = append(mr.Validity, q.Validity)
	}
}

// ShutdownNow signals the handler to shut down
func (w *ValidityWrite) ShutdownNow() {
	close(w.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
