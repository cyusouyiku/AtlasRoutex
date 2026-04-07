package handler

import (
	"encoding/json"
	"net/http"

	apidto "atlas-routex/internal/api/dto"
	appplanner "atlas-routex/internal/application/planner"
)

type PlanHandler struct {
	planner *appplanner.PlanUsecase
	adjust  *appplanner.AdjustUsecase
}

func NewPlanHandler(planner *appplanner.PlanUsecase, adjust *appplanner.AdjustUsecase) *PlanHandler {
	return &PlanHandler{planner: planner, adjust: adjust}
}

func (h *PlanHandler) Plan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apidto.Fail("METHOD_NOT_ALLOWED", "method not allowed"))
		return
	}
	if h.planner == nil {
		writeJSON(w, http.StatusInternalServerError, apidto.Fail("INTERNAL", "planner usecase is not configured"))
		return
	}

	var req appplanner.PlanCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("BAD_REQUEST", "invalid json body"))
		return
	}
	in, err := req.ToPlanInput()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("INVALID_ARGUMENT", err.Error()))
		return
	}
	it, err := h.planner.Execute(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("PLAN_FAILED", err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, apidto.Success(appplanner.ItineraryToSummaryDTO(it)))
}

func (h *PlanHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apidto.Fail("METHOD_NOT_ALLOWED", "method not allowed"))
		return
	}
	if h.adjust == nil {
		writeJSON(w, http.StatusInternalServerError, apidto.Fail("INTERNAL", "adjust usecase is not configured"))
		return
	}

	var req appplanner.AdjustRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("BAD_REQUEST", "invalid json body"))
		return
	}
	in, err := req.ToAdjustInput()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("INVALID_ARGUMENT", err.Error()))
		return
	}
	it, err := h.adjust.Execute(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("ADJUST_FAILED", err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, apidto.Success(appplanner.ItineraryToSummaryDTO(it)))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
