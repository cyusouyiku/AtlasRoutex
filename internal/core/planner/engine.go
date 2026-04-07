package planner

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Run(candidateIDs []string, dayCount int) [][]string {
if dayCount <= 0 {
dayCount = 1
}
out := make([][]string, dayCount)
for i, id := range candidateIDs {
out[i%dayCount] = append(out[i%dayCount], id)
}
return out
}
