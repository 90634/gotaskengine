package gotaskengine

// Node a package of conveyor used in factory.
type Node struct {
	parents []*Node
	child   *Node
	value   Conveyor
}

func (n *Node) findRoots(rC chan *Node) {
	if len(n.parents) == 0 {
		rC <- n
		return
	}

	for _, v := range n.parents {
		v.findRoots(rC)
	}
}

func (n *Node) run() {
	n.value.Run()
	for _, v := range n.parents {
		v.run()
	}
}

func (n *Node) stop() {
	n.value.Stop()
	if n.child != nil {
		n.child.stop()
	}
}

// Graph, factory uses this struct manage conveyors.
type Graph struct {
	roots  []*Node
	leaves []*Node
	vertex []*Node
}

func (g *Graph) addNode(n *Node) {
	for i, v := range g.vertex {
		if v.value == n.value {
			g.vertex[i].parents = append(g.vertex[i].parents, n.parents...)
			g.vertex[i].child = n.child
			return
		}
	}
	g.vertex = append(g.vertex, n)
}

func (g *Graph) _makeLeaves() {
	for _, v := range g.vertex {
		if v.child == nil {
			g.leaves = append(g.leaves, v)
		}
	}
}

func (g *Graph) _makeRoots() {
	tmpC := make(chan *Node, 8)
	go func() {
		for n := range tmpC {
			g.roots = append(g.roots, n) // there are duplicate nodes.
		}
	}()

	for _, v := range g.leaves {
		v.findRoots(tmpC)
	}
	close(tmpC)
}

func (g *Graph) makeIndexes() {
	g._makeLeaves()
	g._makeRoots()
}

func (g *Graph) runFromLeaves() {
	for _, n := range g.leaves {
		n.run()
	}
}
func (g *Graph) stopFromRoot() {
	for _, n := range g.roots {
		n.stop()
	}
}
