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

// ModeRead handles read requests for modes
type ModeRead struct {
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

// newModeRead return a new ModeRead handler with input buffer of length
func newModeRead(length int) (string, *ModeRead) {
	r := &ModeRead{}
	r.handlerName = generateHandlerName() + `_r`
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return r.handlerName, r
}

// Register initializes resources provided by the Soma app
func (r *ModeRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (r *ModeRead) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionList,
		msg.ActionShow,
	} {
		hmap.Request(msg.SectionMode, action, r.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (r *ModeRead) Intake() chan msg.Request {
	return r.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (r *ModeRead) PriorityIntake() chan msg.Request {
	return r.Intake()
}

// Run is the event loop for ModeRead
func (r *ModeRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.ModeList: &r.stmtList,
		stmt.ModeShow: &r.stmtShow,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`mode`, err, stmt.Name(statement))
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
func (r *ModeRead) process(q *msg.Request) {
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

// list returns all modes
func (r *ModeRead) list(q *msg.Request, mr *msg.Result) {
	var (
		mode string
		rows *sql.Rows
		err  error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&mode); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Mode = append(mr.Mode, proto.Mode{
			Mode: mode,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns details about a specific mode
func (r *ModeRead) show(q *msg.Request, mr *msg.Result) {
	var (
		mode string
		err  error
	)

	if err = r.stmtShow.QueryRow(
		q.Mode.Mode,
	).Scan(
		&mode,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Mode = append(mr.Mode, proto.Mode{
		Mode: mode,
	})
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *ModeRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
