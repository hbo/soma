/*-
 * Copyright (c) 2016,2018, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

import (
	"github.com/mjolnir42/soma/lib/proto"
)

const (
	TreekeeperStatements = ``

	TreekeeperStartJob = `
UPDATE soma.job
SET    started_at = $2::timestamptz,
       status = 'in_progress'
WHERE  id = $1::uuid
AND    started_at IS NULL;`

	TreekeeperGetViewFromCapability = `
SELECT capability_view
FROM   soma.monitoring_capabilities
WHERE  capability_id = $1::uuid;`

	TreekeeperGetComputedDeployments = `
SELECT scic.check_instance_id,
       scic.check_instance_config_id,
       scic.deployment_details
FROM   soma.checks sc
JOIN   soma.check_instances sci
  ON   sc.check_id = sci.check_id
JOIN   soma.check_instance_configurations scic
  ON   sci.check_instance_id = scic.check_instance_id
WHERE  scic.status = '` + proto.DeploymentComputed + `'::varchar
  AND  sc.repository_id = $1::uuid;`

	TreekeeperGetPreviousDeployment = `
SELECT check_instance_config_id,
       version,
       status,
       deployment_details
FROM   soma.check_instance_configurations
WHERE  status != '` + proto.DeploymentComputed + `'::varchar
AND    status != '` + proto.DeploymentAwaitingComputation + `'::varchar
AND    check_instance_id = $1::uuid
ORDER  BY version DESC
LIMIT  1;`

	TreekeeperUpdateConfigStatus = `
UPDATE soma.check_instance_configurations
SET    status = $1::varchar,
       next_status = $2::varchar,
       status_last_updated_at = NOW()::timestamptz
WHERE  check_instance_config_id = $3::uuid;`

	TreekeeperUpdateCheckInstance = `
UPDATE soma.check_instances
SET    last_configuration_created = $1::timestamptz,
       update_available = $2::boolean,
       current_instance_config_id = $3::uuid
WHERE  check_instance_id = $4::uuid;`

	TreekeeperUpdateExistingCheckInstance = `
UPDATE soma.check_instances
SET    last_configuration_created = $1::timestamptz,
       update_available = $2::boolean
WHERE  check_instance_id = $3::uuid;`

	TreekeeperDeleteDuplicateDetails = `
DELETE FROM soma.check_instance_configurations
WHERE       check_instance_config_id = $1::uuid;`

	TreekeeperSetDependency = `
INSERT INTO soma.check_instance_configuration_dependencies (
            blocked_instance_config_id,
            blocking_instance_config_id,
            unblocking_state)
SELECT $1::uuid,
       $2::uuid,
       $3::varchar;`
)

func init() {
	m[TreekeeperDeleteDuplicateDetails] = `TreekeeperDeleteDuplicateDetails`
	m[TreekeeperGetComputedDeployments] = `TreekeeperGetComputedDeployments`
	m[TreekeeperGetPreviousDeployment] = `TreekeeperGetPreviousDeployment`
	m[TreekeeperGetViewFromCapability] = `TreekeeperGetViewFromCapability`
	m[TreekeeperSetDependency] = `TreekeeperSetDependency`
	m[TreekeeperStartJob] = `TreekeeperStartJob`
	m[TreekeeperUpdateCheckInstance] = `TreekeeperUpdateCheckInstance`
	m[TreekeeperUpdateConfigStatus] = `TreekeeperUpdateConfigStatus`
	m[TreekeeperUpdateExistingCheckInstance] = `TreekeeperUpdateExistingCheckInstance`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
