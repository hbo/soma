/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/mjolnir42/soma/lib/proto"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Group struct {
	ID              uuid.UUID
	Name            string
	State           string
	Team            uuid.UUID
	Type            string
	Parent          GroupReceiver `json:"-"`
	Fault           *Fault        `json:"-"`
	Action          chan *Action  `json:"-"`
	PropertyOncall  map[string]Property
	PropertyService map[string]Property
	PropertySystem  map[string]Property
	PropertyCustom  map[string]Property
	Checks          map[string]Check
	CheckInstances  map[string][]string
	Instances       map[string]CheckInstance
	Children        map[string]GroupAttacher `json:"-"`
	loadedInstances map[string]map[string]CheckInstance
	ordNumChildGrp  int
	ordNumChildClr  int
	ordNumChildNod  int
	ordChildrenGrp  map[int]string
	ordChildrenClr  map[int]string
	ordChildrenNod  map[int]string
	hasUpdate       bool
	log             *log.Logger
	lock            *sync.RWMutex
}

type GroupSpec struct {
	ID   string
	Name string
	Team string
}

//
// NEW
func NewGroup(spec GroupSpec) *Group {
	if !specGroupCheck(spec) {
		fmt.Printf("%#v\n", spec) // XXX DEBUG
		panic(`No.`)
	}

	teg := new(Group)
	teg.lock = &sync.RWMutex{}
	teg.ID, _ = uuid.FromString(spec.ID)
	teg.Name = spec.Name
	teg.Team, _ = uuid.FromString(spec.Team)
	teg.Type = "group"
	teg.State = "floating"
	teg.Parent = nil
	teg.Children = make(map[string]GroupAttacher)
	teg.PropertyOncall = make(map[string]Property)
	teg.PropertyService = make(map[string]Property)
	teg.PropertySystem = make(map[string]Property)
	teg.PropertyCustom = make(map[string]Property)
	teg.Checks = make(map[string]Check)
	teg.CheckInstances = make(map[string][]string)
	teg.Instances = make(map[string]CheckInstance)
	teg.loadedInstances = make(map[string]map[string]CheckInstance)
	teg.ordNumChildGrp = 0
	teg.ordNumChildClr = 0
	teg.ordNumChildNod = 0
	teg.ordChildrenGrp = make(map[int]string)
	teg.ordChildrenClr = make(map[int]string)
	teg.ordChildrenNod = make(map[int]string)

	return teg
}

func (teg Group) Clone() *Group {
	teg.lock.RLock()
	defer teg.lock.RUnlock()
	cl := Group{
		Name:           teg.Name,
		State:          teg.State,
		Type:           teg.Type,
		ordNumChildGrp: teg.ordNumChildGrp,
		ordNumChildClr: teg.ordNumChildClr,
		ordNumChildNod: teg.ordNumChildNod,
		log:            teg.log,
		lock:           &sync.RWMutex{},
	}
	cl.ID, _ = uuid.FromString(teg.ID.String())
	cl.Team, _ = uuid.FromString(teg.Team.String())

	f := make(map[string]GroupAttacher, 0)
	for k, child := range teg.Children {
		f[k] = child.CloneGroup()
	}
	cl.Children = f

	pO := make(map[string]Property)
	for k, prop := range teg.PropertyOncall {
		pO[k] = prop.Clone()
	}
	cl.PropertyOncall = pO

	pSv := make(map[string]Property)
	for k, prop := range teg.PropertyService {
		pSv[k] = prop.Clone()
	}
	cl.PropertyService = pSv

	pSy := make(map[string]Property)
	for k, prop := range teg.PropertySystem {
		pSy[k] = prop.Clone()
	}
	cl.PropertySystem = pSy

	pC := make(map[string]Property)
	for k, prop := range teg.PropertyCustom {
		pC[k] = prop.Clone()
	}
	cl.PropertyCustom = pC

	cK := make(map[string]Check)
	for k, chk := range teg.Checks {
		cK[k] = chk.Clone()
	}
	cl.Checks = cK

	cki := make(map[string]CheckInstance)
	for k, chki := range teg.Instances {
		cki[k] = chki.Clone()
	}
	cl.Instances = cki
	cl.loadedInstances = make(map[string]map[string]CheckInstance)

	ci := make(map[string][]string)
	for k := range teg.CheckInstances {
		for _, str := range teg.CheckInstances[k] {
			t := str
			ci[k] = append(ci[k], t)
		}
	}
	cl.CheckInstances = ci

	chLG := make(map[int]string)
	for i, s := range teg.ordChildrenGrp {
		chLG[i] = s
	}
	cl.ordChildrenGrp = chLG

	chLC := make(map[int]string)
	for i, s := range teg.ordChildrenClr {
		chLC[i] = s
	}
	cl.ordChildrenClr = chLC

	chLN := make(map[int]string)
	for i, s := range teg.ordChildrenNod {
		chLN[i] = s
	}
	cl.ordChildrenNod = chLN

	return &cl
}

func (teg Group) CloneBucket() BucketAttacher {
	return teg.Clone()
}

func (teg Group) CloneGroup() GroupAttacher {
	return teg.Clone()
}

//
// Interface: Builder
func (teg *Group) GetID() string {
	return teg.ID.String()
}

func (teg *Group) GetName() string {
	return teg.Name
}

func (teg *Group) GetType() string {
	return teg.Type
}

func (teg *Group) setParent(p Receiver) {
	switch p.(type) {
	case *Bucket:
		teg.setGroupParent(p.(GroupReceiver))
		teg.State = "standalone"
	case *Group:
		teg.setGroupParent(p.(GroupReceiver))
		teg.State = "grouped"
	default:
		fmt.Printf("Type: %s\n", reflect.TypeOf(p))
		panic(`Group.setParent`)
	}
}

func (teg *Group) setAction(c chan *Action) {
	teg.Action = c
}

func (teg *Group) setActionDeep(c chan *Action) {
	teg.setAction(c)
	for ch := range teg.Children {
		teg.Children[ch].setActionDeep(c)
	}
}

func (teg *Group) setLog(newlog *log.Logger) {
	teg.log = newlog
}

func (teg *Group) setLoggerDeep(newlog *log.Logger) {
	teg.setLog(newlog)
	for ch := range teg.Children {
		teg.Children[ch].setLoggerDeep(newlog)
	}
}

// GroupReceiver == can receive Groups as children
func (teg *Group) setGroupParent(p GroupReceiver) {
	teg.Parent = p
}

func (teg *Group) updateParentRecursive(p Receiver) {
	teg.setParent(p)
	var wg sync.WaitGroup
	for child := range teg.Children {
		wg.Add(1)
		c := child
		go func(str Receiver) {
			defer wg.Done()
			teg.Children[c].updateParentRecursive(str)
		}(teg)
	}
	wg.Wait()
}

func (teg *Group) clearParent() {
	teg.Parent = nil
	teg.State = "floating"
}

func (teg *Group) setFault(f *Fault) {
	teg.Fault = f
}

func (teg *Group) updateFaultRecursive(f *Fault) {
	teg.setFault(f)
	var wg sync.WaitGroup
	for child := range teg.Children {
		wg.Add(1)
		c := child
		go func(ptr *Fault) {
			defer wg.Done()
			teg.Children[c].updateFaultRecursive(ptr)
		}(f)
	}
	wg.Wait()
}

//
// Interface: Bucketeer
func (teg *Group) GetBucket() Receiver {
	if teg.Parent == nil {
		if teg.Fault == nil {
			panic(`Group.GetBucket called without Parent`)
		} else {
			return teg.Fault
		}
	}
	return teg.Parent.(Bucketeer).GetBucket()
}

func (teg *Group) GetRepository() string {
	return teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetRepository()
}

func (teg *Group) GetRepositoryName() string {
	return teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetRepositoryName()
}

func (teg *Group) GetEnvironment() string {
	return teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetEnvironment()
}

//
//
func (teg *Group) ComputeCheckInstances() {
	teg.log.Printf("TK[%s]: Action=%s, ObjectType=%s, ObjectID=%s",
		teg.GetRepositoryName(),
		`ComputeCheckInstances`,
		`group`,
		teg.ID.String(),
	)
	var wg sync.WaitGroup
	switch deterministicInheritanceOrder {
	case true:
		// groups
		for i := 0; i < teg.ordNumChildGrp; i++ {
			if child, ok := teg.ordChildrenGrp[i]; ok {
				teg.Children[child].ComputeCheckInstances()
			}
		}
		// clusters
		for i := 0; i < teg.ordNumChildClr; i++ {
			if child, ok := teg.ordChildrenClr[i]; ok {
				teg.Children[child].ComputeCheckInstances()
			}
		}
		// nodes
		for i := 0; i < teg.ordNumChildNod; i++ {
			if child, ok := teg.ordChildrenNod[i]; ok {
				teg.Children[child].ComputeCheckInstances()
			}
		}
	default:
		for child := range teg.Children {
			wg.Add(1)
			go func(ch string) {
				defer wg.Done()
				teg.Children[ch].ComputeCheckInstances()
			}(child)
		}
	}
	teg.updateCheckInstances()
	wg.Wait()
}

//
//
func (teg *Group) ClearLoadInfo() {
	var wg sync.WaitGroup
	for child := range teg.Children {
		wg.Add(1)
		c := child
		go func() {
			defer wg.Done()
			teg.Children[c].ClearLoadInfo()
		}()
	}
	wg.Wait()
	teg.loadedInstances = map[string]map[string]CheckInstance{}
}

//
//
func (teg *Group) export() proto.Group {
	bucket := teg.Parent.(Bucketeer).GetBucket()
	return proto.Group{
		ID:          teg.ID.String(),
		Name:        teg.Name,
		BucketID:    bucket.(Builder).GetID(),
		ObjectState: teg.State,
		TeamID:      teg.Team.String(),
	}
}

func (teg *Group) actionCreate() {
	teg.Action <- &Action{
		Action: ActionCreate,
		Type:   teg.Type,
		Group:  teg.export(),
	}
}

func (teg *Group) actionUpdate() {
	teg.Action <- &Action{
		Action: ActionUpdate,
		Type:   teg.Type,
		Group:  teg.export(),
	}
}

func (teg *Group) actionDelete() {
	teg.Action <- &Action{
		Action: ActionDelete,
		Type:   teg.Type,
		Group:  teg.export(),
	}
}

func (teg *Group) actionRename() {
	teg.Action <- &Action{
		Action: ActionRename,
		Type:   teg.Type,
		Group:  teg.export(),
	}
}

func (teg *Group) actionRepossess() {
	teg.Action <- &Action{
		Action: ActionRepossess,
		Type:   teg.Type,
		Group:  teg.export(),
	}
}

func (teg *Group) actionMemberNew(a Action) {
	a.Action = ActionMemberNew
	a.Type = teg.Type
	a.Group = teg.export()

	teg.Action <- &a
}

func (teg *Group) actionMemberRemoved(a Action) {
	a.Action = ActionMemberRemoved
	a.Type = teg.Type
	a.Group = teg.export()

	teg.Action <- &a
}

//
func (teg *Group) actionPropertyNew(a Action) {
	a.Action = ActionPropertyNew
	teg.actionProperty(a)
}

func (teg *Group) actionPropertyUpdate(a Action) {
	a.Action = ActionPropertyUpdate
	teg.actionProperty(a)
}

func (teg *Group) actionPropertyDelete(a Action) {
	a.Action = ActionPropertyDelete
	teg.actionProperty(a)
}

func (teg *Group) actionProperty(a Action) {
	a.Type = teg.Type
	a.Group = teg.export()
	a.Property.RepositoryID = teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetRepository()
	a.Property.BucketID = teg.Parent.(Bucketeer).GetBucket().(Builder).GetID()

	switch a.Property.Type {
	case "custom":
		a.Property.Custom.RepositoryID = a.Property.RepositoryID
	case "service":
		a.Property.Service.TeamID = teg.Team.String()
	}

	teg.Action <- &a
}

//
func (teg *Group) actionCheckNew(a Action) {
	a.Check.RepositoryID = teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetRepository()
	a.Check.BucketID = teg.Parent.(Bucketeer).GetBucket().(Builder).GetID()
	teg.actionDispatch(ActionCheckNew, a)
}

func (teg *Group) actionCheckRemoved(a Action) {
	a.Check.RepositoryID = teg.Parent.(Bucketeer).GetBucket().(Bucketeer).GetRepository()
	a.Check.BucketID = teg.Parent.(Bucketeer).GetBucket().(Builder).GetID()
	teg.actionDispatch(ActionCheckRemoved, a)
}

func (teg *Group) setupCheckAction(c Check) Action {
	return c.MakeAction()
}

func (teg *Group) actionCheckInstanceCreate(a Action) {
	teg.actionDispatch(ActionCheckInstanceCreate, a)
}

func (teg *Group) actionCheckInstanceUpdate(a Action) {
	teg.actionDispatch(ActionCheckInstanceUpdate, a)
}

func (teg *Group) actionCheckInstanceDelete(a Action) {
	teg.actionDispatch(ActionCheckInstanceDelete, a)
}

func (teg *Group) actionDispatch(action string, a Action) {
	a.Action = action
	a.Type = teg.Type
	a.Group = teg.export()

	teg.Action <- &a
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
