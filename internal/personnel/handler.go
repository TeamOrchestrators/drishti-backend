package personnel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Handler translates personnel HTTP requests to service calls.
type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "could not list personnel", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) AssignmentFormOptions(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.AssignmentFormOptions(r.Context())
	if err != nil {
		http.Error(w, "could not load personnel assignment form options", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) AddToExpedition(w http.ResponseWriter, r *http.Request) {
	var request AddPersonnelRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.AddToExpedition(r.Context(), request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "assignment is not eligible: check expedition dates, personnel availability, origin station, and active movements", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
