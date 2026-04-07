package handler

import (
	"net/http"

	apidto "atlas-routex/internal/api/dto"
	"atlas-routex/internal/domain/repository"
)

type UserHandler struct {
	users repository.UserRepository
}

func NewUserHandler(users repository.UserRepository) *UserHandler { return &UserHandler{users: users} }

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		writeJSON(w, http.StatusInternalServerError, apidto.Fail("INTERNAL", "user repository is not configured"))
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "demo-user-1"
	}
	user, err := h.users.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apidto.Fail("INTERNAL", err.Error()))
		return
	}
	if user == nil {
		writeJSON(w, http.StatusNotFound, apidto.Fail("NOT_FOUND", "user not found"))
		return
	}
	writeJSON(w, http.StatusOK, apidto.Success(user))
}
