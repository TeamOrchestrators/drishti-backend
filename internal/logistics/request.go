package logistics

import (
	"time"

	"github.com/google/uuid"
)

type CreateCargoRequest struct {
	CargoCode            string     `json:"cargo_code"`
	OriginStationID      uuid.UUID  `json:"origin_station_id"`
	DestinationStationID uuid.UUID  `json:"destination_station_id"`
	ExpeditionID         *uuid.UUID `json:"expedition_id,omitempty"`
	LogisticsBatchID     *uuid.UUID `json:"logistics_batch_id,omitempty"`
	Priority             string     `json:"priority,omitempty"`
	Status               string     `json:"status,omitempty"`
	Notes                string     `json:"notes,omitempty"`
}

type CreateBatchRequest struct {
	BatchCode          string    `json:"batch_code"`
	ExpeditionID       uuid.UUID `json:"expedition_id"`
	Status             string    `json:"status,omitempty"`
	PlannedDispatchAt  time.Time `json:"planned_dispatch_at"`
	EstimatedArrivalAt time.Time `json:"estimated_arrival_at"`
	Notes              string    `json:"notes,omitempty"`
}

type UpdateBatchRequest struct {
	Status             string    `json:"status"`
	PlannedDispatchAt  time.Time `json:"planned_dispatch_at"`
	EstimatedArrivalAt time.Time `json:"estimated_arrival_at"`
	Notes              string    `json:"notes,omitempty"`
}

type AssignCargoBatchRequest struct {
	LogisticsBatchID uuid.UUID `json:"logistics_batch_id"`
}

// RecordCargoQRScanRequest is sent by a mobile scanner after resolving a cargo QR token.
// GPS coordinates are mandatory so every checkpoint can be mapped and audited.
type RecordCargoQRScanRequest struct {
	EventType            string     `json:"event_type,omitempty"`
	Status               string     `json:"status,omitempty"`
	ApplyToBatch         bool       `json:"apply_to_batch,omitempty"`
	StationID            *uuid.UUID `json:"station_id,omitempty"`
	ScannedByPersonnelID *uuid.UUID `json:"scanned_by_personnel_id,omitempty"`
	Latitude             *float64   `json:"latitude,omitempty"`
	Longitude            *float64   `json:"longitude,omitempty"`
	Notes                string     `json:"notes,omitempty"`
}
