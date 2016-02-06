package somatree

//
// Interface: SomaTreeReceiver
func (tec *SomaTreeElemCluster) Receive(r ReceiveRequest) {
	if receiveRequestCheck(r, tec) {
		switch r.ChildType {
		case "node":
			tec.receiveNode(r)
		default:
			panic(`SomaTreeElemCluster.Receive`)
		}
	}
	// no passing along since only nodes are a SomeTreeClusterAttacher
	// and nodes can have no children
	return
}

//
// Interface: SomaTreeNodeReceiver
func (tec *SomaTreeElemCluster) receiveNode(r ReceiveRequest) {
	if receiveRequestCheck(r, tec) {
		switch r.ChildType {
		case "node":
			tec.Children[r.Node.GetID()] = r.Node
			r.Node.setParent(tec)
			r.Node.setAction(tec.Action)
			r.Node.setFault(tec.Fault)

			tec.Action <- &Action{
				Action:    "member_new",
				Type:      "cluster",
				Id:        tec.Id.String(),
				Name:      tec.Name,
				Team:      tec.Team.String(),
				ChildType: "node",
				ChildId:   r.Node.GetID(),
			}
		default:
			panic(`SomaTreeElemCluster.receiveNode`)
		}
		return
	}
	panic(`SomaTreeElemCluster.receiveNode`)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
