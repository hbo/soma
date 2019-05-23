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

// PropertyRead handles read requests for properties
type PropertyRead struct {
	Input            chan msg.Request
	Shutdown         chan struct{}
	handlerName      string
	conn             *sql.DB
	stmtListCustom   *sql.Stmt
	stmtListNative   *sql.Stmt
	stmtListService  *sql.Stmt
	stmtListSystem   *sql.Stmt
	stmtListTemplate *sql.Stmt
	stmtShowCustom   *sql.Stmt
	stmtShowNative   *sql.Stmt
	stmtShowService  *sql.Stmt
	stmtShowSystem   *sql.Stmt
	stmtShowTemplate *sql.Stmt
	appLog           *logrus.Logger
	reqLog           *logrus.Logger
	errLog           *logrus.Logger
}

// newPropertyRead return a new PropertyRead handler with input
// buffer of length
func newPropertyRead(length int) (string, *PropertyRead) {
	r := &PropertyRead{}
	r.handlerName = generateHandlerName() + `_r`
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return r.handlerName, r
}

// Register initializes resources provided by the Soma app
func (r *PropertyRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (r *PropertyRead) RegisterRequests(hmap *handler.Map) {
	for _, section := range []string{
		msg.SectionPropertyCustom,
		msg.SectionPropertyNative,
		msg.SectionPropertyService,
		msg.SectionPropertySystem,
		msg.SectionPropertyTemplate,
	} {
		for _, action := range []string{
			msg.ActionList,
			msg.ActionShow,
			msg.ActionSearch,
		} {
			hmap.Request(section, action, r.handlerName)
		}
	}
}

// Run is the event loop for PropertyRead
func (r *PropertyRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.PropertyCustomList:   &r.stmtListCustom,
		stmt.PropertyCustomShow:   &r.stmtShowCustom,
		stmt.PropertyNativeList:   &r.stmtListNative,
		stmt.PropertyNativeShow:   &r.stmtShowNative,
		stmt.PropertyServiceList:  &r.stmtListService,
		stmt.PropertyServiceShow:  &r.stmtShowService,
		stmt.PropertySystemList:   &r.stmtListSystem,
		stmt.PropertySystemShow:   &r.stmtShowSystem,
		stmt.PropertyTemplateList: &r.stmtListTemplate,
		stmt.PropertyTemplateShow: &r.stmtShowTemplate,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`property`, err, stmt.Name(statement))
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

// Intake exposes the Input channel as part of the handler interface
func (r *PropertyRead) Intake() chan msg.Request {
	return r.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (r *PropertyRead) PriorityIntake() chan msg.Request {
	return r.Intake()
}

// process is the request dispatcher
func (r *PropertyRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	logRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionSearch:
		r.search(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all properties
func (r *PropertyRead) list(q *msg.Request, mr *msg.Result) {
	switch q.Property.Type {
	case msg.PropertyCustom:
		r.listCustom(q, mr)
	case msg.PropertyNative:
		r.listNative(q, mr)
	case msg.PropertyService:
		r.listService(q, mr)
	case msg.PropertySystem:
		r.listSystem(q, mr)
	case msg.PropertyTemplate:
		r.listTemplate(q, mr)
	default:
		mr.NotImplemented(fmt.Errorf("Unknown property type: %s",
			q.Property.Type))
	}
}

// listCustom returns all custom properties for a repository
func (r *PropertyRead) listCustom(q *msg.Request, mr *msg.Result) {
	var (
		property, repository, id string
		rows                     *sql.Rows
		err                      error
	)

	if rows, err = r.stmtListCustom.Query(
		q.Property.Custom.RepositoryID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&id, &repository, &property); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Custom: &proto.PropertyCustom{
				ID:           id,
				RepositoryID: repository,
				Name:         property,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// listNative returns all native properties
func (r *PropertyRead) listNative(q *msg.Request, mr *msg.Result) {
	var (
		property string
		rows     *sql.Rows
		err      error
	)

	if rows, err = r.stmtListNative.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&property); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Native: &proto.PropertyNative{
				Name: property,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// listService returns all service properties for a team
func (r *PropertyRead) listService(q *msg.Request, mr *msg.Result) {
	var (
		id, name, teamID string
		rows             *sql.Rows
		err              error
	)

	if rows, err = r.stmtListService.Query(
		q.Property.Service.TeamID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&id, &name, &teamID); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Service: &proto.PropertyService{
				ID:     id,
				Name:   name,
				TeamID: teamID,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// listSystem returns all system properties
func (r *PropertyRead) listSystem(q *msg.Request, mr *msg.Result) {
	var (
		property string
		rows     *sql.Rows
		err      error
	)

	if rows, err = r.stmtListSystem.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&property); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			System: &proto.PropertySystem{
				Name: property,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// listTemplate returns all service templates
func (r *PropertyRead) listTemplate(q *msg.Request, mr *msg.Result) {
	var (
		id, name string
		rows     *sql.Rows
		err      error
	)

	if rows, err = r.stmtListTemplate.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Service: &proto.PropertyService{
				ID:   id,
				Name: name,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns details about a specific property
func (r *PropertyRead) show(q *msg.Request, mr *msg.Result) {
	switch q.Property.Type {
	case msg.PropertyCustom:
		r.showCustom(q, mr)
	case msg.PropertyNative:
		r.showNative(q, mr)
	case msg.PropertyService:
		r.showService(q, mr)
	case msg.PropertySystem:
		r.showSystem(q, mr)
	case msg.PropertyTemplate:
		r.showTemplate(q, mr)
	default:
		mr.NotImplemented(fmt.Errorf("Unknown property type: %s",
			q.Property.Type))
	}
}

// showCustom returns the details for a specific custom property
func (r *PropertyRead) showCustom(q *msg.Request, mr *msg.Result) {
	var (
		property, repository, id string
		err                      error
	)

	if err = r.stmtShowCustom.QueryRow(
		q.Property.Custom.ID,
		q.Property.Custom.RepositoryID,
	).Scan(
		&id,
		&repository,
		&property,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Property = append(mr.Property, proto.Property{
		Type: q.Property.Type,
		Custom: &proto.PropertyCustom{
			ID:           id,
			RepositoryID: repository,
			Name:         property,
		},
	})
	mr.OK()
}

// showNative returns the details for a specific native property
func (r *PropertyRead) showNative(q *msg.Request, mr *msg.Result) {
	var (
		property string
		err      error
	)

	if err = r.stmtShowNative.QueryRow(
		q.Property.Native.Name,
	).Scan(
		&property,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Property = append(mr.Property, proto.Property{
		Type: q.Property.Type,
		Native: &proto.PropertyNative{
			Name: property,
		},
	})
	mr.OK()
}

// showService returns the details for a specific service
func (r *PropertyRead) showService(q *msg.Request, mr *msg.Result) {
	var (
		id, name, teamID, attribute, value string
		rows                               *sql.Rows
		err                                error
		service                            proto.PropertyService
	)

	if rows, err = r.stmtShowService.Query(
		q.Property.Service.ID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	service.Attributes = make([]proto.ServiceAttribute, 0)

	for rows.Next() {
		if err = rows.Scan(
			&id,
			&name,
			&teamID,
			&attribute,
			&value,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}

		service.ID = id
		service.Name = name
		service.TeamID = teamID
		service.Attributes = append(service.Attributes,
			proto.ServiceAttribute{
				Name:  attribute,
				Value: value,
			})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	mr.Property = append(mr.Property, proto.Property{
		Type:    q.Property.Type,
		Service: &service,
	})
	mr.OK()
}

// showSystem returns the details about a specific system property
func (r *PropertyRead) showSystem(q *msg.Request, mr *msg.Result) {
	var (
		property string
		err      error
	)

	if err = r.stmtShowSystem.QueryRow(
		q.Property.System.Name,
	).Scan(
		&property,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Property = append(mr.Property, proto.Property{
		Type: q.Property.Type,
		System: &proto.PropertySystem{
			Name: property,
		},
	})
	mr.OK()
}

// showTemplate returns the details about a specific service
// template
func (r *PropertyRead) showTemplate(q *msg.Request, mr *msg.Result) {
	var (
		id, name, attribute, value string
		rows                       *sql.Rows
		err                        error
		template                   proto.PropertyService
	)

	if rows, err = r.stmtShowTemplate.Query(
		q.Property.Service.ID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	template.Attributes = make([]proto.ServiceAttribute, 0)

	for rows.Next() {
		if err = rows.Scan(
			&id,
			&name,
			&attribute,
			&value,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}

		template.ID = id
		template.Name = name
		template.Attributes = append(template.Attributes,
			proto.ServiceAttribute{
				Name:  attribute,
				Value: value,
			})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	mr.Property = append(mr.Property, proto.Property{
		Type:    q.Property.Type,
		Service: &template,
	})
	mr.OK()
}

// search routes search requests
func (r *PropertyRead) search(q *msg.Request, mr *msg.Result) {
	switch q.Property.Type {
	case msg.PropertyCustom:
		r.searchCustom(q, mr)
	case msg.PropertyNative:
		r.listNative(q, mr) // XXX BUG
	case msg.PropertyService:
		r.searchService(q, mr)
	case msg.PropertySystem:
		r.listSystem(q, mr) // XXX BUG
	case msg.PropertyTemplate:
		r.listTemplate(q, mr) // XXX BUG
	default:
		mr.NotImplemented(fmt.Errorf("Unknown property type: %s",
			q.Property.Type))
	}
}

// searchCustom returns all custom properties for a repository
func (r *PropertyRead) searchCustom(q *msg.Request, mr *msg.Result) {
	var (
		property, repository, id string
		rows                     *sql.Rows
		err                      error
	)

	if rows, err = r.stmtListCustom.Query(
		q.Search.Property.Custom.RepositoryID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&id, &repository, &property); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Custom: &proto.PropertyCustom{
				ID:           id,
				RepositoryID: repository,
				Name:         property,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// searchService returns all service properties for a team
func (r *PropertyRead) searchService(q *msg.Request, mr *msg.Result) {
	var (
		id, name, teamID string
		rows             *sql.Rows
		err              error
	)

	if rows, err = r.stmtListService.Query(
		q.Search.Property.Service.TeamID,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&id, &name, &teamID); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Property = append(mr.Property, proto.Property{
			Type: q.Property.Type,
			Service: &proto.PropertyService{
				ID:     id,
				Name:   name,
				TeamID: teamID,
			},
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *PropertyRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
