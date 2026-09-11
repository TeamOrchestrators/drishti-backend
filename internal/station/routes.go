package station

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	mux.HandleFunc("GET /api/stations", NewHandler(NewService(queries)).List)
}
