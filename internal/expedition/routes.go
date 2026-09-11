package expedition

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	service := NewService(queries)
	handler := NewHandler(service)

	mux.HandleFunc("GET /api/expeditions", handler.List)
	mux.HandleFunc("GET /api/expeditions/form-options", handler.FormOptions)
	mux.HandleFunc("POST /api/expeditions", handler.Create)
	mux.HandleFunc("PUT /api/expeditions/{id}", handler.Update)
}
