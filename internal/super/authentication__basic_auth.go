/*-
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package super // import "github.com/mjolnir42/soma/internal/super"

import (
	"fmt"
	"time"

	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/auth"
)

func (s *Supervisor) validateBasicAuth(q *msg.Request) {
	var tok *svToken
	result := msg.FromRequest(q)

	// basic auth always fails for root if root is disabled
	if q.Super.BasicAuth.User == `root` && s.rootDisabled {
		result.ServerError(
			fmt.Errorf(`Attempted authentication on disabled root account`))
		goto unauthorized
	}

	// basic auth always fails for root if root is restricted and
	// the request comes from an unrestricted endpoint. Note: there
	// are currently no restricted endpoints (https over unix socket)
	if q.Super.BasicAuth.User == `root` && s.rootRestricted && !q.Super.RestrictedEndpoint {
		result.ServerError(
			fmt.Errorf(`Attempted root authentication on unrestricted endpoint`))
		goto unauthorized
	}

	tok = s.tokens.read(q.Super.BasicAuth.Token)
	if tok == nil && !s.readonly {
		// rw instance knows every token
		result.ServerError(fmt.Errorf(`Unknown Token (TokenMap)`))
		goto unauthorized
	} else if tok == nil {
		if !s.fetchTokenFromDB(q.Super.BasicAuth.Token) {
			result.ServerError(fmt.Errorf(`Unknown Token (pgSQL)`))
			goto unauthorized
		}
		tok = s.tokens.read(q.Super.BasicAuth.Token)
	}
	if time.Now().UTC().Before(tok.validFrom.UTC()) ||
		time.Now().UTC().After(tok.expiresAt.UTC()) {
		result.Unauthorized(fmt.Errorf(`Token expired`))
		goto unauthorized
	}

	if auth.VerifyExtracted(q.Super.BasicAuth.User, q.RemoteAddr, tok.binToken, s.key,
		s.seed, tok.binExpiresAt, tok.salt) {
		// valid token
		result.Super = &msg.Supervisor{Verdict: 200}
		result.OK()
		q.Reply <- result
	}

unauthorized:
	result.Super = &msg.Supervisor{Verdict: 401}
	q.Reply <- result
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
