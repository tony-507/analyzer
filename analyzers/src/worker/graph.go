package worker

type Graph struct {
	nodes []*Plugin
}

func GetEmptyGraph() Graph {
	return Graph{}
}

func (g *Graph) AddNode(node *Plugin) {
	g.nodes = append(g.nodes, node)
}

func (g *Graph) SetCallback(w *Worker, curNode *Plugin) {
	if curNode == nil {
		for _, root := range g.nodes {
			g.SetCallback(w, root)
		}
	} else {
		if len(curNode.children) != 0 {
			for _, child := range curNode.children {
				g.SetCallback(w, child)
			}
		}
		curNode.setCallback(w.HandleRequests)
	}
}

func AddPath(parent *Plugin, children []*Plugin) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}
