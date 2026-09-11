package emergency

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// Handler translates emergency HTTP requests to service calls.
type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "could not list emergencies", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) ListActive(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "could not list active emergencies", http.StatusInternalServerError)
		return
	}
	active := make([]EmergencyResponse, 0, len(response.Emergencies))
	for _, emergency := range response.Emergencies {
		if emergency.Status == "active" || emergency.Status == "acknowledged" || emergency.Status == "responding" {
			active = append(active, emergency)
		}
	}
	response.Emergencies = active
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) DeviceFormOptions(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.DeviceFormOptions(r.Context())
	if err != nil {
		http.Error(w, "could not load device form options", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var request RegisterDeviceRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.RegisterDevice(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := pathUUID(w, r, "deviceID")
	if !ok {
		return
	}
	var request HeartbeatRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.Heartbeat(r.Context(), deviceID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
func (h *Handler) InitiateSOS(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := pathUUID(w, r, "deviceID")
	if !ok {
		return
	}
	var request InitiateSOSRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.InitiateSOS(r.Context(), deviceID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
func (h *Handler) ConfirmSOS(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := pathUUID(w, r, "deviceID")
	if !ok {
		return
	}
	confirmationID, ok := pathUUID(w, r, "confirmationID")
	if !ok {
		return
	}
	var request ConfirmSOSRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.ConfirmSOS(r.Context(), deviceID, confirmationID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	emergencyID, ok := pathUUID(w, r, "emergencyID")
	if !ok {
		return
	}
	var request UpdateStatusRequest
	if !decode(w, r, &request) {
		return
	}
	response, err := h.service.UpdateStatus(r.Context(), emergencyID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	emergencyID, ok := pathUUID(w, r, "emergencyID")
	if !ok {
		return
	}
	response, err := h.service.UpdateStatus(r.Context(), emergencyID, UpdateStatusRequest{Status: "resolved"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(key))
	if err != nil {
		http.Error(w, "invalid "+key, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return value, true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
