package station

// CreateRequest is the input shape for a future station-creation endpoint.
type CreateRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location string `json:"location"`
}
