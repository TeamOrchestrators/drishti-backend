package emergency

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	handler := NewHandler(NewService(queries))
	mux.HandleFunc("GET /api/emergencies", handler.List)
	mux.HandleFunc("GET /api/emergencies/active", handler.ListActive)
	mux.HandleFunc("PATCH /api/emergencies/{emergencyID}/status", handler.UpdateStatus)
	mux.HandleFunc("POST /api/emergencies/{emergencyID}/resolve", handler.Resolve)
	mux.HandleFunc("GET /api/emergency/device/simulate/form-options", handler.DeviceFormOptions)
	mux.HandleFunc("POST /api/emergency/device/simulate/register", handler.RegisterDevice)
	mux.HandleFunc("POST /api/emergency/device/simulate/{deviceID}/heartbeat", handler.Heartbeat)
	mux.HandleFunc("POST /api/emergency/device/simulate/{deviceID}/sos/initiate", handler.InitiateSOS)
	mux.HandleFunc("POST /api/emergency/device/simulate/{deviceID}/sos/{confirmationID}/confirm", handler.ConfirmSOS)
}
