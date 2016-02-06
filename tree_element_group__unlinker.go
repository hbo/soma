package somatree

//
// Interface: SomaTreeUnlinker
func (teg *SomaTreeElemGroup) Unlink(u UnlinkRequest) {
	if unlinkRequestCheck(u, teg) {
		switch u.ChildType {
		case "group":
			teg.unlinkGroup(u)
		case "cluster":
			teg.unlinkCluster(u)
		case "node":
			teg.unlinkNode(u)
		default:
			panic(`SomaTreeElemGroup.Unlink`)
		}
		return
	}
loop:
	for child, _ := range teg.Children {
		if teg.Children[child].(SomaTreeBuilder).GetType() == "node" {
			continue loop
		}
		teg.Children[child].(SomaTreeUnlinker).Unlink(u)
	}
}

//
// Interface: SomaTreeGroupUnlinker
func (teg *SomaTreeElemGroup) unlinkGroup(u UnlinkRequest) {
	if unlinkRequestCheck(u, teg) {
		switch u.ChildType {
		case "group":
			if _, ok := teg.Children[u.ChildId]; ok {
				if u.ChildName == teg.Children[u.ChildId].GetName() {
					teg.Children[u.ChildId].clearParent()
					delete(teg.Children, u.ChildId)

					teg.Action <- &Action{
						Action:    "member_removed",
						Type:      "group",
						Id:        teg.Id.String(),
						Name:      teg.Name,
						Team:      teg.Team.String(),
						ChildType: "group",
						ChildId:   u.ChildId,
					}
				}
			}
		default:
			panic(`SomaTreeElemGroup.unlinkGroup`)
		}
		return
	}
	panic(`SomaTreeElemGroup.unlinkGroup`)
}

//
// Interface: SomaTreeClusterUnlinker
func (teg *SomaTreeElemGroup) unlinkCluster(u UnlinkRequest) {
	if unlinkRequestCheck(u, teg) {
		switch u.ChildType {
		case "cluster":
			if _, ok := teg.Children[u.ChildId]; ok {
				if u.ChildName == teg.Children[u.ChildId].GetName() {
					teg.Children[u.ChildId].clearParent()
					delete(teg.Children, u.ChildId)

					teg.Action <- &Action{
						Action:    "member_removed",
						Type:      "group",
						Id:        teg.Id.String(),
						Name:      teg.Name,
						Team:      teg.Team.String(),
						ChildType: "cluster",
						ChildId:   u.ChildId,
					}
				}
			}
		default:
			panic(`SomaTreeElemGroup.unlinkCluster`)
		}
		return
	}
	panic(`SomaTreeElemGroup.unlinkCluster`)
}

//
// Interface: SomaTreeNodeUnlinker
func (teg *SomaTreeElemGroup) unlinkNode(u UnlinkRequest) {
	if unlinkRequestCheck(u, teg) {
		switch u.ChildType {
		case "node":
			if _, ok := teg.Children[u.ChildId]; ok {
				if u.ChildName == teg.Children[u.ChildId].GetName() {
					teg.Children[u.ChildId].clearParent()
					delete(teg.Children, u.ChildId)

					teg.Action <- &Action{
						Action:    "member_removed",
						Type:      "group",
						Id:        teg.Id.String(),
						Name:      teg.Name,
						Team:      teg.Team.String(),
						ChildType: "node",
						ChildId:   u.ChildId,
					}
				}
			}
		default:
			panic(`SomaTreeElemGroup.unlinkNode`)
		}
		return
	}
	panic(`SomaTreeElemGroup.unlinkNode`)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
