package logistics

import "time"

type CargoResponse struct {
	ID                     string  `json:"id"`
	CargoCode              string  `json:"cargo_code"`
	OriginStationName      string  `json:"origin_station_name"`
	DestinationStationName string  `json:"destination_station_name"`
	ExpeditionName         string  `json:"expedition_name"`
	LogisticsBatchCode     string  `json:"logistics_batch_code"`
	Priority               string  `json:"priority"`
	Status                 string  `json:"status"`
	Notes                  *string `json:"notes,omitempty"`
}

type LogisticsBatchResponse struct {
	ID                     string    `json:"id"`
	BatchCode              string    `json:"batch_code"`
	ExpeditionName         string    `json:"expedition_name"`
	OriginStationName      string    `json:"origin_station_name"`
	DestinationStationName string    `json:"destination_station_name"`
	Status                 string    `json:"status"`
	PlannedDispatchAt      time.Time `json:"planned_dispatch_at"`
	EstimatedArrivalAt     time.Time `json:"estimated_arrival_at"`
	CargoCount             int       `json:"cargo_count"`
}

type LogisticsListResponse struct {
	Cargo   []CargoResponse          `json:"cargo"`
	Batches []LogisticsBatchResponse `json:"batches"`
}

type MutationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type CreateCargoResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	CargoID string `json:"cargo_id"`
	QRToken string `json:"qr_token"`
}

type CargoQRItemResponse struct {
	ItemID           string  `json:"item_id"`
	ItemCode         string  `json:"item_code"`
	ItemName         string  `json:"item_name"`
	Unit             string  `json:"unit"`
	Quantity         float64 `json:"quantity"`
	DeclaredWeightKg float64 `json:"declared_weight_kg"`
	Notes            *string `json:"notes,omitempty"`
}

type CargoQRScanResponse struct {
	ID                     string    `json:"id"`
	EventType              string    `json:"event_type"`
	StationName            string    `json:"station_name"`
	ScannedByPersonnelName string    `json:"scanned_by_personnel_name"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	ScannedAt              time.Time `json:"scanned_at"`
	Notes                  *string   `json:"notes,omitempty"`
}

type CargoQRDetailResponse struct {
	ID                     string                `json:"id"`
	QRToken                string                `json:"qr_token"`
	CargoCode              string                `json:"cargo_code"`
	OriginStationName      string                `json:"origin_station_name"`
	DestinationStationName string                `json:"destination_station_name"`
	ExpeditionCode         string                `json:"expedition_code"`
	ExpeditionName         string                `json:"expedition_name"`
	LogisticsBatchCode     string                `json:"logistics_batch_code"`
	Priority               string                `json:"priority"`
	Status                 string                `json:"status"`
	Notes                  *string               `json:"notes,omitempty"`
	DispatchedAt           *time.Time            `json:"dispatched_at,omitempty"`
	ReceivedAt             *time.Time            `json:"received_at,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	Items                  []CargoQRItemResponse `json:"items"`
	ScanHistory            []CargoQRScanResponse `json:"scan_history"`
}

type RecordCargoQRScanResponse struct {
	Status      bool      `json:"status"`
	Message     string    `json:"message"`
	ScanID      string    `json:"scan_id"`
	CargoID     string    `json:"cargo_id"`
	CargoStatus string    `json:"cargo_status"`
	ScannedAt   time.Time `json:"scanned_at"`
}
