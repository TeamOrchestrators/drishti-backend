package expedition

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct{ queries *db.Queries }

func NewService(queries *db.Queries) *Service { return &Service{queries: queries} }

func (s *Service) List(ctx context.Context) ([]ListResponse, error) {
	rows, err := s.queries.ListExpeditions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list expeditions: %w", err)
	}

	responses := make([]ListResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, ListResponse{
			ID:                     pgtypeUUIDToString(row.ID),
			ExpeditionName:         row.ExpeditionName,
			ExpeditionCode:         row.ExpeditionCode,
			OriginStationName:      row.OriginStationName,
			DestinationStationName: row.DestinationStationName,
			TeamLeadName:           row.TeamLeadName,
			CrewAssigned:           int(row.CrewAssigned),
			StartDate:              pgtypeTimestampToTime(row.StartDate),
			EndDate:                pgtypeTimestampToTime(row.EndDate),
			Status:                 row.Status,
		})
	}

	return responses, nil
}

func (s *Service) FormOptions(ctx context.Context) (FormOptionsResponse, error) {
	stations, err := s.queries.ListStations(ctx)
	if err != nil {
		return FormOptionsResponse{}, fmt.Errorf("list station options: %w", err)
	}
	personnel, err := s.queries.ListActivePersonnel(ctx)
	if err != nil {
		return FormOptionsResponse{}, fmt.Errorf("list personnel options: %w", err)
	}

	response := FormOptionsResponse{
		Stations:  make([]StationOption, 0, len(stations)),
		Personnel: make([]PersonnelOption, 0, len(personnel)),
	}
	for _, station := range stations {
		response.Stations = append(response.Stations, StationOption{
			ID:   pgtypeUUIDToString(station.ID),
			Code: station.Code,
			Name: station.Name,
		})
	}
	for _, person := range personnel {
		response.Personnel = append(response.Personnel, PersonnelOption{
			ID:            pgtypeUUIDToString(person.ID),
			PersonnelCode: person.PersonnelCode,
			FullName:      person.FullName,
			Role:          person.Role,
		})
	}

	return response, nil
}

func (s *Service) Create(ctx context.Context, request CreateRequestBody) (CreateResponse, error) {
	params, err := buildCreateParams(request)
	if err != nil {
		return CreateResponse{}, err
	}

	if _, err := s.queries.CreateExpedition(ctx, params); err != nil {
		return CreateResponse{}, fmt.Errorf("create expedition: %w", err)
	}

	return CreateResponse{Status: true, StatusCode: 201, Message: "expedition created"}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, request CreateRequestBody) (CreateResponse, error) {
	createParams, err := buildCreateParams(request)
	if err != nil {
		return CreateResponse{}, err
	}

	params := db.UpdateExpeditionParams{
		ID:                   uuidToPgtype(id),
		ExpeditionCode:       createParams.ExpeditionCode,
		Name:                 createParams.Name,
		Purpose:              createParams.Purpose,
		OriginStationID:      createParams.OriginStationID,
		DestinationStationID: createParams.DestinationStationID,
		LeadPersonnelID:      createParams.LeadPersonnelID,
		PlannedStartAt:       createParams.PlannedStartAt,
		PlannedEndAt:         createParams.PlannedEndAt,
		Status:               createParams.Status,
	}
	if _, err := s.queries.UpdateExpedition(ctx, params); err != nil {
		return CreateResponse{}, fmt.Errorf("update expedition: %w", err)
	}

	return CreateResponse{Status: true, StatusCode: 200, Message: "expedition updated"}, nil
}

func buildCreateParams(request CreateRequestBody) (db.CreateExpeditionParams, error) {
	if strings.TrimSpace(request.ExpeditionCode) == "" || strings.TrimSpace(request.Name) == "" {
		return db.CreateExpeditionParams{}, fmt.Errorf("expedition_code and name are required")
	}
	if request.OriginStationId == request.DestinationStationId {
		return db.CreateExpeditionParams{}, fmt.Errorf("origin and destination stations must differ")
	}
	if request.EndDate.Before(request.StartDate) {
		return db.CreateExpeditionParams{}, fmt.Errorf("end_date must not be before start_date")
	}

	status := request.Status
	if status == "" {
		status = "draft"
	}
	if !isValidStatus(status) {
		return db.CreateExpeditionParams{}, fmt.Errorf("invalid expedition status")
	}

	purpose := strings.TrimSpace(request.Purpose)
	params := db.CreateExpeditionParams{
		ExpeditionCode:       strings.TrimSpace(request.ExpeditionCode),
		Name:                 strings.TrimSpace(request.Name),
		OriginStationID:      uuidToPgtype(request.OriginStationId),
		DestinationStationID: uuidToPgtype(request.DestinationStationId),
		LeadPersonnelID:      uuidToPgtype(request.TeamLeaderId),
		PlannedStartAt:       timeToPgtype(request.StartDate),
		PlannedEndAt:         timeToPgtype(request.EndDate),
		Status:               status,
	}
	if purpose != "" {
		params.Purpose = &purpose
	}

	return params, nil
}

func isValidStatus(status string) bool {
	switch status {
	case "draft", "planned", "ready", "active", "sheltered", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func timeToPgtype(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgtypeUUIDToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func pgtypeTimestampToTime(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
