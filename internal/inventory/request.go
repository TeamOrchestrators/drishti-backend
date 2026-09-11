package inventory

type CreateItemRequest struct {
	ItemCode        string   `json:"item_code"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Unit            string   `json:"unit"`
	IsCritical      bool     `json:"is_critical"`
	OpeningQuantity float64  `json:"opening_quantity"`
	MinimumQuantity float64  `json:"minimum_quantity"`
	MaximumQuantity *float64 `json:"maximum_quantity,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

type StockAdjustmentRequest struct {
	Operation string  `json:"operation"`
	Quantity  float64 `json:"quantity"`
	Notes     string  `json:"notes,omitempty"`
}
