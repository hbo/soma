/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package super // import "github.com/mjolnir42/soma/internal/super"

import (
	"database/sql"

	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

func (s *Supervisor) rightRead(q *msg.Request, mr *msg.Result) {
	switch q.Action {
	case msg.ActionSearch:
		switch q.Grant.Category {
		case msg.CategorySystem,
			msg.CategoryGlobal,
			msg.CategoryGrantGlobal,
			msg.CategoryPermission,
			msg.CategoryGrantPermission,
			msg.CategoryOperation,
			msg.CategoryGrantOperation:
			s.rightSearchGlobal(q, mr)
		case msg.CategoryRepository,
			msg.CategoryGrantRepository,
			msg.CategoryTeam,
			msg.CategoryGrantTeam,
			msg.CategoryMonitoring,
			msg.CategoryGrantMonitoring:
			s.rightSearchScoped(q, mr)
		}
	}
}

func (s *Supervisor) rightSearchGlobal(q *msg.Request, mr *msg.Result) {
	var (
		err     error
		grantID string
	)

	if err = s.stmtSearchAuthorizationGlobal.QueryRow(
		q.Grant.PermissionId,
		q.Grant.Category,
		q.Grant.RecipientId,
		q.Grant.RecipientType,
	).Scan(
		&grantID,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Grant = append(mr.Grant, proto.Grant{
		Id:            grantID,
		PermissionId:  q.Grant.PermissionId,
		Category:      q.Grant.Category,
		RecipientId:   q.Grant.RecipientId,
		RecipientType: q.Grant.RecipientType,
	})
	mr.OK()
}

func (s *Supervisor) rightSearchScoped(q *msg.Request, mr *msg.Result) {
	var (
		err     error
		grantID string
		scope   *sql.Stmt
	)

	switch q.Grant.Category {
	case msg.CategoryRepository,
		msg.CategoryGrantRepository:
		scope = s.stmtSearchAuthorizationRepository
	case msg.CategoryTeam,
		msg.CategoryGrantTeam:
		scope = s.stmtSearchAuthorizationTeam
	case msg.CategoryMonitoring,
		msg.CategoryGrantMonitoring:
		scope = s.stmtSearchAuthorizationMonitoring
	}
	if err = scope.QueryRow(
		q.Grant.PermissionId,
		q.Grant.Category,
		q.Grant.RecipientId,
		q.Grant.RecipientType,
		q.Grant.ObjectType,
		q.Grant.ObjectId,
	).Scan(
		&grantID,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Grant = append(mr.Grant, proto.Grant{
		Id:            grantID,
		PermissionId:  q.Grant.PermissionId,
		Category:      q.Grant.Category,
		RecipientId:   q.Grant.RecipientId,
		RecipientType: q.Grant.RecipientType,
		ObjectType:    q.Grant.ObjectType,
		ObjectId:      q.Grant.ObjectId,
	})
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix