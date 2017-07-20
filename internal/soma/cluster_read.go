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

type ClusterRead struct {
	Input           chan msg.Request
	Shutdown        chan struct{}
	conn            *sql.DB
	stmtList        *sql.Stmt
	stmtShow        *sql.Stmt
	stmtMemberList  *sql.Stmt
	stmtPropOncall  *sql.Stmt
	stmtPropService *sql.Stmt
	stmtPropSystem  *sql.Stmt
	stmtPropCustom  *sql.Stmt
	appLog          *logrus.Logger
	reqLog          *logrus.Logger
	errLog          *logrus.Logger
}

// newClusterRead returns a new ClusterRead handler with input
// buffer of length
func newClusterRead(length int) (r *ClusterRead) {
	r = &ClusterRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// register initializes resources provided by the Soma app
func (r *ClusterRead) register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// run is the event loop for ClusterRead
func (r *ClusterRead) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.ClusterList:       r.stmtList,
		stmt.ClusterShow:       r.stmtShow,
		stmt.ClusterMemberList: r.stmtMemberList,
		stmt.ClusterOncProps:   r.stmtPropOncall,
		stmt.ClusterSvcProps:   r.stmtPropService,
		stmt.ClusterSysProps:   r.stmtPropSystem,
		stmt.ClusterCstProps:   r.stmtPropCustom,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`cluster`, err, stmt.Name(statement))
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

// process is the request dispatcher
func (r *ClusterRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionMemberList:
		r.memberList(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all clusters
func (r *ClusterRead) list(q *msg.Request, mr *msg.Result) {
	var (
		clusterID, clusterName, bucketID string
		rows                             *sql.Rows
		err                              error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&clusterID,
			&clusterName,
			&bucketID,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Cluster = append(mr.Cluster, proto.Cluster{
			Id:       clusterID,
			Name:     clusterName,
			BucketId: bucketID,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details of a specific cluster
func (r *ClusterRead) show(q *msg.Request, mr *msg.Result) {
	var (
		clusterID, clusterName, clusterState string
		bucketID, teamID                     string
		err                                  error
		tx                                   *sql.Tx
		checkConfigs                         *[]proto.CheckConfig
		cluster                              proto.Cluster
	)

	if err = r.stmtShow.QueryRow(
		q.Cluster.Id,
	).Scan(
		&clusterID,
		&bucketID,
		&clusterName,
		&clusterState,
		&teamID,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		goto fail
	}
	cluster = proto.Cluster{
		Id:          clusterID,
		Name:        clusterName,
		BucketId:    bucketID,
		ObjectState: clusterState,
		TeamId:      teamID,
	}

	// add properties
	cluster.Properties = &[]proto.Property{}

	if err = r.oncallProperties(&cluster); err != nil {
		goto fail
	}
	if err = r.serviceProperties(&cluster); err != nil {
		goto fail
	}
	if err = r.systemProperties(&cluster); err != nil {
		goto fail
	}
	if err = r.customProperties(&cluster); err != nil {
		goto fail
	}

	// add check configuration and instance information
	if tx, err = r.conn.Begin(); err != nil {
		goto fail
	}
	if checkConfigs, err = exportCheckConfigObjectTX(
		tx,
		q.Cluster.Id,
	); err != nil {
		tx.Rollback()
		goto fail
	}
	if checkConfigs != nil && len(*checkConfigs) > 0 {
		cluster.Details = &proto.Details{
			CheckConfigs: checkConfigs,
		}
	}

	mr.Cluster = append(mr.Cluster, cluster)
	mr.OK()
	return

fail:
	mr.ServerError(err, q.Section)
}

// memberList resturns the cluster members
func (r *ClusterRead) memberList(q *msg.Request, mr *msg.Result) {
	var (
		memberNodeID, memberNodeName, clusterName string
		rows                                      *sql.Rows
		err                                       error
	)

	if rows, err = r.stmtMemberList.Query(
		q.Cluster.Id,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	cluster := proto.Cluster{}
	cluster.Id = q.Cluster.Id
	cluster.Members = &[]proto.Node{}

	for rows.Next() {
		if cluster.Members == nil {
			cluster.Members = &[]proto.Node{}
		}

		if err = rows.Scan(
			&memberNodeID,
			&memberNodeName,
			&clusterName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		cluster.Name = clusterName
		*cluster.Members = append(*cluster.Members, proto.Node{
			Id:   memberNodeID,
			Name: memberNodeName,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if len(*cluster.Members) == 0 {
		// trigger ,omitempty in JSON export
		cluster.Members = nil
	}
	mr.Cluster = append(mr.Cluster, cluster)
	mr.OK()
}

// shutdownNow signals the handler to shut down
func (r *ClusterRead) shutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
