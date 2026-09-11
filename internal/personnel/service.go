package personnel

import (
	"context"
	"fmt"
	"time"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/google/uuid"
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
