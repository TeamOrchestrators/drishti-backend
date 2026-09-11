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
