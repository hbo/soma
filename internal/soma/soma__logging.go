/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/msg"
)

func logRequest(l *logrus.Logger, q *msg.Request) {
	l.WithField(`RequestID`, q.ID.String()).
		WithField(`IPAddr`, q.RemoteAddr).
		WithField(`UserName`, q.AuthUser).
		WithField(`Section`, q.Section).
		WithField(`Action`, q.Action).
		WithField(`Request`, fmt.Sprintf("%s::%s", q.Section, q.Action)).
		WithField(`Phase`, `request`).
		Debug(`received:soma`)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
