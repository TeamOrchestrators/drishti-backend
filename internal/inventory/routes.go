package inventory

import (
	"net/http"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
)

func RegisterRoutes(mux *http.ServeMux, queries *db.Queries) {
	handler := NewHandler(NewService(queries))
	mux.HandleFunc("GET /api/inventory", handler.ListForQuery)
	mux.HandleFunc("POST /api/inventory/items", handler.CreateItemForQuery)
	mux.HandleFunc("GET /api/stations/{stationID}/inventory", handler.List)
	mux.HandleFunc("POST /api/stations/{stationID}/inventory/items", handler.CreateItem)
	mux.HandleFunc("POST /api/stations/{stationID}/inventory/{inventoryID}/stock", handler.AdjustStock)
}
