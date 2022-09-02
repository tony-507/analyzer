package worker

type GraphNode struct {
	val         *Plugin
	m_parameter interface{}
	children    []*GraphNode
	parent      []*GraphNode
}

type Graph struct {
	roots []*GraphNode
}

func CreateNode(val *Plugin, m_parameter interface{}) GraphNode {
	node := GraphNode{val: val, children: make([]*GraphNode, 0), m_parameter: m_parameter}
	node.parent = nil
	return node
}

func (node *GraphNode) GetVal() *Plugin {
	return node.val
}

func (node *GraphNode) Traverse(i int) *GraphNode {
	// Traverse to i-th node
	return node.children[i]
}

func (node *GraphNode) Back(i int) *GraphNode {
	// Traverse to i-th node
	return node.parent[i]
}

func GetEmptyGraph() Graph {
	return Graph{}
}

func (g *Graph) AddRoot(rootNode *GraphNode) {
	g.roots = append(g.roots, rootNode)
}

func (g *Graph) GetRoots() []*GraphNode {
	return g.roots
}

func (g *Graph) SetCallback(w *Worker, curNode *GraphNode) {
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
		curNode.GetVal().SetCallback(w)
	}
}

func AddPath(parent *GraphNode, children []*GraphNode) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}
