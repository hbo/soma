/*-
 * Copyright (c) 2018, Jörg Pernfuß
 * Copyright (c) 2018, 1&1 IONOS SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest

import (
	"net/http"

	"github.com/mjolnir42/soma/internal/msg"
)

// replyBadRequest returns a 400 error
func (x *Rest) replyBadRequest(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.BadRequest(err, q.Section)
	x.send(w, &result)
}

// replyForbidden returns a 403 error
func (x *Rest) replyForbidden(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.Forbidden(err, q.Section)
	x.send(w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
