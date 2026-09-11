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
