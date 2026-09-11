package expedition

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequestBody
	if !decodeRequest(w, r, &request) {
		return
	}

	response, err := h.service.Create(r.Context(), request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "could not list expeditions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) FormOptions(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.FormOptions(r.Context())
	if err != nil {
		http.Error(w, "could not load expedition form options", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid expedition id", http.StatusBadRequest)
		return
	}

	var request CreateRequestBody
	if !decodeRequest(w, r, &request) {
		return
	}

	response, err := h.service.Update(r.Context(), id, request)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "expedition not found", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
