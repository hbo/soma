/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

// insert a new token into the database
const InsertToken = `
INSERT INTO auth.tokens (
    token,
    salt,
    valid_from,
    valid_until
) VALUES (
    $1::varchar,
    $2::varchar,
    $3::timestamptz,
    $4::timestamptz
);`

// lookup a specific token (readonly instances)
const SelectToken = `
SELECT salt,
       valid_from,
       valid_until
FROM   auth.tokens
WHERE  token = $1::varchar;`

// startup loading all tokens
const LoadAllTokens = `
SELECT token,
       salt,
       valid_from,
       valid_until
FROM   auth.tokens
WHERE  NOW() < valid_until;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
