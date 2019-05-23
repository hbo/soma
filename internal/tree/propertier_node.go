/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

import (
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/satori/go.uuid"
)

// Implementation of the `Propertier` interface

//
// Propertier:> Add Property

func (ten *Node) SetProperty(p Property) {
	// if deleteOK is true, then prop is the property that can be
	// deleted
	if dupe, deleteOK, prop := ten.checkDuplicate(p); dupe && !deleteOK {
		ten.Fault.Error <- &Error{Action: `duplicate_set_property`}
		return
	} else if dupe && deleteOK {
		srcUUID, _ := uuid.FromString(prop.GetSourceInstance())
		switch prop.GetType() {
		case `custom`:
			cstUUID, _ := uuid.FromString(prop.GetKey())
			ten.deletePropertyInherited(&PropertyCustom{
				SourceID:  srcUUID,
				View:      prop.GetView(),
				Inherited: true,
				CustomID:  cstUUID,
				Key:       prop.(*PropertyCustom).GetKeyField(),
				Value:     prop.(*PropertyCustom).GetValueField(),
			})
		case `service`:
			// GetValue for serviceproperty returns the uuid to never
			// match, we do not set it
			ten.deletePropertyInherited(&PropertyService{
				SourceID:    srcUUID,
				View:        prop.GetView(),
				Inherited:   true,
				ServiceID:   uuid.Must(uuid.FromString(prop.GetKey())),
				ServiceName: prop.GetValue(),
			})
		case `system`:
			ten.deletePropertyInherited(&PropertySystem{
				SourceID:  srcUUID,
				View:      prop.GetView(),
				Inherited: true,
				Key:       prop.GetKey(),
				Value:     prop.GetValue(),
			})
		case `oncall`:
			// GetValue for oncallproperty returns the uuid to never
			// match, we do not set it
			oncUUID, _ := uuid.FromString(prop.GetKey())
			ten.deletePropertyInherited(&PropertyOncall{
				SourceID:  srcUUID,
				View:      prop.GetView(),
				Inherited: true,
				OncallID:  oncUUID,
				Name:      prop.(*PropertyOncall).GetName(),
				Number:    prop.(*PropertyOncall).GetNumber(),
			})
		}
	}
	p.SetID(p.GetInstanceID(ten.Type, ten.ID, ten.log))
	if p.Equal(uuid.Nil) {
		p.SetID(uuid.Must(uuid.NewV4()))
	}
	// this property is the source instance
	p.SetInheritedFrom(ten.ID)
	p.SetInherited(false)
	p.SetSourceType(ten.Type)
	if i, e := uuid.FromString(p.GetID()); e == nil {
		p.SetSourceID(i)
	}
	// send a scrubbed copy down
	f := p.Clone()
	f.SetInherited(true)
	f.SetID(uuid.UUID{})
	if f.hasInheritance() {
		ten.setPropertyOnChildren(f)
	}
	// scrub instance startup information prior to storing
	p.clearInstances()
	ten.addProperty(p)
	ten.actionPropertyNew(p.MakeAction())
}

func (ten *Node) setPropertyInherited(p Property) {
	f := p.Clone()
	f.SetID(f.GetInstanceID(ten.Type, ten.ID, ten.log))
	if f.Equal(uuid.Nil) {
		f.SetID(uuid.Must(uuid.NewV4()))
	}
	f.clearInstances()

	if !f.GetIsInherited() {
		ten.Fault.Error <- &Error{
			Action: `node.setPropertyInherited on inherited=false`}
		return
	}
	if dupe, deleteOK, _ := ten.checkDuplicate(p); dupe && deleteOK {
		// we received an inherited SetProperty from above us in the
		// tree for a property that is duplicate, but we are not the
		// source of the duplicate -> corrupt tree
		ten.Fault.Error <- &Error{
			Action: `node.setPropertyInherited corruption detected`}
		return
	} else if dupe && !deleteOK {
		// we received an inherited SetProperty from above us in the
		// tree for a property that is duplicate; we have a locally
		// set property -> stop inheritance, no error
		return
	}
	ten.addProperty(f)
	// no inheritPropertyDeep(), nodes have no children
	ten.actionPropertyNew(f.MakeAction())
}

func (ten *Node) setPropertyOnChildren(p Property) {
	// noop, satisfy interface
}

func (ten *Node) addProperty(p Property) {
	ten.hasUpdate = true
	switch p.GetType() {
	case `custom`:
		ten.PropertyCustom[p.GetID()] = p
	case `system`:
		ten.PropertySystem[p.GetID()] = p
	case `service`:
		ten.PropertyService[p.GetID()] = p
	case `oncall`:
		ten.PropertyOncall[p.GetID()] = p
	default:
		ten.hasUpdate = false
		ten.Fault.Error <- &Error{Action: `node.addProperty unknown type`}
	}
}

//
// Propertier:> Update Property

func (ten *Node) UpdateProperty(p Property) {
	if !ten.verifySourceInstance(
		p.GetSourceInstance(),
		p.GetType(),
	) {
		ten.Fault.Error <- &Error{Action: `update_property_on_non_source`}
		return
	}

	// keep a copy for ourselves, no shared pointers
	p.SetInheritedFrom(ten.ID)
	p.SetSourceType(ten.Type)
	p.SetInherited(true)
	f := p.Clone()
	f.SetInherited(false)
	if ten.switchProperty(f) {
		ten.updatePropertyOnChildren(p)
	}
}

func (ten *Node) updatePropertyInherited(p Property) {
	// keep a copy for ourselves, no shared pointers
	f := p.Clone()
	if !f.GetIsInherited() {
		ten.Fault.Error <- &Error{
			Action: `node.updatePropertyInherited on inherited=false`}
		return
	}
	if ten.switchProperty(f) {
		ten.updatePropertyOnChildren(p)
	}
}

func (ten *Node) updatePropertyOnChildren(p Property) {
	// noop, satisfy interface
}

func (ten *Node) switchProperty(p Property) bool {
	uid := ten.findIDForSource(
		p.GetSourceInstance(),
		p.GetType(),
	)
	if uid == `` {
		// we do not have the property for which we received an update
		if dupe, deleteOK, _ := ten.checkDuplicate(p); dupe && !deleteOK {
			// the update is duplicate to an property for which we
			// have the source instance, ie we just received an update
			// for which we have an overwrite. Ignore it and do not
			// inherit it further down
			return false
		}
		ten.Fault.Error <- &Error{
			Action: `node.switchProperty property not found`}
		return false
	}
	updID, _ := uuid.FromString(uid)
	p.SetID(updID)
	ten.addProperty(p)
	ten.actionPropertyUpdate(p.MakeAction())
	// nodes have no children, we require no handling of changes in
	// inheritance here
	return true
}

func (ten *Node) getCurrentProperty(p Property) Property {
	// noop, satisfy interface
	return nil
}

//
// Propertier:> Delete Property

func (ten *Node) DeleteProperty(p Property) {
	if !ten.verifySourceInstance(
		p.GetSourceInstance(),
		p.GetType(),
	) {
		ten.Fault.Error <- &Error{Action: `node.DeleteProperty on !source`}
		return
	}

	var flow Property
	resync := false
	delID := ten.findIDForSource(
		p.GetSourceInstance(),
		p.GetType(),
	)
	if delID != `` {
		// this is a delete for a locally set property. It might be a
		// delete for an overwrite property, in which case we need to
		// ask the parent to sync it to us again.
		// If it was an overwrite, the parent should have a property
		// we would consider a dupe if it were to be passed down to
		// us.
		// If p is considered a dupe, then flow is set to the prop we
		// need to inherit.
		var delProp Property
		switch p.GetType() {
		case `custom`:
			delProp = ten.PropertyCustom[delID]
		case `system`:
			delProp = ten.PropertySystem[delID]
		case `service`:
			delProp = ten.PropertyService[delID]
		case `oncall`:
			delProp = ten.PropertyOncall[delID]
		}
		resync, _, flow = ten.Parent.(Propertier).checkDuplicate(
			delProp,
		)
	}

	p.SetInherited(false)
	if ten.rmProperty(p) {
		p.SetInherited(true)
		ten.deletePropertyOnChildren(p)
	}

	// now that the property is deleted from us and our children,
	// request resync if required
	if resync {
		ten.Parent.(Propertier).resyncProperty(
			flow.GetSourceInstance(),
			p.GetType(),
			ten.ID.String(),
		)
	}
}

func (ten *Node) deletePropertyInherited(p Property) {
	if ten.rmProperty(p) {
		ten.deletePropertyOnChildren(p)
	}
}

func (ten *Node) deletePropertyOnChildren(p Property) {
	// noop, satisfy interface
}

func (ten *Node) deletePropertyAllInherited() {
	ten.lock.Lock()
	defer ten.lock.Unlock()
	for _, p := range ten.PropertyCustom {
		if !p.GetIsInherited() {
			continue
		}
		ten.deletePropertyInherited(p.Clone())
	}
	for _, p := range ten.PropertySystem {
		if !p.GetIsInherited() {
			continue
		}
		ten.deletePropertyInherited(p.Clone())
	}
	for _, p := range ten.PropertyService {
		if !p.GetIsInherited() {
			continue
		}
		ten.deletePropertyInherited(p.Clone())
	}
	for _, p := range ten.PropertyOncall {
		if !p.GetIsInherited() {
			continue
		}
		ten.deletePropertyInherited(p.Clone())
	}
}

func (ten *Node) deletePropertyAllLocal() {
	ten.lock.Lock()
	defer ten.lock.Unlock()
	for _, p := range ten.PropertyCustom {
		if p.GetIsInherited() {
			continue
		}
		ten.DeleteProperty(p.Clone())
	}
	for _, p := range ten.PropertySystem {
		if p.GetIsInherited() {
			continue
		}
		ten.DeleteProperty(p.Clone())
	}
	for _, p := range ten.PropertyService {
		if p.GetIsInherited() {
			continue
		}
		ten.DeleteProperty(p.Clone())
	}
	for _, p := range ten.PropertyOncall {
		if p.GetIsInherited() {
			continue
		}
		ten.DeleteProperty(p.Clone())
	}
}

func (ten *Node) rmProperty(p Property) bool {
	delID := ten.findIDForSource(
		p.GetSourceInstance(),
		p.GetType(),
	)
	if delID == `` {
		// we do not have the property for which we received a delete
		if dupe, deleteOK, _ := ten.checkDuplicate(p); dupe && !deleteOK {
			// the delete is duplicate to a property for which we
			// have the source instance, ie we just received a delete
			// for which we have an overwrite. Ignore it and do not
			// inherit it further down
			return false
		}

		ten.Fault.Error <- &Error{
			Action: `node.rmProperty property not found`}
		return false
	}

	hasInheritance := false
	ten.hasUpdate = true
	switch p.GetType() {
	case `custom`:
		ten.actionPropertyDelete(
			ten.PropertyCustom[delID].MakeAction(),
		)
		hasInheritance = ten.PropertyCustom[delID].hasInheritance()
		delete(ten.PropertyCustom, delID)
	case `service`:
		ten.actionPropertyDelete(
			ten.PropertyService[delID].MakeAction(),
		)
		hasInheritance = ten.PropertyService[delID].hasInheritance()
		delete(ten.PropertyService, delID)
	case `system`:
		ten.actionPropertyDelete(
			ten.PropertySystem[delID].MakeAction(),
		)
		hasInheritance = ten.PropertySystem[delID].hasInheritance()
		delete(ten.PropertySystem, delID)
	case `oncall`:
		ten.actionPropertyDelete(
			ten.PropertyOncall[delID].MakeAction(),
		)
		hasInheritance = ten.PropertyOncall[delID].hasInheritance()
		delete(ten.PropertyOncall, delID)
	default:
		ten.hasUpdate = false
		ten.Fault.Error <- &Error{Action: `node.rmProperty unknown type`}
		return false
	}
	return hasInheritance
}

//
// Propertier:> Utility

//
func (ten *Node) verifySourceInstance(id, prop string) bool {
	switch prop {
	case `custom`:
		if _, ok := ten.PropertyCustom[id]; !ok {
			goto bailout
		}
		return ten.PropertyCustom[id].GetSourceInstance() == id
	case `service`:
		if _, ok := ten.PropertyService[id]; !ok {
			goto bailout
		}
		return ten.PropertyService[id].GetSourceInstance() == id
	case `system`:
		if _, ok := ten.PropertySystem[id]; !ok {
			goto bailout
		}
		return ten.PropertySystem[id].GetSourceInstance() == id
	case `oncall`:
		if _, ok := ten.PropertyOncall[id]; !ok {
			goto bailout
		}
		return ten.PropertyOncall[id].GetSourceInstance() == id
	}

bailout:
	ten.Fault.Error <- &Error{
		Action: `node.verifySourceInstance not found`}
	return false
}

func (ten *Node) findIDForSource(source, prop string) string {
	switch prop {
	case `custom`:
		for id := range ten.PropertyCustom {
			if ten.PropertyCustom[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `system`:
		for id := range ten.PropertySystem {
			if ten.PropertySystem[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `service`:
		for id := range ten.PropertyService {
			if ten.PropertyService[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `oncall`:
		for id := range ten.PropertyOncall {
			if ten.PropertyOncall[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	}
	return ``
}

func (ten *Node) syncProperty(childID string) {
	// noop, satisfy interface
}

func (ten *Node) checkProperty(propType string, propID string) bool {
	// noop, satisfy interface
	return false
}

// Checks if this property is already defined on this node, and
// whether it was inherited, ie. can be deleted so it can be
// overwritten
func (ten *Node) checkDuplicate(p Property) (bool, bool, Property) {
	var dupe, deleteOK bool
	var prop Property

propswitch:
	switch p.GetType() {
	case "custom":
		for _, pVal := range ten.PropertyCustom {
			dupe, deleteOK, prop = isDupe(pVal, p)
			if dupe {
				break propswitch
			}
		}
	case "service":
		for _, pVal := range ten.PropertyService {
			dupe, deleteOK, prop = isDupe(pVal, p)
			if dupe {
				break propswitch
			}
		}
	case "oncall":
		for _, pVal := range ten.PropertyOncall {
			dupe, deleteOK, prop = isDupe(pVal, p)
			if dupe {
				break propswitch
			}
		}
	case msg.PropertySystem:
		for _, pVal := range ten.PropertySystem {
			switch p.GetKey() {
			case msg.SystemPropertyTag:
				// tags are only dupes if the value is the same as well
				fallthrough
			case msg.SystemPropertyDisableCheckConfiguration:
				// disable_check_configuration checks values as well
				if p.GetValue() == pVal.GetValue() {
					// same value, can be a dupe
					dupe, deleteOK, prop = isDupe(pVal, p)
					if dupe {
						break propswitch
					}
				}
			default:
				dupe, deleteOK, prop = isDupe(pVal, p)
				if dupe {
					break propswitch
				}
			}
		}
	default:
		// trigger error path
		ten.Fault.Error <- &Error{Action: `node.checkDuplicate unknown type`}
		dupe = true
		deleteOK = false
	}
	return dupe, deleteOK, prop
}

func (ten *Node) resyncProperty(srcID, pType, childID string) {
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
