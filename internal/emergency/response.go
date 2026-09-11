package emergency

import "time"

type DeviceFormOptionsResponse struct {
	Personnel []DevicePersonnelOption `json:"personnel"`
}

type DevicePersonnelOption struct {
	PersonnelID        string  `json:"personnel_id"`
	PersonnelName      string  `json:"personnel_name"`
	Role               string  `json:"role"`
	CurrentStationName string  `json:"current_station_name"`
	DeviceID           *string `json:"device_id,omitempty"`
	DeviceLabel        *string `json:"device_label,omitempty"`
	DeviceStatus       *string `json:"device_status,omitempty"`
}

type DeviceResponse struct {
	DeviceID        string     `json:"device_id"`
	PersonnelID     string     `json:"personnel_id"`
	DeviceLabel     string     `json:"device_label"`
	Status          string     `json:"status"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
}

type HeartbeatResponse struct {
	Acknowledged         bool      `json:"acknowledged"`
	DeviceID             string    `json:"device_id"`
	PersonnelID          string    `json:"personnel_id"`
	EmergencyID          *string   `json:"emergency_id,omitempty"`
	ServerReceivedAt     time.Time `json:"server_received_at"`
	NextHeartbeatSeconds int       `json:"next_heartbeat_seconds"`
}

type SOSInitiatedResponse struct {
	ConfirmationID string    `json:"confirmation_id"`
	DeviceID       string    `json:"device_id"`
	PersonnelID    string    `json:"personnel_id"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type EmergencyResponse struct {
	ID                string     `json:"id"`
	EmergencyCode     string     `json:"emergency_code"`
	EmergencyType     string     `json:"emergency_type"`
	Severity          string     `json:"severity"`
	Status            string     `json:"status"`
	Summary           string     `json:"summary"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	LocationAccuracyM float64    `json:"location_accuracy_m"`
	ReportedAt        time.Time  `json:"reported_at"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	ReportedByName    string     `json:"reported_by_name"`
	DeviceID          *string    `json:"device_id,omitempty"`
	DeviceLabel       string     `json:"device_label"`
	LastHeartbeatAt   *time.Time `json:"last_heartbeat_at,omitempty"`
	HeadingDeg        float64    `json:"heading_deg"`
	SpeedMps          float64    `json:"speed_mps"`
	BatteryPercent    float64    `json:"battery_percent"`
	PeopleAffected    int        `json:"people_affected"`
	ResourcesNeeded   string     `json:"resources_needed"`
}

type EmergencyListResponse struct {
	Emergencies []EmergencyResponse `json:"emergencies"`
}

type EmergencyMutationResponse struct {
	ID            string     `json:"id"`
	EmergencyCode string     `json:"emergency_code"`
	Status        string     `json:"status"`
	ReportedAt    *time.Time `json:"reported_at,omitempty"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}
