/*-
Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package super // import "github.com/mjolnir42/soma/internal/super"

import (
	"database/sql"
	"time"

	"github.com/mjolnir42/scrypth64"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	uuid "github.com/satori/go.uuid"
)

func (s *Supervisor) startupLoad() {

	s.startupRoot()

	if !s.readonly {
		s.startupCredentials()
	}

	s.startupTokens()
}

func (s *Supervisor) startupRoot() {
	var (
		err                  error
		flag, crypt          string
		mcf                  scrypth64.Mcf
		validFrom, expiresAt time.Time
		state                bool
		rows                 *sql.Rows
	)

	rows, err = s.conn.Query(stmt.LoadRootFlags)
	if err != nil {
		s.errLog.Fatal(`supervisor/load-root-flags,query: `, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&flag,
			&state,
		); err != nil {
			s.errLog.Fatal(`supervisor/load-root-flags,scan: `, err)
		}
		switch flag {
		case `disabled`:
			s.rootDisabled = state
		case `restricted`:
			s.rootRestricted = state
		}
	}
	if err = rows.Err(); err != nil {
		s.errLog.Fatal(`supervisor/load-root-flags,next: `, err)
	}

	// only load root credentials on master instance
	if !s.readonly {
		if err = s.conn.QueryRow(stmt.LoadRootPassword).Scan(
			&crypt,
			&validFrom,
			&expiresAt,
		); err == sql.ErrNoRows {
			// root bootstrap outstanding
			return
		} else if err != nil {
			s.errLog.Fatal(`supervisor/load-root-password: `, err)
		}
		if mcf, err = scrypth64.FromString(crypt); err != nil {
			s.errLog.Fatal(`supervisor/string-to-mcf: `, err)
		}
		s.credentials.insert(`root`, uuid.Nil, validFrom.UTC(),
			msg.PosTimeInf.UTC(), mcf)
	}
}

func (s *Supervisor) startupCredentials() {
	var (
		err                  error
		rows                 *sql.Rows
		userID, user, crypt  string
		reset                bool
		validFrom, expiresAt time.Time
		id                   uuid.UUID
		mcf                  scrypth64.Mcf
	)

	rows, err = s.conn.Query(stmt.LoadAllUserCredentials)
	if err != nil {
		s.errLog.Fatal(`supervisor/load-credentials,query: `, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&userID,
			&crypt,
			&reset,
			&validFrom,
			&expiresAt,
			&user,
		); err != nil {
			s.errLog.Fatal(`supervisor/load-credentials,scan: `, err)
		}

		if id, err = uuid.FromString(userID); err != nil {
			s.errLog.Fatal(`supervisor/string-to-uuid: `, err)
		}
		if mcf, err = scrypth64.FromString(crypt); err != nil {
			s.errLog.Fatal(`supervisor/string-to-mcf: `, err)
		}

		s.credentials.restore(user, id, validFrom, expiresAt, mcf, reset, true)
	}
	if err = rows.Err(); err != nil {
		s.errLog.Fatal(`supervisor/load-credentials,next: `, err)
	}
}

func (s *Supervisor) startupTokens() {
	var (
		err                         error
		token, salt, valid, expires string
		validFrom, expiresAt        time.Time
		rows                        *sql.Rows
	)

	rows, err = s.conn.Query(stmt.LoadAllTokens)
	if err != nil {
		s.errLog.Fatal(`supervisor/load-tokens,query: `, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&token,
			&salt,
			&validFrom,
			&expiresAt,
		); err != nil {
			s.errLog.Fatal(`supervisor/load-tokens,scan: `, err)
		}
		valid = validFrom.Format(msg.RFC3339Milli)
		expires = expiresAt.Format(msg.RFC3339Milli)

		if err = s.tokens.insert(token, valid, expires, salt); err != nil {
			s.errLog.Fatal(`supervisor/load-tokens,insert: `, err)
		}
	}
	if err = rows.Err(); err != nil {
		s.errLog.Fatal(`supervisor/load-tokens,next: `, err)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
