package worker

type Graph struct {
	roots []*Plugin
}

func (pg *Plugin) Traverse(i int) *Plugin {
	// Traverse to i-th node
	return pg.children[i]
}

func (pg *Plugin) Back(i int) *Plugin {
	// Traverse to i-th node
	return pg.parent[i]
}

func GetEmptyGraph() Graph {
	return Graph{}
}

func (g *Graph) AddRoot(rootNode *Plugin) {
	g.roots = append(g.roots, rootNode)
}

func (g *Graph) GetRoots() []*Plugin {
	return g.roots
}

func (g *Graph) SetCallback(w *Worker, curNode *Plugin) {
	if curNode == nil {
		for _, root := range g.roots {
			g.SetCallback(w, root)
		}
	} else {
		if len(curNode.children) != 0 {
			for _, child := range curNode.children {
				g.SetCallback(w, child)
			}
		}
		curNode.SetCallback(w)
	}
}

func AddPath(parent *Plugin, children []*Plugin) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}
