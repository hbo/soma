package somatree

import "github.com/satori/go.uuid"

// Implementation of the `Checker` interface

//
// Checker:> Add Check

func (ten *SomaTreeElemNode) SetCheck(c Check) {
	c.Id = c.GetItemId(ten.Type, ten.Id)
	if uuid.Equal(c.Id, uuid.Nil) {
		c.Id = uuid.NewV4()
	}
	// this check is the source check
	c.InheritedFrom = ten.Id
	c.Inherited = false
	c.SourceId, _ = uuid.FromString(c.Id.String())
	c.SourceType = ten.Type
	// scrub checkitem startup information prior to storing
	c.Items = nil
	ten.addCheck(c)
}
func (ten *SomaTreeElemNode) setCheckInherited(c Check) {
	// we keep a local copy, that way we know it is ours....
	f := c.clone()
	f.Id = f.GetItemId(ten.Type, ten.Id)
	if uuid.Equal(f.Id, uuid.Nil) {
		f.Id = uuid.NewV4()
	}
	f.Items = nil
	ten.addCheck(f)
}

func (ten *SomaTreeElemNode) setCheckOnChildren(c Check) {
}

func (ten *SomaTreeElemNode) addCheck(c Check) {
	ten.Checks[c.Id.String()] = c
	ten.actionCheckNew(ten.setupCheckAction(c))
}

//
// Checker:> Remove Check

func (ten *SomaTreeElemNode) DeleteCheck(c Check) {
	ten.rmCheck(c)
}

func (ten *SomaTreeElemNode) deleteCheckInherited(c Check) {
	ten.rmCheck(c)
}

func (ten *SomaTreeElemNode) deleteCheckOnChildren(c Check) {
}

func (ten *SomaTreeElemNode) rmCheck(c Check) {
	for id, _ := range ten.Checks {
		if uuid.Equal(ten.Checks[id].SourceId, c.SourceId) {
			ten.actionCheckRemoved(ten.setupCheckAction(ten.Checks[id]))
			delete(ten.Checks, id)
			return
		}
	}
}

// noop, satisfy interface
func (ten *SomaTreeElemNode) syncCheck(childId string) {
}

func (ten *SomaTreeElemNode) checkCheck(checkId string) bool {
	if _, ok := ten.Checks[checkId]; ok {
		return true
	}
	return false
}

//
func (ten *SomaTreeElemNode) LoadInstance(i CheckInstance) {
	ckId := i.CheckId.String()
	ckInstId := i.InstanceId.String()
	if ten.loadedInstances[ckId] == nil {
		ten.loadedInstances[ckId] = map[string]CheckInstance{}
	}
	ten.loadedInstances[ckId][ckInstId] = i
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
