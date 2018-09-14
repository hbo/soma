/*-
 * Copyright (c) 2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"github.com/julienschmidt/httprouter"
)

// setupRouter returns a configured httprouter
func (x *Rest) setupRouter() *httprouter.Router {
	router := httprouter.New()

	router.HEAD(`/`, x.CheckShutdown(x.Ping))

	router.GET(`/attribute/:attribute`, x.Verify(x.AttributeShow))
	router.GET(`/attribute/`, x.Verify(x.AttributeList))
	router.GET(`/capability/:capabilityID`, x.Verify(x.CapabilityShow))
	router.GET(`/capability/`, x.Verify(x.CapabilityList))
	router.GET(`/category/:category/section/:sectionID/action/:actionID`, x.Verify(x.ActionShow))
	router.GET(`/category/:category/section/:sectionID/action/`, x.Verify(x.ActionList))
	router.GET(`/category/:category/section/:sectionID`, x.Verify(x.SectionShow))
	router.GET(`/category/:category/section/`, x.Verify(x.SectionList))
	router.GET(`/category/:category`, x.Verify(x.CategoryShow))
	router.GET(`/category/`, x.Verify(x.CategoryList))
	router.GET(`/checkconfig/:repositoryID/:checkID`, x.Verify(x.CheckConfigShow))
	router.GET(`/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigList))
	router.GET(`/datacenter/:datacenter`, x.Verify(x.DatacenterShow))
	router.GET(`/datacenter/`, x.Verify(x.DatacenterList))
	router.GET(`/entity/:entity`, x.Verify(x.EntityShow))
	router.GET(`/entity/`, x.Verify(x.EntityList))
	router.GET(`/environment/:environment`, x.Verify(x.EnvironmentShow))
	router.GET(`/environment/`, x.Verify(x.EnvironmentList))
	router.GET(`/hostdeployment/:monitoringID/:assetID`, x.CheckShutdown(x.HostDeploymentFetch))
	router.GET(`/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/instance/:instanceID`, x.Verify(x.ScopeSelectInstanceShow))
	router.GET(`/instance/`, x.Verify(x.ScopeSelectInstanceList))
	router.GET(`/job/:jobID`, x.Verify(x.JobShow))
	router.GET(`/job/`, x.Verify(x.ScopeSelectJobList))
	router.GET(`/level/:level`, x.Verify(x.LevelShow))
	router.GET(`/level/`, x.Verify(x.LevelList))
	router.GET(`/metric/:metric`, x.Verify(x.MetricShow))
	router.GET(`/metric/`, x.Verify(x.MetricList))
	router.GET(`/mode/:mode`, x.Verify(x.ModeShow))
	router.GET(`/mode/`, x.Verify(x.ModeList))
	router.GET(`/monitoringsystem/:monitoringID`, x.Verify(x.MonitoringShow))
	router.GET(`/monitoringsystem/`, x.Verify(x.ScopeSelectMonitoringList))
	router.GET(`/oncall/:oncall`, x.Verify(x.OncallShow))
	router.GET(`/oncall/`, x.Verify(x.OncallList))
	router.GET(`/predicate/:predicate`, x.Verify(x.PredicateShow))
	router.GET(`/predicate/`, x.Verify(x.PredicateList))
	router.GET(`/provider/:provider`, x.Verify(x.ProviderShow))
	router.GET(`/provider/`, x.Verify(x.ProviderList))
	router.GET(`/server/:serverID`, x.Verify(x.ServerShow))
	router.GET(`/server/`, x.Verify(x.ServerList))
	router.GET(`/state/:state`, x.Verify(x.StateShow))
	router.GET(`/state/`, x.Verify(x.StateList))
	router.GET(`/status/:status`, x.Verify(x.StatusShow))
	router.GET(`/status/`, x.Verify(x.StatusList))
	router.GET(`/sync/datacenter/`, x.Verify(x.DatacenterSync))
	router.GET(`/sync/server/`, x.Verify(x.ServerSync))
	router.GET(`/sync/team/`, x.Verify(x.TeamMgmtSync))
	router.GET(`/sync/user/`, x.Verify(x.UserMgmtSync))
	router.GET(`/team/:teamID`, x.Verify(x.ScopeSelectTeamShow))
	router.GET(`/team/`, x.Verify(x.TeamMgmtList))
	router.GET(`/unit/:unit`, x.Verify(x.UnitShow))
	router.GET(`/unit/`, x.Verify(x.UnitList))
	router.GET(`/user/:user`, x.Verify(x.ScopeSelectUserShow))
	router.GET(`/user/`, x.Verify(x.UserMgmtList))
	router.GET(`/validity/:property`, x.Verify(x.ValidityShow))
	router.GET(`/validity/`, x.Verify(x.ValidityList))
	router.GET(`/view/:view`, x.Verify(x.ViewShow))
	router.GET(`/view/`, x.Verify(x.ViewList))
	router.GET(`/workflow/`, x.Verify(x.WorkflowList))
	router.GET(`/workflow/summary`, x.Verify(x.WorkflowSummary))
	router.GET(rtBucket, x.Verify(x.BucketList))
	router.GET(rtBucketID, x.Verify(x.BucketShow))
	router.GET(rtBucketInstance, x.Verify(x.InstanceList))
	router.GET(rtBucketInstanceID, x.Verify(x.InstanceShow))
	router.GET(rtBucketInstanceVersions, x.Verify(x.InstanceVersions))
	router.GET(rtBucketMember, x.Verify(x.BucketMemberList))
	router.GET(rtBucketTree, x.Verify(x.BucketTree))
	router.GET(rtCluster, x.Verify(x.ClusterList))
	router.GET(rtClusterID, x.Verify(x.ClusterShow))
	router.GET(rtClusterInstance, x.Verify(x.InstanceList))
	router.GET(rtClusterInstanceID, x.Verify(x.InstanceShow))
	router.GET(rtClusterInstanceVersions, x.Verify(x.InstanceVersions))
	router.GET(rtClusterMember, x.Verify(x.ClusterMemberList))
	router.GET(rtClusterTree, x.Verify(x.ClusterTree))
	router.GET(rtGroup, x.Verify(x.GroupList))
	router.GET(rtGroupID, x.Verify(x.GroupShow))
	router.GET(rtGroupInstance, x.Verify(x.InstanceList))
	router.GET(rtGroupInstanceID, x.Verify(x.InstanceShow))
	router.GET(rtGroupInstanceVersions, x.Verify(x.InstanceVersions))
	router.GET(rtGroupMember, x.Verify(x.GroupMemberList))
	router.GET(rtGroupTree, x.Verify(x.GroupTree))
	router.GET(rtNode, x.Verify(x.NodeList))
	router.GET(rtNodeConfig, x.Verify(x.NodeShowConfig))
	router.GET(rtNodeID, x.Verify(x.NodeShow))
	router.GET(rtNodeInstance, x.Verify(x.InstanceList))
	router.GET(rtNodeInstanceID, x.Verify(x.InstanceShow))
	router.GET(rtNodeInstanceVersions, x.Verify(x.InstanceVersions))
	router.GET(rtNodeTree, x.Verify(x.NodeConfigTree))
	router.GET(rtPermission, x.Verify(x.PermissionList))
	router.GET(rtPermissionID, x.Verify(x.PermissionShow))
	router.GET(rtPropertyMgmt, x.Verify(x.PropertyMgmtList))
	router.GET(rtPropertyMgmtID, x.Verify(x.PropertyMgmtShow))
	router.GET(rtRepository, x.Verify(x.RepositoryConfigList))
	router.GET(rtRepositoryID, x.Verify(x.RepositoryConfigShow))
	router.GET(rtRepositoryInstance, x.Verify(x.InstanceList))
	router.GET(rtRepositoryInstanceID, x.Verify(x.InstanceShow))
	router.GET(rtRepositoryInstanceVersions, x.Verify(x.InstanceVersions))
	router.GET(rtRepositoryPropertyMgmt, x.Verify(x.PropertyMgmtList))
	router.GET(rtRepositoryPropertyMgmtID, x.Verify(x.PropertyMgmtShow))
	router.GET(rtRepositoryTree, x.Verify(x.RepositoryConfigTree))
	router.GET(rtRight, x.Verify(x.RightList))
	router.GET(rtRightID, x.Verify(x.RightShow))
	router.GET(rtSyncNode, x.Verify(x.NodeMgmtSync))
	router.GET(rtTeamPropertyMgmt, x.Verify(x.PropertyMgmtList))
	router.GET(rtTeamPropertyMgmtID, x.Verify(x.PropertyMgmtShow))
	router.HEAD(`/authenticate/validate`, x.Verify(x.SupervisorValidate))
	router.POST(`/hostdeployment/:monitoringID/:assetID`, x.CheckShutdown(x.HostDeploymentAssemble))
	router.POST(`/search/action/`, x.Verify(x.ActionSearch))
	router.POST(`/search/capability/`, x.Verify(x.CapabilitySearch))
	router.POST(`/search/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigSearch))
	router.POST(`/search/job/`, x.Verify(x.JobSearch))
	router.POST(`/search/level/`, x.Verify(x.LevelSearch))
	router.POST(`/search/monitoringsystem/`, x.Verify(x.ScopeSelectMonitoringSearch))
	router.POST(`/search/oncall/`, x.Verify(x.OncallSearch))
	router.POST(`/search/section/`, x.Verify(x.SectionSearch))
	router.POST(`/search/server/`, x.Verify(x.ServerSearch))
	router.POST(`/search/team/`, x.Verify(x.ScopeSelectTeamSearch))
	router.POST(`/search/user/`, x.Verify(x.ScopeSelectUserSearch))
	router.POST(`/search/workflow/`, x.Verify(x.WorkflowSearch))
	router.POST(rtSearchBucket, x.Verify(x.BucketSearch))
	router.POST(rtSearchCluster, x.Verify(x.ClusterSearch))
	router.POST(rtSearchCustomProperty, x.Verify(x.PropertyMgmtSearch))
	router.POST(rtSearchGLobalProperty, x.Verify(x.PropertyMgmtSearch))
	router.POST(rtSearchGroup, x.Verify(x.GroupList))
	router.POST(rtSearchNode, x.Verify(x.NodeSearch))
	router.POST(rtSearchPermission, x.Verify(x.PermissionSearch))
	router.POST(rtSearchRepository, x.Verify(x.RepositoryConfigSearch))
	router.POST(rtSearchRight, x.Verify(x.RightSearch))
	router.POST(rtSearchServiceProperty, x.Verify(x.PropertyMgmtSearch))

	if !x.conf.ReadOnly {
		if !x.conf.Observer {
			router.DELETE(`/accounts/tokens/:account`, x.Verify(x.SupervisorTokenInvalidateAccount))
			router.DELETE(`/attribute/:attribute`, x.Verify(x.AttributeRemove))
			router.DELETE(`/capability/:capabilityID`, x.Verify(x.CapabilityRemove))
			router.DELETE(`/category/:category/section/:sectionID/action/:actionID`, x.Verify(x.ActionRemove))
			router.DELETE(`/category/:category/section/:sectionID`, x.Verify(x.SectionRemove))
			router.DELETE(`/category/:category`, x.Verify(x.CategoryRemove))
			router.DELETE(`/checkconfig/:repositoryID/:checkID`, x.Verify(x.CheckConfigDestroy))
			router.DELETE(`/datacenter/:datacenter`, x.Verify(x.DatacenterRemove))
			router.DELETE(`/entity/:entity`, x.Verify(x.EntityRemove))
			router.DELETE(`/environment/:environment`, x.Verify(x.EnvironmentRemove))
			router.DELETE(`/level/:level`, x.Verify(x.LevelRemove))
			router.DELETE(`/metric/:metric`, x.Verify(x.MetricRemove))
			router.DELETE(`/mode/:mode`, x.Verify(x.ModeRemove))
			router.DELETE(`/monitoringsystem/:monitoring`, x.Verify(x.MonitoringMgmtRemove))
			router.DELETE(`/oncall/:oncall`, x.Verify(x.OncallRemove))
			router.DELETE(`/predicate/:predicate`, x.Verify(x.PredicateRemove))
			router.DELETE(`/provider/:provider`, x.Verify(x.ProviderRemove))
			router.DELETE(`/server/:serverID`, x.Verify(x.ServerRemove))
			router.DELETE(`/state/:state`, x.Verify(x.StateRemove))
			router.DELETE(`/status/:status`, x.Verify(x.StatusRemove))
			router.DELETE(`/team/:teamID`, x.Verify(x.TeamMgmtRemove))
			router.DELETE(`/tokens/global`, x.Verify(x.SupervisorTokenInvalidateGlobal))
			router.DELETE(`/tokens/self/active`, x.Verify(x.SupervisorTokenInvalidate))
			router.DELETE(`/tokens/self/all`, x.Verify(x.SupervisorTokenInvalidateSelf))
			router.DELETE(`/unit/:unit`, x.Verify(x.UnitRemove))
			router.DELETE(`/user/:userID`, x.Verify(x.UserMgmtRemove))
			router.DELETE(`/validity/:property`, x.Verify(x.ValidityRemove))
			router.DELETE(`/view/:view`, x.Verify(x.ViewRemove))
			router.DELETE(rtBucketID, x.Verify(x.BucketDestroy))
			router.DELETE(rtBucketMemberID, x.Verify(x.BucketMemberUnassign))
			router.DELETE(rtBucketPropertyID, x.Verify(x.BucketPropertyDestroy))
			router.DELETE(rtClusterID, x.Verify(x.ClusterDestroy))
			router.DELETE(rtClusterMemberID, x.Verify(x.ClusterMemberUnassign))
			router.DELETE(rtClusterPropertyID, x.Verify(x.ClusterPropertyDestroy))
			router.DELETE(rtGroupID, x.Verify(x.GroupDestroy))
			router.DELETE(rtGroupMemberID, x.Verify(x.GroupMemberUnassign))
			router.DELETE(rtGroupPropertyID, x.Verify(x.GroupPropertyDestroy))
			router.DELETE(rtNode, x.Verify(x.NodeMgmtRemove))
			router.DELETE(rtNodeID, x.Verify(x.NodeMgmtRemove))
			router.DELETE(rtNodePropertyID, x.Verify(x.NodeConfigPropertyDestroy))
			router.DELETE(rtNodeUnassign, x.Verify(x.NodeConfigUnassign))
			router.DELETE(rtPermissionID, x.Verify(x.PermissionRemove))
			router.DELETE(rtPropertyMgmtID, x.Verify(x.PropertyMgmtRemove))
			router.DELETE(rtRepositoryID, x.Verify(x.RepositoryDestroy))
			router.DELETE(rtRepositoryPropertyID, x.Verify(x.RepositoryConfigPropertyDestroy))
			router.DELETE(rtRepositoryPropertyMgmtID, x.Verify(x.PropertyMgmtCustomRemove))
			router.DELETE(rtRightID, x.Verify(x.RightRevoke))
			router.DELETE(rtTeamPropertyMgmtID, x.Verify(x.PropertyMgmtServiceRemove))
			router.GET(rtAliasDeploymentID, x.CheckShutdown(x.DeploymentShow))
			router.GET(rtDeployment, x.CheckShutdown(x.DeploymentList))
			router.GET(rtDeploymentID, x.CheckShutdown(x.DeploymentShow))
			router.GET(rtDeploymentState, x.CheckShutdown(x.DeploymentPending))
			router.GET(rtDeploymentStateID, x.CheckShutdown(x.DeploymentFilter))
			router.GET(rtRepositoryAudit, x.Verify(x.RepositoryAudit))
			router.PATCH(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordChange))
			router.PATCH(`/oncall/:oncall`, x.Verify(x.OncallUpdate))
			router.PATCH(`/workflow/retry`, x.Verify(x.WorkflowRetry))
			router.PATCH(`/workflow/set/:instanceconfigID`, x.Verify(x.WorkflowSet))
			router.PATCH(rtAliasDeploymentIDAction, x.CheckShutdown(x.DeploymentUpdate))
			router.PATCH(rtClusterID, x.Verify(x.ClusterRename))
			router.PATCH(rtDeploymentIDAction, x.CheckShutdown(x.DeploymentUpdate))
			router.PATCH(rtPermissionID, x.Verify(x.PermissionEdit))
			router.POST(`/attribute/`, x.Verify(x.AttributeAdd))
			router.POST(`/capability/`, x.Verify(x.CapabilityAdd))
			router.POST(`/category/:category/section/:sectionID/action/`, x.Verify(x.ActionAdd))
			router.POST(`/category/:category/section/`, x.Verify(x.SectionAdd))
			router.POST(`/category/`, x.Verify(x.CategoryAdd))
			router.POST(`/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigCreate))
			router.POST(`/datacenter/`, x.Verify(x.DatacenterAdd))
			router.POST(`/entity/`, x.Verify(x.EntityAdd))
			router.POST(`/environment/`, x.Verify(x.EnvironmentAdd))
			router.POST(`/kex/`, x.CheckShutdown(x.SupervisorKex))
			router.POST(`/level/`, x.Verify(x.LevelAdd))
			router.POST(`/metric/`, x.Verify(x.MetricAdd))
			router.POST(`/mode/`, x.Verify(x.ModeAdd))
			router.POST(`/monitoringsystem/`, x.Verify(x.MonitoringMgmtAdd))
			router.POST(`/oncall/`, x.Verify(x.OncallAdd))
			router.POST(`/predicate/`, x.Verify(x.PredicateAdd))
			router.POST(`/provider/`, x.Verify(x.ProviderAdd))
			router.POST(`/server/:serverID`, x.Verify(x.ServerAddNull))
			router.POST(`/server/`, x.Verify(x.ServerAdd))
			router.POST(`/state/`, x.Verify(x.StateAdd))
			router.POST(`/status/`, x.Verify(x.StatusAdd))
			router.POST(`/system/`, x.Verify(x.SystemOperation))
			router.POST(`/team/`, x.Verify(x.TeamMgmtAdd))
			router.POST(`/unit/`, x.Verify(x.UnitAdd))
			router.POST(`/user/`, x.Verify(x.UserMgmtAdd))
			router.POST(`/validity/`, x.Verify(x.ValidityAdd))
			router.POST(`/view/`, x.Verify(x.ViewAdd))
			router.POST(rtBucket, x.Verify(x.BucketCreate))
			router.POST(rtBucketMember, x.Verify(x.BucketMemberAssign))
			router.POST(rtBucketProperty, x.Verify(x.BucketPropertyCreate))
			router.POST(rtCluster, x.Verify(x.ClusterCreate))
			router.POST(rtClusterMember, x.Verify(x.ClusterMemberAssign))
			router.POST(rtClusterProperty, x.Verify(x.ClusterPropertyCreate))
			router.POST(rtGroup, x.Verify(x.GroupCreate))
			router.POST(rtGroupMember, x.Verify(x.GroupMemberAssign))
			router.POST(rtGroupProperty, x.Verify(x.GroupPropertyCreate))
			router.POST(rtNode, x.Verify(x.NodeMgmtAdd))
			router.POST(rtNodeProperty, x.Verify(x.NodeConfigPropertyCreate))
			router.POST(rtPermission, x.Verify(x.PermissionAdd))
			router.POST(rtPropertyMgmt, x.Verify(x.PropertyMgmtAdd))
			router.POST(rtRepository, x.Verify(x.RepositoryMgmtCreate))
			router.POST(rtRepositoryProperty, x.Verify(x.RepositoryConfigPropertyCreate))
			router.POST(rtRepositoryPropertyMgmt, x.Verify(x.PropertyMgmtCustomAdd))
			router.POST(rtRight, x.Verify(x.RightGrant))
			router.POST(rtTeamPropertyMgmt, x.Verify(x.PropertyMgmtServiceAdd))
			router.PUT(`/accounts/activate/root/:kexID`, x.CheckShutdown(x.SupervisorActivateRoot))
			router.PUT(`/accounts/activate/user/:kexID`, x.CheckShutdown(x.SupervisorActivateUser))
			router.PUT(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordReset))
			router.PUT(`/datacenter/:datacenter`, x.Verify(x.DatacenterRename))
			router.PUT(`/entity/:entity`, x.Verify(x.EntityRename))
			router.PUT(`/environment/:environment`, x.Verify(x.EnvironmentRename))
			router.PUT(`/job/:jobID`, x.Verify(x.ScopeSelectJobWait))
			router.PUT(`/server/:serverID`, x.Verify(x.ServerUpdate))
			router.PUT(`/state/:state`, x.Verify(x.StateRename))
			router.PUT(`/team/:teamID`, x.Verify(x.TeamMgmtUpdate))
			router.PUT(`/tokens/request/:kexID`, x.CheckShutdown(x.SupervisorTokenRequest))
			router.PUT(`/user/:userID`, x.Verify(x.UserMgmtUpdate))
			router.PUT(`/view/:view`, x.Verify(x.ViewRename))
			router.PUT(rtBucketPropertyID, x.Verify(x.BucketPropertyUpdate))
			router.PUT(rtClusterPropertyID, x.Verify(x.ClusterPropertyUpdate))
			router.PUT(rtGroupPropertyID, x.Verify(x.GroupPropertyUpdate))
			router.PUT(rtNodeConfig, x.Verify(x.NodeConfigAssign))
			router.PUT(rtNodeID, x.Verify(x.NodeMgmtUpdate))
			router.PUT(rtNodePropertyID, x.Verify(x.NodeConfigPropertyUpdate))
			router.PUT(rtRepositoryPropertyID, x.Verify(x.RepositoryConfigPropertyUpdate))
		}
	}
	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
