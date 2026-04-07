package solver

import (
	"context"
	"time"
)

// Constraint 表示求解时使用的一条约束描述。
type Constraint struct {
	Name  string
	Hard  bool
	Value any
}

// Problem 是求解器的统一输入。
type Problem struct {
	CandidateIDs []string
	StartDate    time.Time
	EndDate      time.Time
	DayCount     int
	Budget       float64
	Constraints  []Constraint
}

// Solution 是求解器返回的最小结果抽象。
type Solution interface {
	Score() float64
	Constraints() []Constraint
}

// Solver 定义统一求解端口。
type Solver interface {
	Solve(ctx context.Context, problem Problem) (Solution, error)
	Name() string
}

// BasicSolution 是一个轻量实现，便于当前项目先编译通过。
type BasicSolution struct {
	Value            float64
	AppliedRules     []Constraint
	DailyAssignments [][]string
}

func (s BasicSolution) Score() float64 {
	return s.Value
}

func (s BasicSolution) Constraints() []Constraint {
	return s.AppliedRules
}
