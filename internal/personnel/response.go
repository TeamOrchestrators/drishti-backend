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

type PersonnelDeviceResponse struct {
	ID                string     `json:"id"`
	Label             string     `json:"label"`
	Status            string     `json:"status"`
	LastHeartbeatAt   *time.Time `json:"last_heartbeat_at,omitempty"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	LocationAccuracyM float64    `json:"location_accuracy_m"`
	AltitudeM         float64    `json:"altitude_m"`
	HeadingDeg        float64    `json:"heading_deg"`
	SpeedMps          float64    `json:"speed_mps"`
	BatteryPercent    float64    `json:"battery_percent"`
}

type PersonnelMovementDetail struct {
	ID                     string     `json:"id"`
	MovementType           string     `json:"movement_type"`
	OriginStationName      string     `json:"origin_station_name"`
	DestinationStationName string     `json:"destination_station_name"`
	Status                 string     `json:"status"`
	DepartedAt             *time.Time `json:"departed_at,omitempty"`
	EstimatedArrivalAt     *time.Time `json:"estimated_arrival_at,omitempty"`
	ArrivedAt              *time.Time `json:"arrived_at,omitempty"`
	Notes                  *string    `json:"notes,omitempty"`
	ExpeditionCode         string     `json:"expedition_code"`
	ExpeditionName         string     `json:"expedition_name"`
}

type PersonnelExpeditionAssignment struct {
	ExpeditionID           string     `json:"expedition_id"`
	ExpeditionCode         string     `json:"expedition_code"`
	ExpeditionName         string     `json:"expedition_name"`
	ExpeditionStatus       string     `json:"expedition_status"`
	AssignmentRole         *string    `json:"assignment_role,omitempty"`
	AssignedAt             time.Time  `json:"assigned_at"`
	ReleasedAt             *time.Time `json:"released_at,omitempty"`
	OriginStationName      string     `json:"origin_station_name"`
	DestinationStationName string     `json:"destination_station_name"`
}

type PersonnelEmergencyResponse struct {
	EmergencyID     string     `json:"emergency_id"`
	EmergencyCode   string     `json:"emergency_code"`
	EmergencyType   string     `json:"emergency_type"`
	Severity        string     `json:"severity"`
	Status          string     `json:"status"`
	Summary         string     `json:"summary"`
	ReportedAt      time.Time  `json:"reported_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	InvolvementType string     `json:"involvement_type"`
}

type PersonnelLocationPoint struct {
	ID                string    `json:"id"`
	SignalType        string    `json:"signal_type"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	LocationAccuracyM float64   `json:"location_accuracy_m"`
	HeadingDeg        float64   `json:"heading_deg"`
	SpeedMps          float64   `json:"speed_mps"`
	BatteryPercent    float64   `json:"battery_percent"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type PersonnelDetailResponse struct {
	ID                     string                          `json:"id"`
	PersonnelCode          string                          `json:"personnel_code"`
	Name                   string                          `json:"name"`
	Role                   string                          `json:"role"`
	MedicalClearanceStatus string                          `json:"medical_clearance_status"`
	Status                 string                          `json:"status"`
	CurrentStationName     string                          `json:"current_station_name"`
	Device                 *PersonnelDeviceResponse        `json:"device,omitempty"`
	ActiveMovement         *PersonnelMovementDetail        `json:"active_movement,omitempty"`
	MovementTimeline       []PersonnelMovementDetail       `json:"movement_timeline"`
	ExpeditionAssignments  []PersonnelExpeditionAssignment `json:"expedition_assignments"`
	Emergencies            []PersonnelEmergencyResponse    `json:"emergencies"`
	LocationHistory        []PersonnelLocationPoint        `json:"location_history"`
}
