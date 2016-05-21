/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const LoadAllUserCredentials = `
SELECT aua.user_id,
       aua.crypt,
       aua.reset_pending,
       aua.valid_from,
       aua.valid_until,
       iu.user_uid
FROM   inventory.users iu
JOIN   auth.user_authentication aua
ON     iu.user_id = aua.user_id
WHERE  iu.user_id != '00000000-0000-0000-0000-000000000000'::uuid
AND    NOW() < aua.valid_until
AND    NOT iu.user_is_deleted
AND    iu.user_is_active;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix