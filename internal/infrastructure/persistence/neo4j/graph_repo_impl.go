package neo4j

import "context"

type GraphRepo struct{}

func NewGraphRepo() *GraphRepo { return &GraphRepo{} }

func (r *GraphRepo) UpsertEdge(_ context.Context, from, to string, weight float64) error {
return nil
}
