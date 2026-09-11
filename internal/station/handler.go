package station

import "net/http"

// Handler translates station HTTP requests to service calls.
type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "station endpoint is not implemented", http.StatusNotImplemented)
}
