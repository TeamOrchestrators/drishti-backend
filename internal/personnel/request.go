package personnel

import (
	"time"

	"github.com/google/uuid"
)

type AddPersonnelRequest struct {
	ExpeditionId         uuid.UUID `json:"expedition_id"`
	PersonnelId          uuid.UUID `json:"personnel_id"`
	OriginStationId      uuid.UUID `json:"origin_station_id"`
	DestinationStationId uuid.UUID `json:"destination_station_id"`
	DepartureTime        time.Time `json:"departure_time"`
	ArrivalTime          time.Time `json:"arrival_time"`
	MovementStatus       string    `json:"movement_status"`
}
