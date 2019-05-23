/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

import (
	log "github.com/sirupsen/logrus"
	uuid "github.com/satori/go.uuid"
)

// deterministicInheritanceOrder switches off concurrency in some tree
// actions. This is activated in unit testing to create repeatable test
// runs
var deterministicInheritanceOrder = false

type Tree struct {
	ID      uuid.UUID
	Name    string
	Type    string
	Child   *Repository
	Snap    *Repository
	Action  chan *Action `json:"-"`
	log     *log.Logger
	errChan chan *Error
}

type Spec struct {
	ID     string
	Name   string
	Action chan *Action
	Log    *log.Logger
}

func New(spec Spec) *Tree {
	st := new(Tree)
	st.ID, _ = uuid.FromString(spec.ID)
	st.Name = spec.Name
	st.Action = spec.Action
	st.Type = "root"
	st.log = spec.Log
	return st
}

func (st *Tree) Begin() {
	t := st.Child.Clone()
	st.Snap = &t
	st.Snap.updateParentRecursive(st)

	newFault().Attach(
		AttachRequest{
			Root:       st.Snap,
			ParentType: st.Snap.Type,
			ParentName: st.Snap.Name,
		},
	)
}

func (st *Tree) Rollback() {
	err := st.Child.Fault.Error
	ac := st.Child.Action

	st.Child = st.Snap
	st.Snap = nil
	st.Child.setActionDeep(ac)
	st.Child.setError(err)
}

func (st *Tree) Commit() {
	st.Snap = nil
}

func (st *Tree) AttachError(err Error) {
	if st.Child != nil {
		st.Child.Fault.Error <- &err
	}
}

func (st *Tree) SwitchLogger(newlog *log.Logger) {
	st.log = newlog
	if st.Child != nil {
		st.Child.setLoggerDeep(newlog)
	}
}

//
// Interface: Builder
func (st *Tree) GetID() string {
	return st.ID.String()
}

func (st *Tree) GetName() string {
	return st.Name
}

func (st *Tree) GetType() string {
	return st.Type
}

//
func (st *Tree) SetTeamID(newTeamID string) {
	st.Child.SetTeamID(newTeamID)
}

//
func (st *Tree) RegisterErrChan(c chan *Error) {
	st.errChan = c
}

//
func (st *Tree) SetError() {
	if st.Child != nil {
		st.Child.setError(st.errChan)
		return
	}
	go func() {
		st.errChan <- &Error{
			Action: `SetError()`,
			Text:   `tree.SetError called without attached child`,
		}
	}()
}

func (st *Tree) GetErrors() []error {
	if st.Child != nil {
		return st.Child.getErrors()
	}
	return []error{}
}

// Interface: Receiver
func (st *Tree) Receive(r ReceiveRequest) {
	switch {
	case r.ParentType == "root" &&
		r.ParentID == st.ID.String() &&
		r.ChildType == "repository":
		st.receiveRepository(r)
	default:
		if st.Child != nil {
			st.Child.Receive(r)
		} else {
			panic("not allowed")
		}
	}
}

// Interface: Unlinker
func (st *Tree) Unlink(u UnlinkRequest) {
	switch {
	case u.ParentType == "root" &&
		(u.ParentID == st.ID.String() ||
			u.ParentName == st.Name) &&
		u.ChildType == "repository" &&
		u.ChildName == st.Child.GetName():
		st.unlinkRepository(u)
	default:
		if st.Child != nil {
			st.Child.Unlink(u)
		} else {
			panic("not allowed")
		}
	}
}

// Interface: RepositoryReceiver
func (st *Tree) receiveRepository(r ReceiveRequest) {
	switch {
	case r.ParentType == "root" &&
		r.ParentID == st.ID.String() &&
		r.ChildType == "repository":
		st.Child = r.Repository
		r.Repository.setParent(st)
		r.Repository.setAction(st.Action)
		r.Repository.setLog(st.log)
	default:
		panic("not allowed")
	}
}

// Interface: RepositoryUnlinker
func (st *Tree) unlinkRepository(u UnlinkRequest) {
	switch {
	case u.ParentType == "root" &&
		u.ParentID == st.ID.String() &&
		u.ChildType == "repository" &&
		u.ChildName == st.Child.GetName():
		st.Child = nil
	default:
		panic("not allowed")
	}
}

// Interface: Finder
func (st *Tree) Find(f FindRequest, b bool) Attacher {
	if !b {
		panic(`Tree.Find: root element can never inherit a Find request`)
	}

	res := st.Child.Find(f, false)
	if res != nil {
		return res
	}
	return st.Child.Fault
}

//
func (st *Tree) ComputeCheckInstances() {
	if st.Child == nil {
		panic(`Tree.ComputeCheckInstances: no repository registered`)
	}

	st.log.Printf("Tree[%s]: Action=%s, ObjectType=%s, ObjectID=%s",
		st.Name,
		`ComputeCheckInstances`,
		`tree`,
		st.ID.String(),
	)
	st.Child.ComputeCheckInstances()
	return
}

//
func (st *Tree) ClearLoadInfo() {
	if st.Child == nil {
		panic(`Tree.ClearLoadInfo: no repository registered`)
	}

	st.Child.ClearLoadInfo()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
