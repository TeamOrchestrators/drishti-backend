package station

// Response is the API-safe representation of a station.
type Response struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location string `json:"location"`
}
