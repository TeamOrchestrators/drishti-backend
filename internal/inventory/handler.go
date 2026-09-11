package inventory

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	stationID, ok := pathUUID(w, r, "stationID")
	if !ok {
		return
	}
	response, err := h.service.List(r.Context(), stationID)
	if err != nil {
		http.Error(w, "could not load station inventory", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) ListForQuery(w http.ResponseWriter, r *http.Request) {
	stationID, ok := queryUUID(w, r, "station_id")
	if !ok {
		return
	}
	response, err := h.service.List(r.Context(), stationID)
	if err != nil {
		http.Error(w, "could not load station inventory", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	stationID, ok := pathUUID(w, r, "stationID")
	if !ok {
		return
	}
	var request CreateItemRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.CreateItem(r.Context(), stationID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) CreateItemForQuery(w http.ResponseWriter, r *http.Request) {
	stationID, ok := queryUUID(w, r, "station_id")
	if !ok {
		return
	}
	var request CreateItemRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.CreateItem(r.Context(), stationID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	stationID, ok := pathUUID(w, r, "stationID")
	if !ok {
		return
	}
	inventoryID, ok := pathUUID(w, r, "inventoryID")
	if !ok {
		return
	}
	var request StockAdjustmentRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.AdjustStock(r.Context(), stationID, inventoryID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(key))
	if err != nil {
		http.Error(w, "invalid "+key, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return value, true
}

func queryUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.URL.Query().Get(key))
	if err != nil {
		http.Error(w, "invalid "+key, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return value, true
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
