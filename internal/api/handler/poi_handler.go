package handler

import (
	"net/http"
	"strconv"
	"strings"

	apidto "atlas-routex/internal/api/dto"
	apprecommender "atlas-routex/internal/application/recommender"
)

type POIHandler struct {
	recommender *apprecommender.RecommendUsecase
}

func NewPOIHandler(recommender *apprecommender.RecommendUsecase) *POIHandler {
	return &POIHandler{recommender: recommender}
}

func (h *POIHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.recommender == nil {
		writeJSON(w, http.StatusInternalServerError, apidto.Fail("INTERNAL", "recommender is not configured"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Tokyo"
	}
	var tags []string
	if raw := strings.TrimSpace(r.URL.Query().Get("tags")); raw != "" {
		tags = strings.Split(raw, ",")
	}
	input := &apprecommender.RecommendInput{City: city, Tags: tags, Limit: limit}
	pois, err := h.recommender.Execute(r.Context(), input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apidto.Fail("RECOMMEND_FAILED", err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, apidto.Success(pois))
}
