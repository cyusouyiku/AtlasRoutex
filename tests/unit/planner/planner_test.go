package planner

import (
	"context"
	"testing"
	"time"

	appplanner "atlas-routex/internal/application/planner"
	"atlas-routex/internal/bootstrap"
	"atlas-routex/internal/domain/entity"
)

func TestDefaultPlanningFlowBuildsPlannedItinerary(t *testing.T) {
	db, err := bootstrap.InitDatabases()
	if err != nil {
		t.Fatalf("init databases: %v", err)
	}

	in := &appplanner.PlanInput{
		UserID:           "demo-user-1",
		ItineraryName:    "Planner Test",
		City:             "Tokyo",
		StartDate:        time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
		TotalBudget:      1000,
		Currency:         "CNY",
		Tags:             []string{"culture"},
		MaxCandidatePOIs: 10,
	}

	it, err := db.PlanUsecase.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if it == nil || it.ID == "" {
		t.Fatalf("expected itinerary to be created")
	}
	if it.Status != entity.ItineraryStatusPlanned {
		t.Fatalf("expected status planned, got %s", it.Status)
	}
	if len(it.Days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(it.Days))
	}
}
