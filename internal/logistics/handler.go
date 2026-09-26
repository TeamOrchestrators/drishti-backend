package logistics

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "could not list logistics", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var request CreateBatchRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.CreateBatch(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) CreateCargo(w http.ResponseWriter, r *http.Request) {
	var request CreateCargoRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.CreateCargo(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) AssignCargo(w http.ResponseWriter, r *http.Request) {
	cargoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid cargo id", http.StatusBadRequest)
		return
	}
	var request AssignCargoBatchRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.AssignCargo(r.Context(), cargoID, request.LogisticsBatchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetByQR(w http.ResponseWriter, r *http.Request) {
	cargoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid cargo QR token", http.StatusBadRequest)
		return
	}
	response, err := h.service.GetByQR(r.Context(), cargoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "cargo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not retrieve cargo QR details", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) RecordQRScan(w http.ResponseWriter, r *http.Request) {
	cargoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid cargo QR token", http.StatusBadRequest)
		return
	}
	var request RecordCargoQRScanRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.RecordQRScan(r.Context(), cargoID, request)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "cargo not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) GetBatchTracking(w http.ResponseWriter, r *http.Request) {
	batchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid logistics batch id", http.StatusBadRequest)
		return
	}
	response, err := h.service.GetBatchTracking(r.Context(), batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "logistics batch not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not retrieve logistics batch tracking", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
