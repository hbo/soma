package tree

import "github.com/satori/go.uuid"

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
			ten.deletePropertyInherited(&PropertyCustom{
				SourceId: srcUUID,
			})
		case `service`:
			ten.deletePropertyInherited(&PropertyService{
				SourceId: srcUUID,
			})
		case `system`:
			ten.deletePropertyInherited(&PropertySystem{
				SourceId: srcUUID,
			})
		case `oncall`:
			ten.deletePropertyInherited(&PropertyOncall{
				SourceId: srcUUID,
			})
		}
	}
	p.SetId(p.GetInstanceId(ten.Type, ten.Id))
	if p.Equal(uuid.Nil) {
		p.SetId(uuid.NewV4())
	}
	// this property is the source instance
	p.SetInheritedFrom(ten.Id)
	p.SetInherited(false)
	p.SetSourceType(ten.Type)
	if i, e := uuid.FromString(p.GetID()); e == nil {
		p.SetSourceId(i)
	}
	// send a scrubbed copy down
	f := p.Clone()
	f.SetInherited(true)
	f.SetId(uuid.UUID{})
	ten.setPropertyOnChildren(f)
	// scrub instance startup information prior to storing
	p.clearInstances()
	ten.addProperty(p)
	ten.actionPropertyNew(p.MakeAction())
}

func (ten *Node) setPropertyInherited(p Property) {
	f := p.Clone()
	f.SetId(f.GetInstanceId(ten.Type, ten.Id))
	if f.Equal(uuid.Nil) {
		f.SetId(uuid.NewV4())
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
	p.SetInheritedFrom(ten.Id)
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
	uid := ten.findIdForSource(
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
	updId, _ := uuid.FromString(uid)
	p.SetId(updId)
	ten.addProperty(p)
	ten.actionPropertyUpdate(p.MakeAction())
	return true
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

	ten.rmProperty(p)
	ten.deletePropertyOnChildren(p)
}

func (ten *Node) deletePropertyInherited(p Property) {
	ten.rmProperty(p)
	ten.deletePropertyOnChildren(p)
}

func (ten *Node) deletePropertyOnChildren(p Property) {
	// noop, satisfy interface
}

func (ten *Node) rmProperty(p Property) {
	delId := ten.findIdForSource(
		p.GetSourceInstance(),
		p.GetType(),
	)
	if delId == `` {
		ten.Fault.Error <- &Error{
			Action: `node.rmProperty property not found`}
		return
	}

	switch p.GetType() {
	case `custom`:
		ten.actionPropertyDelete(
			ten.PropertyCustom[delId].MakeAction(),
		)
		delete(ten.PropertyCustom, delId)
	case `service`:
		ten.actionPropertyDelete(
			ten.PropertyService[delId].MakeAction(),
		)
		delete(ten.PropertyService, delId)
	case `system`:
		ten.actionPropertyDelete(
			ten.PropertySystem[delId].MakeAction(),
		)
		delete(ten.PropertySystem, delId)
	case `oncall`:
		ten.actionPropertyDelete(
			ten.PropertyOncall[delId].MakeAction(),
		)
		delete(ten.PropertyOncall, delId)
	default:
		ten.Fault.Error <- &Error{Action: `node.rmProperty unknown type`}
	}
}

//
// Propertier:> Utility

//
func (ten *Node) verifySourceInstance(id, prop string) bool {
	switch prop {
	case `custom`:
		if _, ok := ten.PropertyCustom[id]; !ok {
			return false
		}
		return ten.PropertyCustom[id].GetSourceInstance() == id
	case `service`:
		if _, ok := ten.PropertyService[id]; !ok {
			return false
		}
		return ten.PropertyService[id].GetSourceInstance() == id
	case `system`:
		if _, ok := ten.PropertySystem[id]; !ok {
			return false
		}
		return ten.PropertySystem[id].GetSourceInstance() == id
	case `oncall`:
		if _, ok := ten.PropertyOncall[id]; !ok {
			return false
		}
		return ten.PropertyOncall[id].GetSourceInstance() == id
	default:
		return false
	}
}

func (ten *Node) findIdForSource(source, prop string) string {
	switch prop {
	case `custom`:
		for id, _ := range ten.PropertyCustom {
			if ten.PropertyCustom[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `system`:
		for id, _ := range ten.PropertySystem {
			if ten.PropertySystem[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `service`:
		for id, _ := range ten.PropertyService {
			if ten.PropertyService[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	case `oncall`:
		for id, _ := range ten.PropertyOncall {
			if ten.PropertyOncall[id].GetSourceInstance() != source {
				continue
			}
			return id
		}
	}
	return ``
}

func (ten *Node) syncProperty(childId string) {
	// noop, satisfy interface
}

func (ten *Node) checkProperty(propType string, propId string) bool {
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
	case "system":
		for _, pVal := range ten.PropertySystem {
			// tags are only dupes if the value is the same as well
			if p.GetKey() != `tag` {
				dupe, deleteOK, prop = isDupe(pVal, p)
				if dupe {
					break propswitch
				}
			} else if p.GetValue() == pVal.GetValue() {
				// tag and same value, can be a dupe
				dupe, deleteOK, prop = isDupe(pVal, p)
				if dupe {
					break propswitch
				}
			}
			// tag + different value => pass
		}
	default:
		// trigger error path
		dupe = true
		deleteOK = false
	}
	return dupe, deleteOK, prop
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
