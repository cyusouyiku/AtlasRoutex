package graph

func (g *Graph) Neighbors(id string) []Edge {
return g.Edges[id]
}
