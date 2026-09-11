package personnel

import (
	"time"

	"github.com/google/uuid"
)

type ListPersonnelPerCategory struct {
	MovingPersonnel []ListMovingPersonnel `json:"moving_personnel"`
	TotalPersonnel  []ListTotalPersonnel  `json:"total_personnel"`
	MovementHistory []MovementHistory     `json:"movement_history"`
}

// ListMovingPersonnel This struct is for personnel that are either in transit between stations or are moving from Base to Nation
type ListMovingPersonnel struct {
	PersonnelId            uuid.UUID `json:"personnel_id"`
	Name                   string    `json:"name"`
	Role                   string    `json:"role"`
	CurrentStation         string    `json:"current_station"`
	OriginStationName      string    `json:"origin_station_name"`
	DestinationStationName string    `json:"destination_station_name"`
	DepartureTime          time.Time `json:"departure_time"`
	ArrivalTime            time.Time `json:"arrival_time"`
	Status                 string    `json:"status"`
}

// ListTotalPersonnel This struct is for all personnel registered in database
type ListTotalPersonnel struct {
	PersonnelId    uuid.UUID `json:"personnel_id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	CurrentStation string    `json:"current_station"`
	Status         string    `json:"status"`
}

// MovementHistory this struct is for movement history of all personnel who arrived safely and completed their movement
type MovementHistory struct {
	PersonnelId            uuid.UUID `json:"personnel_id"`
	Name                   string    `json:"name"`
	OriginStationName      string    `json:"origin_station_name"`
	DestinationStationName string    `json:"destination_station_name"`
	DepartureTime          time.Time `json:"departure_time"`
	ArrivalTime            time.Time `json:"arrival_time"`
	Status                 string    `json:"status"`
}

type AssignmentExpeditionOption struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	OriginStationName      string    `json:"origin_station_name"`
	DestinationStationName string    `json:"destination_station_name"`
}

type AssignmentPersonnelOption struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	CurrentStation string    `json:"current_station"`
}

type AssignmentFormOptionsResponse struct {
	Expeditions      []AssignmentExpeditionOption `json:"expeditions"`
	Personnel        []AssignmentPersonnelOption  `json:"personnel"`
	MovementStatuses []string                     `json:"movement_statuses"`
}

type AddPersonnelResponse struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}
