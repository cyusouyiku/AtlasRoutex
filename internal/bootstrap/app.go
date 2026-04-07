package bootstrap

import (
	"net/http"

	"atlas-routex/internal/api/handler"
	"atlas-routex/internal/api/handler/middleware"
)

type App struct {
	Config Config
	DB     *DatabaseSet
}

func NewApp(cfg Config) *App {
	db, err := InitDatabases()
	if err != nil {
		panic(err)
	}
	return &App{Config: cfg, DB: db}
}

func (a *App) BuildMux() http.Handler {
	plan := handler.NewPlanHandler(a.DB.PlanUsecase, a.DB.AdjustUsecase)
	poi := handler.NewPOIHandler(a.DB.RecommendUsecase)
	user := handler.NewUserHandler(a.DB.Users)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/plan", plan.Plan)
	mux.HandleFunc("/api/v1/plan/adjust", plan.Adjust)
	mux.HandleFunc("/api/v1/pois", poi.List)
	mux.HandleFunc("/api/v1/user/profile", user.Profile)

	var h http.Handler = mux
	h = middleware.Recovery(h)
	h = middleware.RateLimit(h)
	h = middleware.Logger(h)
	h = middleware.Auth(h)
	return h
}
