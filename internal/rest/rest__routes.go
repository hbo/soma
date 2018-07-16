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

	router.GET(`/attribute/:attribute`, x.Verify(x.AttributeShow))
	router.GET(`/attribute/`, x.Verify(x.AttributeList))
	router.GET(`/bucket/:bucket/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/bucket/:bucket/instance/:instanceID`, x.Verify(x.InstanceShow))
	router.GET(`/bucket/:bucket/instance/`, x.Verify(x.InstanceList))
	router.GET(`/bucket/:bucketID/tree`, x.Verify(x.BucketTree))
	router.GET(`/bucket/:bucket`, x.Verify(x.BucketShow))
	router.GET(`/bucket/`, x.Verify(x.BucketList))
	router.GET(`/capability/:capabilityID`, x.Verify(x.CapabilityShow))
	router.GET(`/capability/`, x.Verify(x.CapabilityList))
	router.GET(`/category/:category`, x.Verify(x.CategoryShow))
	router.GET(`/category/`, x.Verify(x.CategoryList))
	router.GET(`/checkconfig/:repositoryID/:checkID`, x.Verify(x.CheckConfigShow))
	router.GET(`/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigList))
	router.GET(`/cluster/:clusterID/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/cluster/:clusterID/instance/:instanceID`, x.Verify(x.InstanceShow))
	router.GET(`/cluster/:clusterID/instance/`, x.Verify(x.InstanceList))
	router.GET(`/cluster/:clusterID/members/`, x.Verify(x.ClusterMemberList))
	router.GET(`/cluster/:clusterID/tree`, x.Verify(x.ClusterTree))
	router.GET(`/cluster/:clusterID`, x.Verify(x.ClusterShow))
	router.GET(`/cluster/`, x.Verify(x.ClusterList))
	router.GET(`/datacenter/:datacenter`, x.Verify(x.DatacenterShow))
	router.GET(`/datacenter/`, x.Verify(x.DatacenterList))
	router.GET(`/entity/:entity`, x.Verify(x.EntityShow))
	router.GET(`/entity/`, x.Verify(x.EntityList))
	router.GET(`/environment/:environment`, x.Verify(x.EnvironmentShow))
	router.GET(`/environment/`, x.Verify(x.EnvironmentList))
	router.GET(`/group/:group/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/group/:group/instance/:instanceID`, x.Verify(x.InstanceShow))
	router.GET(`/group/:group/instance/`, x.Verify(x.InstanceList))
	router.GET(`/group/:groupID/tree`, x.Verify(x.GroupTree))
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
	router.GET(`/monitoringsystem/:monitoring`, x.Verify(x.MonitoringShow))
	router.GET(`/monitoringsystem/`, x.Verify(x.ScopeSelectMonitoringList))
	router.GET(`/node/:node/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/node/:node/instance/:instanceID`, x.Verify(x.InstanceShow))
	router.GET(`/node/:node/instance/`, x.Verify(x.InstanceList))
	router.GET(`/node/:nodeID/tree`, x.Verify(x.NodeTree))
	router.GET(`/oncall/:oncall`, x.Verify(x.OncallShow))
	router.GET(`/oncall/`, x.Verify(x.OncallList))
	router.GET(`/predicate/:predicate`, x.Verify(x.PredicateShow))
	router.GET(`/predicate/`, x.Verify(x.PredicateList))
	router.GET(`/provider/:provider`, x.Verify(x.ProviderShow))
	router.GET(`/provider/`, x.Verify(x.ProviderList))
	router.GET(`/repository/:repository/instance/:instanceID/versions`, x.Verify(x.InstanceVersions))
	router.GET(`/repository/:repository/instance/:instanceID`, x.Verify(x.InstanceShow))
	router.GET(`/repository/:repository/instance/`, x.Verify(x.InstanceList))
	router.GET(`/repository/:repositoryID/tree`, x.Verify(x.RepositoryConfigTree))
	router.GET(`/repository/:repositoryID`, x.BasicAuth(x.RepositoryConfigShow))
	router.GET(`/repository/`, x.BasicAuth(x.RepositoryConfigList))
	router.GET(`/section/:section/actions/:action`, x.Verify(x.ActionShow))
	router.GET(`/section/:section/actions/`, x.Verify(x.ActionList))
	router.GET(`/section/:section`, x.Verify(x.SectionShow))
	router.GET(`/section/`, x.Verify(x.SectionList))
	router.GET(`/server/:serverID`, x.Verify(x.ServerShow))
	router.GET(`/server/`, x.Verify(x.ServerList))
	router.GET(`/state/:state`, x.Verify(x.StateShow))
	router.GET(`/state/`, x.Verify(x.StateList))
	router.GET(`/status/:status`, x.Verify(x.StatusShow))
	router.GET(`/status/`, x.Verify(x.StatusList))
	router.GET(`/sync/datacenter/`, x.Verify(x.DatacenterSync))
	router.GET(`/sync/node/`, x.Verify(x.NodeMgmtSync))
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
	router.HEAD(`/authenticate/validate`, x.Verify(x.SupervisorValidate))
	router.POST(`/filter/actions/`, x.Verify(x.ActionSearch))
	router.POST(`/hostdeployment/:monitoringID/:assetID`, x.CheckShutdown(x.HostDeploymentAssemble))
	router.POST(`/search/bucket/`, x.Verify(x.BucketSearch))
	router.POST(`/search/capability/`, x.Verify(x.CapabilitySearch))
	router.POST(`/search/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigSearch))
	router.POST(`/search/cluster/`, x.Verify(x.ClusterSearch))
	router.POST(`/search/job/`, x.Verify(x.JobSearch))
	router.POST(`/search/level/`, x.Verify(x.LevelSearch))
	router.POST(`/search/monitoringsystem/`, x.Verify(x.ScopeSelectMonitoringSearch))
	router.POST(`/search/oncall/`, x.Verify(x.OncallSearch))
	router.POST(`/search/repository/`, x.Verify(x.RepositoryConfigSearch))
	router.POST(`/search/section/`, x.Verify(x.SectionSearch))
	router.POST(`/search/server/`, x.Verify(x.ServerSearch))
	router.POST(`/search/team/`, x.Verify(x.ScopeSelectTeamSearch))
	router.POST(`/search/user/`, x.Verify(x.ScopeSelectUserSearch))
	router.POST(`/search/workflow/`, x.Verify(x.WorkflowSearch))

	if !x.conf.ReadOnly {
		if !x.conf.Observer {
			router.DELETE(`/accounts/tokens/:account`, x.Verify(x.SupervisorTokenInvalidateAccount))
			router.DELETE(`/attribute/:attribute`, x.Verify(x.AttributeRemove))
			router.DELETE(`/bucket/:bucket/property/:type/:source`, x.Verify(x.BucketPropertyDestroy))
			router.DELETE(`/capability/:capabilityID`, x.Verify(x.CapabilityRemove))
			router.DELETE(`/category/:category`, x.Verify(x.CategoryRemove))
			router.DELETE(`/checkconfig/:repositoryID/:checkID`, x.Verify(x.CheckConfigDestroy))
			router.DELETE(`/cluster/:clusterID/property/:propertyType/:sourceID`, x.Verify(x.ClusterPropertyDestroy))
			router.DELETE(`/datacenter/:datacenter`, x.Verify(x.DatacenterRemove))
			router.DELETE(`/entity/:entity`, x.Verify(x.EntityRemove))
			router.DELETE(`/environment/:environment`, x.Verify(x.EnvironmentRemove))
			router.DELETE(`/level/:level`, x.Verify(x.LevelRemove))
			router.DELETE(`/metric/:metric`, x.Verify(x.MetricRemove))
			router.DELETE(`/mode/:mode`, x.Verify(x.ModeRemove))
			router.DELETE(`/monitoringsystem/:monitoring`, x.Verify(x.MonitoringMgmtRemove))
			router.DELETE(`/node/:nodeID`, x.Verify(x.NodeMgmtRemove))
			router.DELETE(`/oncall/:oncall`, x.Verify(x.OncallRemove))
			router.DELETE(`/predicate/:predicate`, x.Verify(x.PredicateRemove))
			router.DELETE(`/provider/:provider`, x.Verify(x.ProviderRemove))
			router.DELETE(`/repository/:repositoryID/property/:propertyType/:sourceInstanceID`, x.BasicAuth(x.RepositoryConfigPropertyDestroy))
			router.DELETE(`/repository/:repositoryID`, x.BasicAuth(x.RepositoryDestroy))
			router.DELETE(`/section/:section/actions/:action`, x.Verify(x.ActionRemove))
			router.DELETE(`/section/:section`, x.Verify(x.SectionRemove))
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
			router.GET(`/repository/:repositoryID/audit`, x.BasicAuth(x.RepositoryAudit))
			router.PATCH(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordChange))
			router.PATCH(`/oncall/:oncall`, x.Verify(x.OncallUpdate))
			router.PATCH(`/view/:view`, x.Verify(x.ViewRename))
			router.PATCH(`/workflow/retry`, x.Verify(x.WorkflowRetry))
			router.PATCH(`/workflow/set/:instanceconfigID`, x.Verify(x.WorkflowSet))
			router.POST(`/attribute/`, x.Verify(x.AttributeAdd))
			router.POST(`/bucket/:bucket/property/:type/`, x.Verify(x.BucketPropertyCreate))
			router.POST(`/bucket/`, x.Verify(x.BucketCreate))
			router.POST(`/capability/`, x.Verify(x.CapabilityAdd))
			router.POST(`/category/`, x.Verify(x.CategoryAdd))
			router.POST(`/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigCreate))
			router.POST(`/cluster/:clusterID/members/`, x.Verify(x.ClusterMemberAssign))
			router.POST(`/cluster/:clusterID/property/:propertyType/`, x.Verify(x.ClusterPropertyCreate))
			router.POST(`/cluster/`, x.Verify(x.ClusterCreate))
			router.POST(`/datacenter/`, x.Verify(x.DatacenterAdd))
			router.POST(`/entity/`, x.Verify(x.EntityAdd))
			router.POST(`/environment/`, x.Verify(x.EnvironmentAdd))
			router.POST(`/kex/`, x.CheckShutdown(x.SupervisorKex))
			router.POST(`/level/`, x.Verify(x.LevelAdd))
			router.POST(`/metric/`, x.Verify(x.MetricAdd))
			router.POST(`/mode/`, x.Verify(x.ModeAdd))
			router.POST(`/monitoringsystem/`, x.Verify(x.MonitoringMgmtAdd))
			router.POST(`/node/`, x.Verify(x.NodeMgmtAdd))
			router.POST(`/oncall/`, x.Verify(x.OncallAdd))
			router.POST(`/predicate/`, x.Verify(x.PredicateAdd))
			router.POST(`/provider/`, x.Verify(x.ProviderAdd))
			router.POST(`/repository/:repositoryID/property/:propertyType/`, x.BasicAuth(x.RepositoryConfigPropertyCreate))
			router.POST(`/repository/`, x.BasicAuth(x.RepositoryMgmtCreate))
			router.POST(`/section/:section/actions/`, x.Verify(x.ActionAdd))
			router.POST(`/section/`, x.Verify(x.SectionAdd))
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
			router.PUT(`/accounts/activate/root/:kexID`, x.CheckShutdown(x.SupervisorActivateRoot))
			router.PUT(`/accounts/activate/user/:kexID`, x.CheckShutdown(x.SupervisorActivateUser))
			router.PUT(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordReset))
			router.PUT(`/datacenter/:datacenter`, x.Verify(x.DatacenterRename))
			router.PUT(`/entity/:entity`, x.Verify(x.EntityRename))
			router.PUT(`/environment/:environment`, x.Verify(x.EnvironmentRename))
			router.PUT(`/job/:jobID`, x.Verify(x.ScopeSelectJobWait))
			router.PUT(`/node/:nodeID`, x.Verify(x.NodeMgmtUpdate))
			router.PUT(`/server/:serverID`, x.Verify(x.ServerUpdate))
			router.PUT(`/state/:state`, x.Verify(x.StateRename))
			router.PUT(`/team/:teamID`, x.Verify(x.TeamMgmtUpdate))
			router.PUT(`/tokens/request/:kexID`, x.CheckShutdown(x.SupervisorTokenRequest))
			router.PUT(`/user/:userID`, x.Verify(x.UserMgmtUpdate))
		}
	}
	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
