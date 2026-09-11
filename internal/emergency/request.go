package emergency

import (
	"time"

	"github.com/google/uuid"
)

type RegisterDeviceRequest struct {
	PersonnelID uuid.UUID `json:"personnel_id"`
	DeviceLabel string    `json:"device_label"`
}

// HeartbeatRequest is sent periodically by the mobile HTTP device simulator.
// IdempotencyKey must be unique for each actual device transmission.
type HeartbeatRequest struct {
	IdempotencyKey    uuid.UUID `json:"idempotency_key"`
	OccurredAt        time.Time `json:"occurred_at"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	LocationAccuracyM float64   `json:"location_accuracy_m"`
	AltitudeM         *float64  `json:"altitude_m,omitempty"`
	HeadingDeg        *float64  `json:"heading_deg,omitempty"`
	SpeedMps          *float64  `json:"speed_mps,omitempty"`
	BatteryPercent    *float64  `json:"battery_percent,omitempty"`
}

type InitiateSOSRequest struct {
	EmergencyType string `json:"emergency_type"`
	Severity      string `json:"severity"`
	Summary       string `json:"summary"`
}

type ConfirmSOSRequest struct {
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}
