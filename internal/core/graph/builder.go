package graph

func (g *Graph) AddNode(n Node) {
g.Nodes[n.ID] = n
}

func (g *Graph) AddEdge(e Edge) {
g.Edges[e.From] = append(g.Edges[e.From], e)
}
