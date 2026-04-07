package graph

type Node struct {
ID   string
Lat  float64
Lng  float64
Tags []string
}

type Edge struct {
From string
To   string
Cost float64
}

type Graph struct {
Nodes map[string]Node
Edges map[string][]Edge
}

func NewGraph() *Graph {
return &Graph{Nodes: map[string]Node{}, Edges: map[string][]Edge{}}
}
