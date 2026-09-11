package personnel

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	handler := NewHandler(NewService(queries))
	mux.HandleFunc("GET /api/personnel", handler.List)
	mux.HandleFunc("GET /api/personnel/assignment-form-options", handler.AssignmentFormOptions)
	mux.HandleFunc("POST /api/personnel/expedition-assignment", handler.AddToExpedition)
}
