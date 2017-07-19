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

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// RepositoryRead handles read requests for buckets
type RepositoryRead struct {
	Input           chan msg.Request
	Shutdown        chan struct{}
	conn            *sql.DB
	stmtList        *sql.Stmt
	stmtShow        *sql.Stmt
	stmtPropOncall  *sql.Stmt
	stmtPropService *sql.Stmt
	stmtPropSystem  *sql.Stmt
	stmtPropCustom  *sql.Stmt
	appLog          *logrus.Logger
	reqLog          *logrus.Logger
	errLog          *logrus.Logger
}

// newBucketRead returns a new BucketRead handler with input
// buffer of length
func newRepositoryRead(length int) (r *RepositoryRead) {
	r = &RepositoryRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// register initializes resources provided by the Soma app
func (r *RepositoryRead) register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// run is the event loop for RepositoryRead
func (r *RepositoryRead) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.ListAllRepositories: r.stmtList,
		stmt.ShowRepository:      r.stmtShow,
		stmt.RepoOncProps:        r.stmtPropOncall,
		stmt.RepoSvcProps:        r.stmtPropService,
		stmt.RepoSysProps:        r.stmtPropSystem,
		stmt.RepoCstProps:        r.stmtPropCustom,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`repository`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
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

// process is the event dispatcher for RepositoryRead
func (r *RepositoryRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case `list`:
		r.list(q, &result)
	case `show`:
		r.show(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all repositories
func (r *RepositoryRead) list(q *msg.Request, mr *msg.Result) {
	var (
		repoID, repoName string
		rows             *sql.Rows
		err              error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&repoID,
			&repoName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Repository = append(mr.Repository, proto.Repository{
			Id:   repoID,
			Name: repoName,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details of a specific repository
func (r *RepositoryRead) show(q *msg.Request, mr *msg.Result) {
	var (
		repoID, repoName, teamID string
		isActive                 bool
		err                      error
	)

	if err = r.stmtShow.QueryRow(
		q.Repository.Id,
	).Scan(
		&repoID,
		&repoName,
		&isActive,
		&teamID,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	repo := proto.Repository{
		Id:        repoID,
		Name:      repoName,
		TeamId:    teamID,
		IsDeleted: false,
		IsActive:  isActive,
	}

	// add properties
	repo.Properties = &[]proto.Property{}

	if err = r.propertyOncall(q, &repo); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if err = r.propertyService(q, &repo); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if err = r.propertySystem(q, &repo); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if err = r.propertyCustom(q, &repo); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	mr.Repository = append(mr.Repository, repo)
	mr.OK()
}

// propertyOncall adds the oncall properties to a repository
func (r *RepositoryRead) propertyOncall(q *msg.Request, o *proto.Repository) error {
	var (
		instanceID, sourceInstanceID string
		view, oncallID, oncallName   string
		rows                         *sql.Rows
		err                          error
	)

	if rows, err = r.stmtPropOncall.Query(
		q.Repository.Id,
	); err != nil {
		return err
	}

	for rows.Next() {
		if err = rows.Scan(
			&instanceID,
			&sourceInstanceID,
			&view,
			&oncallID,
			&oncallName,
		); err != nil {
			rows.Close()
			return err
		}
		*o.Properties = append(*o.Properties,
			proto.Property{
				Type:             `oncall`,
				RepositoryId:     q.Repository.Id,
				InstanceId:       instanceID,
				SourceInstanceId: sourceInstanceID,
				View:             view,
				Oncall: &proto.PropertyOncall{
					Id:   oncallID,
					Name: oncallName,
				},
			},
		)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

// propertyService adds the service properties to a repository
func (r *RepositoryRead) propertyService(q *msg.Request, o *proto.Repository) error {
	var (
		instanceID, sourceInstanceID string
		serviceName, view            string
		rows                         *sql.Rows
		err                          error
	)

	if rows, err = r.stmtPropService.Query(
		q.Repository.Id,
	); err != nil {
		return err
	}

	for rows.Next() {
		if err = rows.Scan(
			&instanceID,
			&sourceInstanceID,
			&view,
			&serviceName,
		); err != nil {
			rows.Close()
			return err
		}
		*o.Properties = append(*o.Properties,
			proto.Property{
				Type:             `service`,
				RepositoryId:     q.Repository.Id,
				InstanceId:       instanceID,
				SourceInstanceId: sourceInstanceID,
				View:             view,
				Service: &proto.PropertyService{
					Name: serviceName,
				},
			},
		)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

// propertySystem adds the system properties to a repository
func (r *RepositoryRead) propertySystem(q *msg.Request, o *proto.Repository) error {
	var (
		instanceID, sourceInstanceID, view string
		systemProp, value                  string
		rows                               *sql.Rows
		err                                error
	)

	if rows, err = r.stmtPropSystem.Query(
		q.Repository.Id,
	); err != nil {
		return err
	}

	for rows.Next() {
		if err = rows.Scan(
			&instanceID,
			&sourceInstanceID,
			&view,
			&systemProp,
			&value,
		); err != nil {
			rows.Close()
			return err
		}
		*o.Properties = append(*o.Properties,
			proto.Property{
				Type:             `system`,
				RepositoryId:     q.Repository.Id,
				InstanceId:       instanceID,
				SourceInstanceId: sourceInstanceID,
				View:             view,
				System: &proto.PropertySystem{
					Name:  systemProp,
					Value: value,
				},
			},
		)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

// propertyCustom adds the custom properties to a repository
func (r *RepositoryRead) propertyCustom(q *msg.Request, o *proto.Repository) error {
	var (
		instanceID, sourceInstanceID, view string
		customID, value, customProp        string
		rows                               *sql.Rows
		err                                error
	)

	if rows, err = r.stmtPropCustom.Query(
		q.Repository.Id,
	); err != nil {
		return err
	}

	for rows.Next() {
		if err = rows.Scan(
			&instanceID,
			&sourceInstanceID,
			&view,
			&customID,
			&value,
			&customProp,
		); err != nil {
			rows.Close()
			return err
		}
		*o.Properties = append(*o.Properties,
			proto.Property{
				Type:             `custom`,
				RepositoryId:     q.Repository.Id,
				InstanceId:       instanceID,
				SourceInstanceId: sourceInstanceID,
				View:             view,
				Custom: &proto.PropertyCustom{
					Id:    customID,
					Name:  customProp,
					Value: value,
				},
			},
		)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

// shutdown signals the handler to shut down
func (r *RepositoryRead) shutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
