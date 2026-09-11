package expedition

import "time"

type CreateResponse struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type ListResponse struct {
	ID                     string    `json:"id"`
	ExpeditionName         string    `json:"expedition_name"`
	ExpeditionCode         string    `json:"expedition_code"`
	OriginStationName      string    `json:"origin_station_name"`
	DestinationStationName string    `json:"destination_station_name"`
	TeamLeadName           string    `json:"team_lead_name"`
	CrewAssigned           int       `json:"crew_assigned"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	Status                 string    `json:"status"`
}

type StationOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type PersonnelOption struct {
	ID            string `json:"id"`
	PersonnelCode string `json:"personnel_code"`
	FullName      string `json:"full_name"`
	Role          string `json:"role"`
}

type FormOptionsResponse struct {
	Stations  []StationOption   `json:"stations"`
	Personnel []PersonnelOption `json:"personnel"`
}
