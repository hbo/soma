/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const GroupBucketId = `
SELECT sg.bucket_id
FROM   soma.groups sg
WHERE  sg.group_id = $1;`

const GroupOncProps = `
SELECT op.instance_id,
       op.source_instance_id,
       op.view,
       op.oncall_duty_id,
       iodt.oncall_duty_name
FROM   soma.group_oncall_properties op
JOIN   inventory.oncall_duty_teams iodt
  ON   op.oncall_duty_id = iodt.oncall_duty_id
WHERE  op.group_id = $1::uuid;`

const GroupSvcProps = `
SELECT sp.instance_id,
       sp.source_instance_id,
       sp.view,
       sp.service_property
FROM   soma.group_service_properties sp
WHERE  sp.group_id = $1::uuid;`

const GroupSysProps = `
SELECT sp.instance_id,
       sp.source_instance_id,
       sp.view,
       sp.system_property,
       sp.value
FROM   soma.group_system_properties sp
WHERE  sp.group_id = $1::uuid;`

const GroupCstProps = `
SELECT cp.instance_id,
       cp.source_instance_id,
       cp.view,
       cp.custom_property_id,
       cp.value,
       scp.custom_property
FROM   soma.group_custom_properties cp
JOIN   soma.custom_properties scp
  ON   cp.custom_property_id = scp.custom_property_id
WHERE  cp.group_id = $1::uuid;`

const GroupSystemPropertyForDelete = `
SELECT view,
       system_property,
       value
FROM   soma.group_system_properties
WHERE  source_instance_id = $1::uuid
  AND  source_instance_id = instance_id;`

const GroupCustomPropertyForDelete = `
SELECT sgcp.view,
       sgcp.custom_property_id,
       sgcp.value,
       scp.custom_property
FROM   soma.group_custom_properties sgcp
JOIN   soma.custom_properties scp
  ON   sgcp.repository_id = scp.repository_id
 AND   sgcp.custom_property_id = scp.custom_property_id
WHERE  sgcp.source_instance_id = $1::uuid
  AND  sgcp.source_instance_id = sgcp.instance_id;`

const GroupOncallPropertyForDelete = `
SELECT sgop.view,
       sgop.oncall_duty_id,
       iodt.oncall_duty_name,
       iodt.oncall_duty_phone_number
FROM   soma.group_oncall_properties sgop
JOIN   inventory.oncall_duty_teams iodt
  ON   sgop.oncall_duty_id = iodt.oncall_duty_id
WHERE  sgop.source_instance_id = $1::uuid
  AND  sgop.source_instance_id = sgop.instance_id;`

const GroupServicePropertyForDelete = `
SELECT sgsp.view,
       sgsp.service_property
FROM   soma.group_service_properties sgsp
JOIN   soma.team_service_properties stsp
  ON   sgsp.organizational_team_id = stsp.organizational_team_id
 AND   sgsp.service_property = stsp.service_property
WHERE  sgsp.source_instance_id = $1::uuid
  AND  sgsp.source_instance_id = sgsp.instance_id;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix