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
	router.GET(`/bucket/:bucket`, x.Verify(x.BucketShow))
	router.GET(`/bucket/`, x.Verify(x.BucketList))
	router.GET(`/capability/:capabilityID`, x.Verify(x.CapabilityShow))
	router.GET(`/capability/`, x.Verify(x.CapabilityList))
	router.GET(`/category/:category`, x.Verify(x.CategoryShow))
	router.GET(`/category/`, x.Verify(x.CategoryList))
	router.GET(`/checkconfig/:repositoryID/:checkID`, x.Verify(x.CheckConfigShow))
	router.GET(`/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigList))
	router.GET(`/cluster/:clusterID/members/`, x.Verify(x.ClusterMemberList))
	router.GET(`/cluster/:clusterID`, x.Verify(x.ClusterShow))
	router.GET(`/cluster/`, x.Verify(x.ClusterList))
	router.GET(`/datacenter/:datacenter`, x.Verify(x.DatacenterShow))
	router.GET(`/datacenter/`, x.Verify(x.DatacenterList))
	router.GET(`/entity/:entity`, x.Verify(x.EntityShow))
	router.GET(`/entity/`, x.Verify(x.EntityList))
	router.GET(`/environment/:environment`, x.Verify(x.EnvironmentShow))
	router.GET(`/environment/`, x.Verify(x.EnvironmentList))
	router.GET(`/level/:level`, x.Verify(x.LevelShow))
	router.GET(`/level/`, x.Verify(x.LevelList))
	router.GET(`/section/:section/actions/:action`, x.Verify(x.ActionShow))
	router.GET(`/section/:section/actions/`, x.Verify(x.ActionList))
	router.GET(`/section/:section`, x.Verify(x.SectionShow))
	router.GET(`/section/`, x.Verify(x.SectionList))
	router.GET(`/state/:state`, x.Verify(x.StateShow))
	router.GET(`/state/`, x.Verify(x.StateList))
	router.GET(`/sync/datacenter/`, x.Verify(x.DatacenterSync))
	router.GET(`/sync/node/`, x.Verify(x.NodeMgmtSync))
	router.GET(`/unit/:unit`, x.Verify(x.UnitShow))
	router.GET(`/unit/`, x.Verify(x.UnitList))
	router.GET(`/validity/:property`, x.Verify(x.ValidityShow))
	router.GET(`/validity/`, x.Verify(x.ValidityList))
	router.GET(`/view/:view`, x.Verify(x.ViewShow))
	router.GET(`/view/`, x.Verify(x.ViewList))
	router.HEAD(`/authenticate/validate`, x.Verify(x.SupervisorValidate))
	router.POST(`/filter/actions/`, x.Verify(x.ActionSearch))
	router.POST(`/search/bucket/`, x.Verify(x.BucketSearch))
	router.POST(`/search/capability/`, x.Verify(x.CapabilitySearch))
	router.POST(`/search/checkconfig/:repositoryID/`, x.Verify(x.CheckConfigSearch))
	router.POST(`/search/cluster/`, x.Verify(x.ClusterSearch))
	router.POST(`/search/level/`, x.Verify(x.LevelSearch))
	router.POST(`/search/section/`, x.Verify(x.SectionSearch))

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
			router.DELETE(`/node/:nodeID`, x.Verify(x.NodeMgmtRemove))
			router.DELETE(`/section/:section/actions/:action`, x.Verify(x.ActionRemove))
			router.DELETE(`/section/:section`, x.Verify(x.SectionRemove))
			router.DELETE(`/state/:state`, x.Verify(x.StateRemove))
			router.DELETE(`/tokens/global`, x.Verify(x.SupervisorTokenInvalidateGlobal))
			router.DELETE(`/tokens/self/active`, x.Verify(x.SupervisorTokenInvalidate))
			router.DELETE(`/tokens/self/all`, x.Verify(x.SupervisorTokenInvalidateSelf))
			router.DELETE(`/unit/:unit`, x.Verify(x.UnitRemove))
			router.DELETE(`/validity/:property`, x.Verify(x.ValidityRemove))
			router.DELETE(`/view/:view`, x.Verify(x.ViewRemove))
			router.PATCH(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordChange))
			router.PATCH(`/view/:view`, x.Verify(x.ViewRename))
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
			router.POST(`/node/`, x.Verify(x.NodeMgmtAdd))
			router.POST(`/section/:section/actions/`, x.Verify(x.ActionAdd))
			router.POST(`/section/`, x.Verify(x.SectionAdd))
			router.POST(`/state/`, x.Verify(x.StateAdd))
			router.POST(`/unit/`, x.Verify(x.UnitAdd))
			router.POST(`/validity/`, x.Verify(x.ValidityAdd))
			router.POST(`/view/`, x.Verify(x.ViewAdd))
			router.PUT(`/accounts/activate/root/:kexID`, x.CheckShutdown(x.SupervisorActivateRoot))
			router.PUT(`/accounts/activate/user/:kexID`, x.CheckShutdown(x.SupervisorActivateUser))
			router.PUT(`/accounts/password/:kexID`, x.CheckShutdown(x.SupervisorPasswordReset))
			router.PUT(`/datacenter/:datacenter`, x.Verify(x.DatacenterRename))
			router.PUT(`/entity/:entity`, x.Verify(x.EntityRename))
			router.PUT(`/environment/:environment`, x.Verify(x.EnvironmentRename))
			router.PUT(`/node/:nodeID`, x.Verify(x.NodeMgmtUpdate))
			router.PUT(`/state/:state`, x.Verify(x.StateRename))
			router.PUT(`/tokens/request/:kexID`, x.CheckShutdown(x.SupervisorTokenRequest))
		}
	}
	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix