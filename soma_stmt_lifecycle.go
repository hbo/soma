package main

const lcStmtActiveUnblockCondition = `
SELECT 	scicd.blocked_instance_config_id,
		scicd.blocking_instance_config_id,
		scicd.unblocking_state,
		p.next_status,
		p.check_instance_id
FROM    soma.check_instance_configuration_dependencies scicd
JOIN    soma.check_instance_configurations scic
ON      scicd.blocking_instance_config_id = scic.check_instance_config_id
AND     scicd.unblocking_state = scic.status
JOIN    soma.check_instance_configurations p
ON      scicd.blocked_instance_config_id = scic.check_instance_config_id;`

const lcStmtUpdateInstance = `
UPDATE	soma.check_instances
SET     update_available = $1::boolean,
        current_instance_config_id = $2::uuid
WHERE   check_instance_id = $3::uuid;`

const lcStmtUpdateConfig = `
UPDATE  soma.check_instance_configurations
SET     status = $1::varchar,
        next_status = $2::varchar,
		awaiting_deletion = $3::boolean
WHERE   check_instance_config_id = $4::uuid;`

const lcStmtDeleteDependency = `
DELETE FROM soma.check_instance_configuration_dependencies
WHERE       blocked_instance_config_id = $1::uuid
AND         blocking_instance_config_id = $2::uuid
AND         unblocking_state = $3::varchar;`

const lcStmtReadyDeployments = `
SELECT scic.check_instance_id,
       scic.monitoring_id,
	   sms.monitoring_callback_uri
FROM   soma.check_instance_configurations scic
JOIN   soma.monitoring_systems sms
ON     scic.monitoring_id = sms.monitoring_id
JOIN   soma.check_instances sci
ON     scic.check_instance_id = sci.check_instance_id
AND    scic.check_instance_config_id = sci.current_instance_config_id
WHERE  (  scic.status = 'awaiting_rollout'
       OR scic.status = 'awaiting_deprovision' )
AND    sms.monitoring_callback_uri IS NOT NULL
AND    sci.update_available;`

const lcStmtClearUpdateFlag = `
UPDATE soma.check_instances
SET    update_available = 'false'::boolean
WHERE  check_instance_id = $1::uuid;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix