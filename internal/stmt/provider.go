/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const ProviderVerify = `
SELECT metric_provider
FROM   soma.metric_providers
WHERE  metric_provider = $1::varchar;`

const ProviderList = `
SELECT metric_provider
FROM   soma.metric_providers;`

const ProviderShow = `
SELECT metric_provider
FROM   soma.metric_providers
WHERE  metric_provider = $1::varchar;`

const ProviderAdd = `
INSERT INTO soma.metric_providers (
            metric_provider)
SELECT $1::varchar
WHERE NOT EXISTS (
   SELECT metric_provider
   FROM   soma.metric_providers
   WHERE  metric_provider = $1::varchar);`

const ProviderDel = `
DELETE FROM soma.metric_providers
WHERE  metric_provider = $1::varchar;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix