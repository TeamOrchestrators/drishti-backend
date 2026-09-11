package expedition

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequestBody struct {
	ExpeditionCode       string    `json:"expedition_code"`
	Name                 string    `json:"name"`
	Purpose              string    `json:"purpose"`
	OriginStationId      uuid.UUID `json:"origin_station_id"`
	DestinationStationId uuid.UUID `json:"destination_station_id"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	TeamLeaderId         uuid.UUID `json:"team_leader_id"`
	Status               string    `json:"status,omitempty"`
}
