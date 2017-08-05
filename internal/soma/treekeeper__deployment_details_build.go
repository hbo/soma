package soma

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mjolnir42/soma/lib/proto"
)

func (tk *TreeKeeper) buildDeploymentDetails() {
	var (
		err                                                 error
		instanceCfgID                                       string
		objID, objType                                      string
		rows, thresh, pkgs, gSysProps, cSysProps, nSysProps *sql.Rows
		gCustProps, cCustProps, nCustProps                  *sql.Rows
		callback                                            sql.NullString
	)

	// TODO:
	// * refactoring switch objType {} block
	// * SQL error handling

	if rows, err = tk.stmtList.Query(tk.meta.repoID); err != nil {
		tk.status.isBroken = true
		return
	}
	defer rows.Close()

deploymentbuilder:
	for rows.Next() {
		detail := proto.Deployment{}

		err = rows.Scan(
			&instanceCfgID,
		)
		if err != nil {
			tk.treeLog.Println(`tk.stmtList.Query().Scan():`, err)
			break deploymentbuilder
		}

		//
		detail.CheckInstance = &proto.CheckInstance{
			InstanceConfigId: instanceCfgID,
		}
		tk.stmtCheckInstance.QueryRow(instanceCfgID).Scan(
			&detail.CheckInstance.Version,
			&detail.CheckInstance.InstanceId,
			&detail.CheckInstance.ConstraintHash,
			&detail.CheckInstance.ConstraintValHash,
			&detail.CheckInstance.InstanceService,
			&detail.CheckInstance.InstanceSvcCfgHash,
			&detail.CheckInstance.InstanceServiceConfig,
			&detail.CheckInstance.CheckId,
			&detail.CheckInstance.ConfigId,
		)

		//
		detail.Check = &proto.Check{
			CheckId: detail.CheckInstance.CheckId,
		}
		tk.stmtCheck.QueryRow(detail.CheckInstance.CheckId).Scan(
			&detail.Check.RepositoryId,
			&detail.Check.SourceCheckId,
			&detail.Check.SourceType,
			&detail.Check.InheritedFrom,
			&detail.Check.CapabilityId,
			&objID,
			&objType,
			&detail.Check.Inheritance,
			&detail.Check.ChildrenOnly,
		)
		detail.ObjectType = objType
		if detail.Check.InheritedFrom != objID {
			detail.Check.IsInherited = true
		}
		detail.Check.CheckConfigId = detail.CheckInstance.ConfigId

		//
		detail.CheckConfig = &proto.CheckConfig{
			Id:           detail.Check.CheckConfigId,
			RepositoryId: detail.Check.RepositoryId,
			BucketId:     detail.Check.BucketId,
			CapabilityId: detail.Check.CapabilityId,
			ObjectId:     objID,
			ObjectType:   objType,
			Inheritance:  detail.Check.Inheritance,
			ChildrenOnly: detail.Check.ChildrenOnly,
		}
		tk.stmtCheckConfig.QueryRow(detail.Check.CheckConfigId).Scan(
			&detail.CheckConfig.Name,
			&detail.CheckConfig.Interval,
			&detail.CheckConfig.IsActive,
			&detail.CheckConfig.IsEnabled,
			&detail.CheckConfig.ExternalId,
		)

		//
		detail.CheckConfig.Thresholds = []proto.CheckConfigThreshold{}
		thresh, err = tk.stmtThreshold.Query(detail.CheckConfig.Id)
		if err != nil {
			// a check config must have 1+ thresholds
			tk.treeLog.Println(`DANGER WILL ROBINSON!`,
				`Failed to get thresholds for:`, detail.CheckConfig.Id)
			continue deploymentbuilder
		}
		defer thresh.Close()

		for thresh.Next() {
			thr := proto.CheckConfigThreshold{
				Predicate: proto.Predicate{},
				Level:     proto.Level{},
			}

			err = thresh.Scan(
				&thr.Predicate.Symbol,
				&thr.Value,
				&thr.Level.Name,
				&thr.Level.ShortName,
				&thr.Level.Numeric,
			)
			if err != nil {
				tk.treeLog.Println(`tk.stmtThreshold.Query().Scan():`, err)
				break deploymentbuilder
			}
			detail.CheckConfig.Thresholds = append(detail.CheckConfig.Thresholds, thr)
		}

		// XXX TODO
		//detail.CheckConfiguration.Constraints = []somaproto.CheckConfigurationConstraint{}
		detail.CheckConfig.Constraints = nil

		//
		detail.Capability = &proto.Capability{
			Id: detail.Check.CapabilityId,
		}
		detail.Monitoring = &proto.Monitoring{}
		detail.Metric = &proto.Metric{}
		detail.Unit = &proto.Unit{}
		tk.stmtCapMonMetric.QueryRow(detail.Capability.Id).Scan(
			&detail.Capability.Metric,
			&detail.Capability.MonitoringId,
			&detail.Capability.View,
			&detail.Capability.Thresholds,
			&detail.Monitoring.Name,
			&detail.Monitoring.Mode,
			&detail.Monitoring.Contact,
			&detail.Monitoring.TeamId,
			&callback,
			&detail.Metric.Unit,
			&detail.Metric.Description,
			&detail.Unit.Name,
		)
		if callback.Valid {
			detail.Monitoring.Callback = callback.String
		} else {
			detail.Monitoring.Callback = ""
		}
		detail.Unit.Unit = detail.Metric.Unit
		detail.Metric.Path = detail.Capability.Metric
		detail.Monitoring.Id = detail.Capability.MonitoringId
		detail.Capability.Name = fmt.Sprintf("%s.%s.%s",
			detail.Monitoring.Name,
			detail.Capability.View,
			detail.Metric.Path,
		)
		detail.View = detail.Capability.View

		//
		detail.Metric.Packages = &[]proto.MetricPackage{}
		pkgs, _ = tk.stmtPkgs.Query(detail.Metric.Path)
		defer pkgs.Close()

		for pkgs.Next() {
			pkg := proto.MetricPackage{}

			err = pkgs.Scan(
				&pkg.Provider,
				&pkg.Name,
			)
			if err != nil {
				tk.treeLog.Println(`tk.stmtPkgs.Query().Scan():`, err)
				break deploymentbuilder
			}
			*detail.Metric.Packages = append(*detail.Metric.Packages, pkg)
		}

		//
		detail.Oncall = &proto.Oncall{}
		detail.Service = &proto.PropertyService{}
		switch objType {
		case "group":
			// fetch the group object
			detail.Group = &proto.Group{
				Id: objID,
			}
			tk.stmtGroup.QueryRow(objID).Scan(
				&detail.Group.BucketId,
				&detail.Group.Name,
				&detail.Group.ObjectState,
				&detail.Group.TeamId,
				&detail.Bucket,
				&detail.Environment,
				&detail.Repository,
			)
			// fetch team information
			detail.Team = &proto.Team{
				Id: detail.Group.TeamId,
			}
			// fetch oncall information if the property is set,
			// otherwise cleanup detail.Oncall
			err = tk.stmtGroupOncall.QueryRow(detail.Group.Id, detail.View).Scan(
				&detail.Oncall.Id,
				&detail.Oncall.Name,
				&detail.Oncall.Number,
			)
			if err == sql.ErrNoRows {
				detail.Oncall = nil
			} else if err != nil {
				tk.treeLog.Println(`tk.stmtGroupOncall.QueryRow():`, err)
				break deploymentbuilder
			}
			// fetch service name, and attributes if applicable
			if detail.CheckInstance.InstanceService != "" {
				err = tk.stmtGroupService.QueryRow(
					detail.CheckInstance.InstanceService,
					detail.View,
				).Scan(
					&detail.Service.Name,
					&detail.Service.TeamId,
				)
				if err == sql.ErrNoRows {
					detail.Service = nil
				} else if err != nil {
					tk.treeLog.Println(`tk.stmtGroupService.QueryRow():`, err)
					break deploymentbuilder
				} else {
					detail.Service.Attributes = []proto.ServiceAttribute{}
					fm := map[string]string{}
					_ = json.Unmarshal([]byte(detail.CheckInstance.InstanceServiceConfig), &fm)
					for k, v := range fm {
						a := proto.ServiceAttribute{
							Name:  k,
							Value: v,
						}
						detail.Service.Attributes = append(detail.Service.Attributes, a)
					}
				}
			}
			// fetch system properties
			detail.Properties = &[]proto.PropertySystem{}
			gSysProps, _ = tk.stmtGroupSysProp.Query(detail.Group.Id, detail.View)
			defer gSysProps.Close()

			for gSysProps.Next() {
				prop := proto.PropertySystem{}
				err = gSysProps.Scan(
					&prop.Name,
					&prop.Value,
				)
				if err != nil {
					tk.treeLog.Println(`tk.stmtGroupSysProp.Query().Scan():`, err)
					break deploymentbuilder
				}
				*detail.Properties = append(*detail.Properties, prop)
				if prop.Name == "group_datacenter" {
					detail.Datacenter = prop.Value
				}
			}
			if len(*detail.Properties) == 0 {
				detail.Properties = nil
			}
			// fetch custom properties
			detail.CustomProperties = &[]proto.PropertyCustom{}
			gCustProps, _ = tk.stmtGroupCustProp.Query(detail.Group.Id, detail.View)
			defer gCustProps.Close()

			for gCustProps.Next() {
				prop := proto.PropertyCustom{}
				err = gCustProps.Scan(
					&prop.Id,
					&prop.Name,
					&prop.Value,
				)
				if err != nil {
					tk.treeLog.Println(`tk.stmtGroupCustProp.Query().Scan():`, err)
					break deploymentbuilder
				}
				*detail.CustomProperties = append(*detail.CustomProperties, prop)
			}
			if len(*detail.CustomProperties) == 0 {
				detail.CustomProperties = nil
			}
		case "cluster":
			// fetch the cluster object
			detail.Cluster = &proto.Cluster{
				Id: objID,
			}
			tk.stmtCluster.QueryRow(objID).Scan(
				&detail.Cluster.Name,
				&detail.Cluster.BucketId,
				&detail.Cluster.ObjectState,
				&detail.Cluster.TeamId,
				&detail.Bucket,
				&detail.Environment,
				&detail.Repository,
			)
			// fetch team information
			detail.Team = &proto.Team{
				Id: detail.Cluster.TeamId,
			}
			// fetch oncall information if the property is set,
			// otherwise cleanup detail.Oncall
			err = tk.stmtClusterOncall.QueryRow(detail.Cluster.Id, detail.View).Scan(
				&detail.Oncall.Id,
				&detail.Oncall.Name,
				&detail.Oncall.Number,
			)
			if err != nil {
				detail.Oncall = nil
			}
			// fetch the service name, and attributes if applicable
			if detail.CheckInstance.InstanceService != "" {
				err = tk.stmtClusterService.QueryRow(
					detail.CheckInstance.InstanceService,
					detail.View,
				).Scan(
					&detail.Service.Name,
					&detail.Service.TeamId,
				)
				if err != nil {
					detail.Service = nil
				} else {
					detail.Service.Attributes = []proto.ServiceAttribute{}
					fm := map[string]string{}
					_ = json.Unmarshal([]byte(detail.CheckInstance.InstanceServiceConfig), &fm)
					for k, v := range fm {
						a := proto.ServiceAttribute{
							Name:  k,
							Value: v,
						}
						detail.Service.Attributes = append(detail.Service.Attributes, a)
					}
				}
			}
			// fetch system properties
			detail.Properties = &[]proto.PropertySystem{}
			cSysProps, _ = tk.stmtClusterSysProp.Query(detail.Cluster.Id, detail.View)
			defer cSysProps.Close()

			for cSysProps.Next() {
				prop := proto.PropertySystem{}
				err = cSysProps.Scan(
					&prop.Name,
					&prop.Value,
				)
				if err != nil {
					tk.treeLog.Println(`tk.stmtClusterSysProp.Query().Scan():`, err)
					break deploymentbuilder
				}
				*detail.Properties = append(*detail.Properties, prop)
				if prop.Name == "cluster_datacenter" {
					detail.Datacenter = prop.Value
				}
			}
			if len(*detail.Properties) == 0 {
				detail.Properties = nil
			}
			// fetch custom properties
			detail.CustomProperties = &[]proto.PropertyCustom{}
			cCustProps, _ = tk.stmtClusterCustProp.Query(detail.Cluster.Id, detail.View)
			defer cCustProps.Close()

			for cCustProps.Next() {
				prop := proto.PropertyCustom{}
				cCustProps.Scan(
					&prop.Id,
					&prop.Name,
					&prop.Value,
				)
				*detail.CustomProperties = append(*detail.CustomProperties, prop)
			}
			if len(*detail.CustomProperties) == 0 {
				detail.CustomProperties = nil
			}
		case "node":
			// fetch the node object
			detail.Server = &proto.Server{}
			detail.Node = &proto.Node{
				Id: objID,
			}
			tk.stmtNode.QueryRow(objID).Scan(
				&detail.Node.AssetId,
				&detail.Node.Name,
				&detail.Node.TeamId,
				&detail.Node.ServerId,
				&detail.Node.State,
				&detail.Node.IsOnline,
				&detail.Node.IsDeleted,
				&detail.Bucket,
				&detail.Environment,
				&detail.Repository,
				&detail.Server.AssetId,
				&detail.Server.Datacenter,
				&detail.Server.Location,
				&detail.Server.Name,
				&detail.Server.IsOnline,
				&detail.Server.IsDeleted,
			)
			detail.Server.Id = detail.Node.ServerId
			detail.Datacenter = detail.Server.Datacenter
			// fetch team information
			detail.Team = &proto.Team{
				Id: detail.Node.TeamId,
			}
			// fetch oncall information if the property is set,
			// otherwise cleanup detail.Oncall
			err = tk.stmtNodeOncall.QueryRow(detail.Node.Id, detail.View).Scan(
				&detail.Oncall.Id,
				&detail.Oncall.Name,
				&detail.Oncall.Number,
			)
			if err != nil {
				detail.Oncall = nil
			}
			// fetch the service name, and attributes if applicable
			if detail.CheckInstance.InstanceService != "" {
				err = tk.stmtNodeService.QueryRow(
					detail.CheckInstance.InstanceService,
					detail.View,
				).Scan(
					&detail.Service.Name,
					&detail.Service.TeamId,
				)
				if err != nil {
					detail.Service = nil
				} else {
					detail.Service.Attributes = []proto.ServiceAttribute{}
					fm := map[string]string{}
					_ = json.Unmarshal([]byte(detail.CheckInstance.InstanceServiceConfig), &fm)
					for k, v := range fm {
						a := proto.ServiceAttribute{
							Name:  k,
							Value: v,
						}
						detail.Service.Attributes = append(detail.Service.Attributes, a)
					}
				}
			}
			// fetch system properties
			detail.Properties = &[]proto.PropertySystem{}
			nSysProps, _ = tk.stmtNodeSysProp.Query(detail.Node.Id, detail.View)
			defer nSysProps.Close()

			for nSysProps.Next() {
				prop := proto.PropertySystem{}
				err = nSysProps.Scan(
					&prop.Name,
					&prop.Value,
				)
				if err != nil {
					tk.treeLog.Println(`tk.stmtNodeSysProp.Query().Scan():`, err)
					break deploymentbuilder
				}
				*detail.Properties = append(*detail.Properties, prop)
			}
			if len(*detail.Properties) == 0 {
				detail.Properties = nil
			}
			// fetch custom properties
			detail.CustomProperties = &[]proto.PropertyCustom{}
			nCustProps, _ = tk.stmtNodeCustProp.Query(detail.Node.Id, detail.View)
			defer nCustProps.Close()

			for nCustProps.Next() {
				prop := proto.PropertyCustom{}
				nCustProps.Scan(
					&prop.Id,
					&prop.Name,
					&prop.Value,
				)
				*detail.CustomProperties = append(*detail.CustomProperties, prop)
			}
			if len(*detail.CustomProperties) == 0 {
				detail.CustomProperties = nil
			}
		}

		tk.stmtTeam.QueryRow(detail.Team.Id).Scan(
			&detail.Team.Name,
			&detail.Team.LdapId,
		)

		// if no datacenter information was gathered, use the default DC
		if detail.Datacenter == "" {
			tk.stmtDefaultDC.QueryRow().Scan(&detail.Datacenter)
		}

		// build JSON of DeploymentDetails
		var detailJSON []byte
		if detailJSON, err = json.Marshal(&detail); err != nil {
			tk.treeLog.Println(`Failed to JSON marshal deployment details:`,
				detail.CheckInstance.InstanceConfigId, err)
			break deploymentbuilder
		}
		if _, err = tk.stmtUpdate.Exec(
			detailJSON,
			detail.Monitoring.Id,
			detail.CheckInstance.InstanceConfigId,
		); err != nil {
			tk.treeLog.Println(`Failed to save DeploymentDetails.JSON:`,
				detail.CheckInstance.InstanceConfigId, err)
			break deploymentbuilder
		}
	}
	// mark the tree as broken to prevent further data processing
	if err != nil {
		tk.status.isBroken = true
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
