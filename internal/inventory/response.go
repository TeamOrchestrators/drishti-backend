package inventory

import "time"

type ItemResponse struct {
	InventoryID       string                 `json:"inventory_id"`
	ItemID            string                 `json:"item_id"`
	ItemCode          string                 `json:"item_code"`
	Name              string                 `json:"name"`
	Category          string                 `json:"category"`
	Unit              string                 `json:"unit"`
	IsCritical        bool                   `json:"is_critical"`
	OnHandQuantity    float64                `json:"on_hand_quantity"`
	ReservedQuantity  float64                `json:"reserved_quantity"`
	AvailableQuantity float64                `json:"available_quantity"`
	MinimumQuantity   float64                `json:"minimum_quantity"`
	MaximumQuantity   *float64               `json:"maximum_quantity,omitempty"`
	StockLevel        string                 `json:"stock_level"`
	Analytics         *ItemAnalyticsResponse `json:"analytics,omitempty"`
}

type InventoryAlertResponse struct {
	InventoryID       string  `json:"inventory_id"`
	StationID         string  `json:"station_id"`
	ItemID            string  `json:"item_id"`
	ItemName          string  `json:"item_name"`
	Status            string  `json:"status"`
	AvailableQuantity float64 `json:"available_quantity"`
	MinimumQuantity   float64 `json:"minimum_quantity"`
	Message           string  `json:"message"`
}

type ItemAnalyticsResponse struct {
	DaysOfStockRemaining    *float64 `json:"days_of_stock_remaining,omitempty"`
	PredictedDepletionDate  *string  `json:"predicted_depletion_date,omitempty"`
	AnomalyDetected         bool     `json:"anomaly_detected"`
	AnomalyMessage          string   `json:"anomaly_message,omitempty"`
	AverageDailyConsumption float64  `json:"average_daily_consumption"`
}

type HistoryResponse struct {
	ID              string    `json:"id"`
	ItemName        string    `json:"item_name"`
	Unit            string    `json:"unit"`
	TransactionType string    `json:"transaction_type"`
	OnHandDelta     float64   `json:"on_hand_delta"`
	ReservedDelta   float64   `json:"reserved_delta"`
	OccurredAt      time.Time `json:"occurred_at"`
	Notes           *string   `json:"notes,omitempty"`
}

type StationInventoryResponse struct {
	Items              []ItemResponse           `json:"items"`
	History            []HistoryResponse        `json:"history"`
	Alerts             []InventoryAlertResponse `json:"alerts"`
	AnalyticsAvailable bool                     `json:"analytics_available"`
}

type MutationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}
