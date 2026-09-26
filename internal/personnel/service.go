package personnel

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service owns personnel business logic and sqlc access.
type Service struct{ queries *db.Queries }

func NewService(queries *db.Queries) *Service { return &Service{queries: queries} }

func (s *Service) List(ctx context.Context) (ListPersonnelPerCategory, error) {
	moving, err := s.queries.ListMovingPersonnel(ctx)
	if err != nil {
		return ListPersonnelPerCategory{}, fmt.Errorf("list moving personnel: %w", err)
	}
	personnel, err := s.queries.ListPersonnel(ctx)
	if err != nil {
		return ListPersonnelPerCategory{}, fmt.Errorf("list personnel: %w", err)
	}
	history, err := s.queries.ListMovementHistory(ctx)
	if err != nil {
		return ListPersonnelPerCategory{}, fmt.Errorf("list movement history: %w", err)
	}

	response := ListPersonnelPerCategory{
		MovingPersonnel: make([]ListMovingPersonnel, 0, len(moving)),
		TotalPersonnel:  make([]ListTotalPersonnel, 0, len(personnel)),
		MovementHistory: make([]MovementHistory, 0, len(history)),
	}
	for _, row := range moving {
		response.MovingPersonnel = append(response.MovingPersonnel, ListMovingPersonnel{
			PersonnelId:            pgtypeUUIDToUUID(row.PersonnelID),
			Name:                   row.Name,
			Role:                   row.Role,
			CurrentStation:         row.CurrentStation,
			OriginStationName:      row.OriginStationName,
			DestinationStationName: row.DestinationStationName,
			DepartureTime:          pgtypeTimeToTime(row.DepartureTime),
			ArrivalTime:            pgtypeTimeToTime(row.ArrivalTime),
			Status:                 row.Status,
		})
	}
	for _, row := range personnel {
		response.TotalPersonnel = append(response.TotalPersonnel, ListTotalPersonnel{
			PersonnelId:    pgtypeUUIDToUUID(row.PersonnelID),
			Name:           row.Name,
			Role:           row.Role,
			CurrentStation: row.CurrentStation,
			Status:         row.Status,
		})
	}
	for _, row := range history {
		response.MovementHistory = append(response.MovementHistory, MovementHistory{
			PersonnelId:            pgtypeUUIDToUUID(row.PersonnelID),
			Name:                   row.Name,
			OriginStationName:      row.OriginStationName,
			DestinationStationName: row.DestinationStationName,
			DepartureTime:          pgtypeTimeToTime(row.DepartureTime),
			ArrivalTime:            pgtypeTimeToTime(row.ArrivalTime),
			Status:                 row.Status,
		})
	}

	return response, nil
}

func (s *Service) AssignmentFormOptions(ctx context.Context) (AssignmentFormOptionsResponse, error) {
	expeditions, err := s.queries.ListAssignableExpeditions(ctx)
	if err != nil {
		return AssignmentFormOptionsResponse{}, fmt.Errorf("list expedition options: %w", err)
	}
	personnel, err := s.queries.ListAssignablePersonnel(ctx)
	if err != nil {
		return AssignmentFormOptionsResponse{}, fmt.Errorf("list personnel options: %w", err)
	}

	response := AssignmentFormOptionsResponse{
		Expeditions: make([]AssignmentExpeditionOption, 0, len(expeditions)),
		Personnel:   make([]AssignmentPersonnelOption, 0, len(personnel)),
		MovementStatuses: []string{
			"planned",
			"in_transit",
			"arrived",
			"cancelled",
		},
	}
	for _, row := range expeditions {
		response.Expeditions = append(response.Expeditions, AssignmentExpeditionOption{
			ID:                     pgtypeUUIDToUUID(row.ID),
			Name:                   row.Name,
			OriginStationName:      row.OriginStationName,
			DestinationStationName: row.DestinationStationName,
		})
	}
	for _, row := range personnel {
		response.Personnel = append(response.Personnel, AssignmentPersonnelOption{
			ID:             pgtypeUUIDToUUID(row.ID),
			Name:           row.Name,
			Role:           row.Role,
			CurrentStation: row.CurrentStation,
		})
	}

	return response, nil
}

func (s *Service) Detail(ctx context.Context, personnelID uuid.UUID) (PersonnelDetailResponse, error) {
	profile, err := s.queries.GetPersonnelDetailProfile(ctx, uuidToPgtype(personnelID))
	if err != nil {
		return PersonnelDetailResponse{}, fmt.Errorf("get personnel profile: %w", err)
	}
	movements, err := s.queries.ListPersonnelMovementTimeline(ctx, uuidToPgtype(personnelID))
	if err != nil {
		return PersonnelDetailResponse{}, fmt.Errorf("list personnel movement timeline: %w", err)
	}
	assignments, err := s.queries.ListPersonnelExpeditionAssignments(ctx, uuidToPgtype(personnelID))
	if err != nil {
		return PersonnelDetailResponse{}, fmt.Errorf("list personnel expedition assignments: %w", err)
	}
	emergencies, err := s.queries.ListPersonnelEmergencies(ctx, uuidToPgtype(personnelID))
	if err != nil {
		return PersonnelDetailResponse{}, fmt.Errorf("list personnel emergencies: %w", err)
	}
	locations, err := s.queries.ListPersonnelLocationHistory(ctx, uuidToPgtype(personnelID))
	if err != nil {
		return PersonnelDetailResponse{}, fmt.Errorf("list personnel location history: %w", err)
	}

	response := PersonnelDetailResponse{
		ID:                     pgtypeUUIDToUUID(profile.PersonnelID).String(),
		PersonnelCode:          profile.PersonnelCode,
		Name:                   profile.FullName,
		Role:                   profile.Role,
		MedicalClearanceStatus: profile.MedicalClearanceStatus,
		Status:                 profile.PersonnelStatus,
		CurrentStationName:     profile.CurrentStationName,
		MovementTimeline:       make([]PersonnelMovementDetail, 0, len(movements)),
		ExpeditionAssignments:  make([]PersonnelExpeditionAssignment, 0, len(assignments)),
		Emergencies:            make([]PersonnelEmergencyResponse, 0, len(emergencies)),
		LocationHistory:        make([]PersonnelLocationPoint, 0, len(locations)),
	}
	if profile.DeviceID.Valid {
		response.Device = &PersonnelDeviceResponse{
			ID:                pgtypeUUIDToUUID(profile.DeviceID).String(),
			Label:             stringPtrValue(profile.DeviceLabel),
			Status:            stringPtrValue(profile.DeviceStatus),
			LastHeartbeatAt:   pgtypeTimePtr(profile.LastHeartbeatAt),
			Latitude:          profile.LastLatitude,
			Longitude:         profile.LastLongitude,
			LocationAccuracyM: profile.LastAccuracyM,
			AltitudeM:         profile.LastAltitudeM,
			HeadingDeg:        profile.LastHeadingDeg,
			SpeedMps:          profile.LastSpeedMps,
			BatteryPercent:    profile.LastBatteryPercent,
		}
	}
	for _, row := range movements {
		response.MovementTimeline = append(response.MovementTimeline, PersonnelMovementDetail{
			ID: uuidToString(row.ID), MovementType: row.MovementType,
			OriginStationName: row.OriginStationName, DestinationStationName: row.DestinationStationName,
			Status: row.Status, DepartedAt: pgtypeTimePtr(row.DepartedAt),
			EstimatedArrivalAt: pgtypeTimePtr(row.EstimatedArrivalAt), ArrivedAt: pgtypeTimePtr(row.ArrivedAt),
			Notes: row.Notes, ExpeditionCode: row.ExpeditionCode, ExpeditionName: row.ExpeditionName,
		})
	}
	activeMovement, err := s.queries.GetPersonnelActiveMovement(ctx, uuidToPgtype(personnelID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return PersonnelDetailResponse{}, fmt.Errorf("get personnel active movement: %w", err)
	}
	if err == nil {
		response.ActiveMovement = &PersonnelMovementDetail{
			ID: uuidToString(activeMovement.ID), MovementType: activeMovement.MovementType,
			OriginStationName: activeMovement.OriginStationName, DestinationStationName: activeMovement.DestinationStationName,
			Status: activeMovement.Status, DepartedAt: pgtypeTimePtr(activeMovement.DepartedAt),
			EstimatedArrivalAt: pgtypeTimePtr(activeMovement.EstimatedArrivalAt),
			ExpeditionCode:     activeMovement.ExpeditionCode, ExpeditionName: activeMovement.ExpeditionName,
		}
	}
	for _, row := range assignments {
		response.ExpeditionAssignments = append(response.ExpeditionAssignments, PersonnelExpeditionAssignment{
			ExpeditionID: uuidToString(row.ExpeditionID), ExpeditionCode: row.ExpeditionCode,
			ExpeditionName: row.ExpeditionName, ExpeditionStatus: row.ExpeditionStatus,
			AssignmentRole: row.AssignmentRole, AssignedAt: pgtypeTimeToTime(row.AssignedAt),
			ReleasedAt: pgtypeTimePtr(row.ReleasedAt), OriginStationName: row.OriginStationName,
			DestinationStationName: row.DestinationStationName,
		})
	}
	for _, row := range emergencies {
		response.Emergencies = append(response.Emergencies, PersonnelEmergencyResponse{
			EmergencyID: uuidToString(row.EmergencyID), EmergencyCode: row.EmergencyCode,
			EmergencyType: row.EmergencyType, Severity: row.Severity, Status: row.EmergencyStatus,
			Summary: row.Summary, ReportedAt: pgtypeTimeToTime(row.ReportedAt),
			ResolvedAt: pgtypeTimePtr(row.ResolvedAt), InvolvementType: row.InvolvementType,
		})
	}
	for _, row := range locations {
		response.LocationHistory = append(response.LocationHistory, PersonnelLocationPoint{
			ID: uuidToString(row.ID), SignalType: row.SignalType, Latitude: row.Latitude,
			Longitude: row.Longitude, LocationAccuracyM: row.LocationAccuracyM,
			HeadingDeg: row.HeadingDeg, SpeedMps: row.SpeedMps,
			BatteryPercent: row.BatteryPercent, OccurredAt: pgtypeTimeToTime(row.OccurredAt),
		})
	}
	return response, nil
}

func (s *Service) AddToExpedition(ctx context.Context, request AddPersonnelRequest) (AddPersonnelResponse, error) {
	if err := validateAssignment(request); err != nil {
		return AddPersonnelResponse{}, err
	}

	_, err := s.queries.AssignPersonnelToExpedition(ctx, db.AssignPersonnelToExpeditionParams{
		ExpeditionID:         uuidToPgtype(request.ExpeditionId),
		PersonnelID:          uuidToPgtype(request.PersonnelId),
		OriginStationID:      uuidToPgtype(request.OriginStationId),
		DestinationStationID: uuidToPgtype(request.DestinationStationId),
		MovementStatus:       request.MovementStatus,
		DepartureTime:        timeToPgtype(request.DepartureTime),
		ArrivalTime:          timeToPgtype(request.ArrivalTime),
	})
	if err != nil {
		return AddPersonnelResponse{}, fmt.Errorf("assign personnel to expedition: %w", err)
	}

	return AddPersonnelResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "personnel assigned to expedition",
	}, nil
}

func validateAssignment(request AddPersonnelRequest) error {
	if request.ExpeditionId == uuid.Nil || request.PersonnelId == uuid.Nil ||
		request.OriginStationId == uuid.Nil || request.DestinationStationId == uuid.Nil {
		return fmt.Errorf("expedition_id, personnel_id, origin_station_id, and destination_station_id are required")
	}
	if request.OriginStationId == request.DestinationStationId {
		return fmt.Errorf("origin and destination stations must differ")
	}
	if request.DepartureTime.IsZero() || request.ArrivalTime.IsZero() {
		return fmt.Errorf("departure_time and arrival_time are required")
	}
	if request.ArrivalTime.Before(request.DepartureTime) {
		return fmt.Errorf("arrival_time must not be before departure_time")
	}
	switch request.MovementStatus {
	case "planned", "in_transit", "arrived", "cancelled":
		return nil
	default:
		return fmt.Errorf("invalid movement_status")
	}
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func timeToPgtype(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgtypeUUIDToUUID(value pgtype.UUID) uuid.UUID {
	if !value.Valid {
		return uuid.Nil
	}
	return uuid.UUID(value.Bytes)
}

func pgtypeTimeToTime(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

func pgtypeTimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func uuidToString(value pgtype.UUID) string { return pgtypeUUIDToUUID(value).String() }

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
