/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

import (
	"strings"
	"sync"

	uuid "github.com/satori/go.uuid"
)

//
// Interface: Attacher
func (ter *Repository) Attach(a AttachRequest) {
	if ter.Parent != nil {
		panic(`Repository.Attach: already attached`)
	}
	switch {
	case a.ParentType == "root":
		ter.attachToRoot(a)
	default:
		panic(`Repository.Attach`)
	}

	if ter.Parent == nil {
		panic(`Repository.Attach: failed`)
	}
	// no need to sync properties, as top level element the repo can't
	// inherit
}

func (ter *Repository) Destroy() {
	if ter.Parent == nil {
		panic(`Repository.Destroy called without Parent to unlink from`)
	}
	// call before unlink since it requires tec.Parent.*
	ter.actionDelete()
	ter.deletePropertyAllLocal()
	ter.deletePropertyAllInherited()
	ter.deleteCheckLocalAll()

	clone := ter.Clone()
	for c := range clone.Children {
		ter.Children[c].Destroy()
	}

	// the Destroy handler of Fault calls
	// updateFaultRecursive(nil) on us
	ter.Fault.Destroy()

	ter.Parent.Unlink(UnlinkRequest{
		ParentType: ter.Parent.(Builder).GetType(),
		ParentID:   ter.Parent.(Builder).GetID(),
		ParentName: ter.Parent.(Builder).GetName(),
		ChildType:  ter.GetType(),
		ChildName:  ter.GetName(),
		ChildID:    ter.GetID(),
	},
	)
	ter.setAction(nil)
}

func (ter *Repository) Detach() {
	ter.Destroy()
}

func (ter *Repository) SetName(newRepoName string) {
	ter.lock.RLock()
	defer ter.lock.RUnlock()
	for i := range ter.Children {
		newBucketName := strings.Replace(
			ter.Children[i].GetName(),
			ter.Name,
			newRepoName,
			1,
		)
		ter.Children[i].SetName(newBucketName)
	}

	ter.Name = newRepoName
	ter.actionRename()
}

func (ter *Repository) SetTeamID(newTeamID string) {
	wg := sync.WaitGroup{}
	ter.lock.RLock()
	defer ter.lock.RUnlock()
	switch deterministicInheritanceOrder {
	case true:
		// buckets
		for i := 0; i < ter.ordNumChildBck; i++ {
			if child, ok := ter.ordChildrenBck[i]; ok {
				ter.Children[child].inheritTeamID(newTeamID)
			}
		}
	default:
		for child := range ter.Children {
			wg.Add(1)
			go func(name, teamID string) {
				defer wg.Done()
				ter.Children[name].inheritTeamID(teamID)
			}(child, newTeamID)
		}
	}
	ter.Team, _ = uuid.FromString(newTeamID)
	ter.actionRepossess()
	wg.Wait()
}

func (ter *Repository) inheritTeamID(newTeamUUID string) {
	ter.Fault.Error <- &Error{
		Action: ActionRepossess,
		Text:   `Repository received inheritTeamID() invocation`,
	}
}

// Interface: RootAttacher
func (ter *Repository) attachToRoot(a AttachRequest) {
	a.Root.Receive(ReceiveRequest{
		ParentType: a.ParentType,
		ParentID:   a.ParentID,
		ParentName: a.ParentName,
		ChildType:  "repository",
		Repository: ter,
	})

	ter.actionCreate()
	ter.Fault.setAction(ter.Action)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
