package logistics

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	handler := NewHandler(NewService(queries))
	mux.HandleFunc("GET /api/cargo", handler.List)
	mux.HandleFunc("POST /api/cargo", handler.CreateCargo)
	mux.HandleFunc("PUT /api/cargo/{id}/logistics-batch", handler.AssignCargo)
	mux.HandleFunc("GET /api/cargo/qr/{id}", handler.GetByQR)
	mux.HandleFunc("POST /api/cargo/qr/{id}/scan", handler.RecordQRScan)
	mux.HandleFunc("POST /api/logistics-batches", handler.CreateBatch)
}
