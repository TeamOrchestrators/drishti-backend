package emergency

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service owns emergency business logic and sqlc access.
type Service struct{ queries *db.Queries }

func NewService(queries *db.Queries) *Service { return &Service{queries: queries} }

func (s *Service) DeviceFormOptions(ctx context.Context) (DeviceFormOptionsResponse, error) {
	rows, err := s.queries.ListEmergencyDeviceFormOptions(ctx)
	if err != nil {
		return DeviceFormOptionsResponse{}, fmt.Errorf("list device form options: %w", err)
	}
	response := DeviceFormOptionsResponse{Personnel: make([]DevicePersonnelOption, 0, len(rows))}
	for _, row := range rows {
		option := DevicePersonnelOption{PersonnelID: uuidFromPg(row.PersonnelID).String(), PersonnelName: row.PersonnelName, Role: row.Role, CurrentStationName: row.CurrentStationName}
		if row.DeviceID.Valid {
			value := uuidFromPg(row.DeviceID).String()
			option.DeviceID = &value
		}
		option.DeviceLabel = row.DeviceLabel
		option.DeviceStatus = row.DeviceStatus
		response.Personnel = append(response.Personnel, option)
	}
	return response, nil
}

func (s *Service) RegisterDevice(ctx context.Context, request RegisterDeviceRequest) (DeviceResponse, error) {
	if request.PersonnelID == uuid.Nil || strings.TrimSpace(request.DeviceLabel) == "" {
		return DeviceResponse{}, fmt.Errorf("personnel_id and device_label are required")
	}
	row, err := s.queries.RegisterEmergencyDevice(ctx, db.RegisterEmergencyDeviceParams{PersonnelID: pgUUID(request.PersonnelID), DeviceLabel: strings.TrimSpace(request.DeviceLabel)})
	if err != nil {
		return DeviceResponse{}, fmt.Errorf("register device: %w", err)
	}
	return DeviceResponse{DeviceID: uuidFromPg(row.ID).String(), PersonnelID: uuidFromPg(row.PersonnelID).String(), DeviceLabel: row.DeviceLabel, Status: row.Status, LastHeartbeatAt: timePtr(row.LastHeartbeatAt)}, nil
}

func (s *Service) Heartbeat(ctx context.Context, deviceID uuid.UUID, request HeartbeatRequest) (HeartbeatResponse, error) {
	if deviceID == uuid.Nil || request.IdempotencyKey == uuid.Nil || request.OccurredAt.IsZero() {
		return HeartbeatResponse{}, fmt.Errorf("device_id, idempotency_key, and occurred_at are required")
	}
	if err := validatePosition(request.Latitude, request.Longitude, request.LocationAccuracyM, request.AltitudeM, request.HeadingDeg, request.SpeedMps, request.BatteryPercent); err != nil {
		return HeartbeatResponse{}, err
	}
	row, err := s.queries.RecordEmergencyHeartbeat(ctx, db.RecordEmergencyHeartbeatParams{
		DeviceID: pgUUID(deviceID), IdempotencyKey: pgUUID(request.IdempotencyKey), OccurredAt: pgTime(request.OccurredAt),
		Latitude: numeric(request.Latitude), Longitude: numeric(request.Longitude), LocationAccuracyM: numeric(request.LocationAccuracyM),
		AltitudeM: numericPtr(request.AltitudeM), HeadingDeg: numericPtr(request.HeadingDeg), SpeedMps: numericPtr(request.SpeedMps), BatteryPercent: numericPtr(request.BatteryPercent),
	})
	if err != nil {
		return HeartbeatResponse{}, fmt.Errorf("record heartbeat: %w", err)
	}
	response := HeartbeatResponse{Acknowledged: true, DeviceID: uuidFromPg(row.DeviceID).String(), PersonnelID: uuidFromPg(row.PersonnelID).String(), ServerReceivedAt: time.Now().UTC(), NextHeartbeatSeconds: 15}
	if row.EmergencyID.Valid {
		value := uuidFromPg(row.EmergencyID).String()
		response.EmergencyID = &value
	}
	return response, nil
}

func (s *Service) InitiateSOS(ctx context.Context, deviceID uuid.UUID, request InitiateSOSRequest) (SOSInitiatedResponse, error) {
	if deviceID == uuid.Nil || strings.TrimSpace(request.EmergencyType) == "" || strings.TrimSpace(request.Summary) == "" || !validSeverity(request.Severity) {
		return SOSInitiatedResponse{}, fmt.Errorf("device_id, emergency_type, summary, and a valid severity are required")
	}
	row, err := s.queries.InitiateEmergencySOS(ctx, db.InitiateEmergencySOSParams{DeviceID: pgUUID(deviceID), EmergencyType: strings.TrimSpace(request.EmergencyType), Severity: request.Severity, Summary: strings.TrimSpace(request.Summary)})
	if err != nil {
		return SOSInitiatedResponse{}, fmt.Errorf("initiate SOS: device must be active and must send a heartbeat first: %w", err)
	}
	return SOSInitiatedResponse{ConfirmationID: uuidFromPg(row.ID).String(), DeviceID: uuidFromPg(row.DeviceID).String(), PersonnelID: uuidFromPg(row.PersonnelID).String(), Status: row.Status, ExpiresAt: timeFromPg(row.ExpiresAt)}, nil
}

func (s *Service) ConfirmSOS(ctx context.Context, deviceID, confirmationID uuid.UUID, request ConfirmSOSRequest) (EmergencyMutationResponse, error) {
	if deviceID == uuid.Nil || confirmationID == uuid.Nil || request.IdempotencyKey == uuid.Nil {
		return EmergencyMutationResponse{}, fmt.Errorf("device_id, confirmation_id, and idempotency_key are required")
	}
	row, err := s.queries.ConfirmEmergencySOS(ctx, db.ConfirmEmergencySOSParams{DeviceID: pgUUID(deviceID), ConfirmationID: pgUUID(confirmationID), IdempotencyKey: pgUUID(request.IdempotencyKey)})
	if err != nil {
		return EmergencyMutationResponse{}, fmt.Errorf("confirm SOS: confirmation is invalid, expired, or already used: %w", err)
	}
	reportedAt := timeFromPg(row.ReportedAt)
	return EmergencyMutationResponse{ID: uuidFromPg(row.ID).String(), EmergencyCode: row.EmergencyCode, Status: row.Status, ReportedAt: &reportedAt}, nil
}

func (s *Service) List(ctx context.Context) (EmergencyListResponse, error) {
	rows, err := s.queries.ListEmergencies(ctx)
	if err != nil {
		return EmergencyListResponse{}, fmt.Errorf("list emergencies: %w", err)
	}
	response := EmergencyListResponse{Emergencies: make([]EmergencyResponse, 0, len(rows))}
	for _, row := range rows {
		item := EmergencyResponse{ID: uuidFromPg(row.ID).String(), EmergencyCode: row.EmergencyCode, EmergencyType: row.EmergencyType, Severity: row.Severity, Status: row.Status, Summary: row.Summary, Latitude: row.Latitude, Longitude: row.Longitude, LocationAccuracyM: row.LocationAccuracyM, ReportedAt: timeFromPg(row.ReportedAt), ResolvedAt: timePtr(row.ResolvedAt), ReportedByName: row.ReportedByName, DeviceLabel: row.DeviceLabel, LastHeartbeatAt: timePtr(row.LastHeartbeatAt), HeadingDeg: row.HeadingDeg, SpeedMps: row.SpeedMps, BatteryPercent: row.BatteryPercent, PeopleAffected: int(row.PeopleAffected), ResourcesNeeded: string(row.ResourcesNeeded)}
		if row.DeviceID.Valid {
			value := uuidFromPg(row.DeviceID).String()
			item.DeviceID = &value
		}
		response.Emergencies = append(response.Emergencies, item)
	}
	return response, nil
}

func (s *Service) UpdateStatus(ctx context.Context, emergencyID uuid.UUID, request UpdateStatusRequest) (EmergencyMutationResponse, error) {
	if emergencyID == uuid.Nil || !validEmergencyStatus(request.Status) {
		return EmergencyMutationResponse{}, fmt.Errorf("a valid emergency status is required")
	}
	row, err := s.queries.UpdateEmergencyStatus(ctx, db.UpdateEmergencyStatusParams{EmergencyID: pgUUID(emergencyID), Status: request.Status})
	if err != nil {
		return EmergencyMutationResponse{}, fmt.Errorf("update emergency status: %w", err)
	}
	return EmergencyMutationResponse{ID: uuidFromPg(row.ID).String(), EmergencyCode: row.EmergencyCode, Status: row.Status, ResolvedAt: timePtr(row.ResolvedAt)}, nil
}

func validSeverity(value string) bool {
	return value == "moderate" || value == "critical" || value == "immediate_response"
}
func validEmergencyStatus(value string) bool {
	return value == "active" || value == "acknowledged" || value == "responding" || value == "resolved" || value == "cancelled"
}

func validatePosition(latitude, longitude, accuracy float64, altitude, heading, speed, battery *float64) error {
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 || accuracy < 0 {
		return fmt.Errorf("invalid latitude, longitude, or location_accuracy_m")
	}
	if (heading != nil && (*heading < 0 || *heading >= 360)) || (speed != nil && *speed < 0) || (battery != nil && (*battery < 0 || *battery > 100)) {
		return fmt.Errorf("invalid heading_deg, speed_mps, or battery_percent")
	}
	_ = altitude
	return nil
}
func numeric(value float64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(int64(math.Round(value * 1000))), Exp: -3, Valid: true}
}
func numericPtr(value *float64) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}
	return numeric(*value)
}
func pgUUID(value uuid.UUID) pgtype.UUID        { return pgtype.UUID{Bytes: value, Valid: true} }
func pgTime(value time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: value, Valid: true} }
func uuidFromPg(value pgtype.UUID) uuid.UUID {
	if !value.Valid {
		return uuid.Nil
	}
	return uuid.UUID(value.Bytes)
}
func timeFromPg(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
