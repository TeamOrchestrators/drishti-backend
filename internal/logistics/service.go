package logistics

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

func (s *Service) List(ctx context.Context) (LogisticsListResponse, error) {
	cargo, err := s.queries.ListCargo(ctx)
	if err != nil {
		return LogisticsListResponse{}, fmt.Errorf("list cargo: %w", err)
	}
	batches, err := s.queries.ListLogisticsBatches(ctx)
	if err != nil {
		return LogisticsListResponse{}, fmt.Errorf("list logistics batches: %w", err)
	}
	response := LogisticsListResponse{Cargo: make([]CargoResponse, 0, len(cargo)), Batches: make([]LogisticsBatchResponse, 0, len(batches))}
	for _, row := range cargo {
		origin := ""
		if row.OriginStationName != nil {
			origin = *row.OriginStationName
		}
		response.Cargo = append(response.Cargo, CargoResponse{ID: uuidFromPg(row.ID).String(), CargoCode: row.CargoCode, OriginStationName: origin, DestinationStationName: row.DestinationStationName, ExpeditionName: row.ExpeditionName, LogisticsBatchCode: row.LogisticsBatchCode, Priority: row.Priority, Status: row.Status, Notes: row.Notes})
	}
	for _, row := range batches {
		response.Batches = append(response.Batches, LogisticsBatchResponse{ID: uuidFromPg(row.ID).String(), BatchCode: row.BatchCode, ExpeditionName: row.ExpeditionName, OriginStationName: row.OriginStationName, DestinationStationName: row.DestinationStationName, Status: row.Status, PlannedDispatchAt: timeFromPg(row.PlannedDispatchAt), EstimatedArrivalAt: timeFromPg(row.EstimatedArrivalAt), CargoCount: int(row.CargoCount)})
	}
	return response, nil
}

func (s *Service) CreateBatch(ctx context.Context, request CreateBatchRequest) (MutationResponse, error) {
	if strings.TrimSpace(request.BatchCode) == "" || request.ExpeditionID == uuid.Nil {
		return MutationResponse{}, fmt.Errorf("batch_code and expedition_id are required")
	}
	if request.EstimatedArrivalAt.Before(request.PlannedDispatchAt) {
		return MutationResponse{}, fmt.Errorf("estimated_arrival_at must not be before planned_dispatch_at")
	}
	status := request.Status
	if status == "" {
		status = "draft"
	}
	if !validBatchStatus(status) {
		return MutationResponse{}, fmt.Errorf("invalid batch status")
	}
	notes := strings.TrimSpace(request.Notes)
	params := db.CreateLogisticsBatchParams{BatchCode: strings.TrimSpace(request.BatchCode), ExpeditionID: pgUUID(request.ExpeditionID), Status: status, PlannedDispatchAt: pgTime(request.PlannedDispatchAt), EstimatedArrivalAt: pgTime(request.EstimatedArrivalAt)}
	if notes != "" {
		params.Notes = &notes
	}
	if _, err := s.queries.CreateLogisticsBatch(ctx, params); err != nil {
		return MutationResponse{}, fmt.Errorf("create logistics batch: %w", err)
	}
	return MutationResponse{Status: true, Message: "logistics batch created"}, nil
}

func (s *Service) CreateCargo(ctx context.Context, request CreateCargoRequest) (CreateCargoResponse, error) {
	if strings.TrimSpace(request.CargoCode) == "" || request.OriginStationID == uuid.Nil || request.DestinationStationID == uuid.Nil {
		return CreateCargoResponse{}, fmt.Errorf("cargo_code, origin_station_id, and destination_station_id are required")
	}
	if request.LogisticsBatchID != nil {
		return CreateCargoResponse{}, fmt.Errorf("create cargo without logistics_batch_id, then use the batch-assignment endpoint")
	}
	if request.OriginStationID == request.DestinationStationID {
		return CreateCargoResponse{}, fmt.Errorf("origin and destination stations must differ")
	}
	priority := request.Priority
	if priority == "" {
		priority = "standard"
	}
	if priority != "standard" && priority != "high" && priority != "critical" {
		return CreateCargoResponse{}, fmt.Errorf("invalid cargo priority")
	}
	status := request.Status
	if status == "" {
		status = "draft"
	}
	if !validCargoStatus(status) {
		return CreateCargoResponse{}, fmt.Errorf("invalid cargo status")
	}
	params := db.CreateCargoParams{CargoCode: strings.TrimSpace(request.CargoCode), OriginStationID: pgUUID(request.OriginStationID), DestinationStationID: pgUUID(request.DestinationStationID), Priority: priority, Status: status}
	if request.ExpeditionID != nil {
		params.ExpeditionID = pgUUID(*request.ExpeditionID)
	}
	if notes := strings.TrimSpace(request.Notes); notes != "" {
		params.Notes = &notes
	}
	cargo, err := s.queries.CreateCargo(ctx, params)
	if err != nil {
		return CreateCargoResponse{}, fmt.Errorf("create cargo: %w", err)
	}
	cargoID := uuidFromPg(cargo.ID).String()
	return CreateCargoResponse{Status: true, Message: "cargo created", CargoID: cargoID, QRToken: cargoID}, nil
}

func (s *Service) AssignCargo(ctx context.Context, cargoID, batchID uuid.UUID) (MutationResponse, error) {
	if cargoID == uuid.Nil || batchID == uuid.Nil {
		return MutationResponse{}, fmt.Errorf("cargo_id and logistics_batch_id are required")
	}
	if _, err := s.queries.AssignCargoToLogisticsBatch(ctx, db.AssignCargoToLogisticsBatchParams{CargoID: pgUUID(cargoID), LogisticsBatchID: pgUUID(batchID)}); err != nil {
		return MutationResponse{}, fmt.Errorf("assign cargo to batch: %w", err)
	}
	return MutationResponse{Status: true, Message: "cargo assigned to logistics batch"}, nil
}

func (s *Service) GetByQR(ctx context.Context, cargoID uuid.UUID) (CargoQRDetailResponse, error) {
	cargo, err := s.queries.GetCargoForQR(ctx, pgUUID(cargoID))
	if err != nil {
		return CargoQRDetailResponse{}, fmt.Errorf("get cargo by QR: %w", err)
	}
	items, err := s.queries.ListCargoItemsForQR(ctx, pgUUID(cargoID))
	if err != nil {
		return CargoQRDetailResponse{}, fmt.Errorf("list cargo QR items: %w", err)
	}
	scans, err := s.queries.ListCargoQRScans(ctx, pgUUID(cargoID))
	if err != nil {
		return CargoQRDetailResponse{}, fmt.Errorf("list cargo QR scans: %w", err)
	}

	response := CargoQRDetailResponse{
		ID:                     uuidFromPg(cargo.ID).String(),
		QRToken:                uuidFromPg(cargo.ID).String(),
		CargoCode:              cargo.CargoCode,
		OriginStationName:      cargo.OriginStationName,
		DestinationStationName: cargo.DestinationStationName,
		ExpeditionCode:         cargo.ExpeditionCode,
		ExpeditionName:         cargo.ExpeditionName,
		LogisticsBatchCode:     cargo.LogisticsBatchCode,
		Priority:               cargo.Priority,
		Status:                 cargo.Status,
		Notes:                  cargo.Notes,
		DispatchedAt:           timePtrFromPg(cargo.DispatchedAt),
		ReceivedAt:             timePtrFromPg(cargo.ReceivedAt),
		CreatedAt:              timeFromPg(cargo.CreatedAt),
		Items:                  make([]CargoQRItemResponse, 0, len(items)),
		ScanHistory:            make([]CargoQRScanResponse, 0, len(scans)),
	}
	for _, item := range items {
		response.Items = append(response.Items, CargoQRItemResponse{
			ItemID:           uuidFromPg(item.ItemID).String(),
			ItemCode:         item.ItemCode,
			ItemName:         item.ItemName,
			Unit:             item.Unit,
			Quantity:         item.Quantity,
			DeclaredWeightKg: item.DeclaredWeightKg,
			Notes:            item.Notes,
		})
	}
	for _, scan := range scans {
		response.ScanHistory = append(response.ScanHistory, CargoQRScanResponse{
			ID:                     uuidFromPg(scan.ID).String(),
			EventType:              scan.EventType,
			StationName:            scan.StationName,
			ScannedByPersonnelName: scan.ScannedByPersonnelName,
			Latitude:               scan.Latitude,
			Longitude:              scan.Longitude,
			ScannedAt:              timeFromPg(scan.ScannedAt),
			Notes:                  scan.Notes,
		})
	}
	return response, nil
}

func (s *Service) RecordQRScan(ctx context.Context, cargoID uuid.UUID, request RecordCargoQRScanRequest) (RecordCargoQRScanResponse, error) {
	cargo, err := s.queries.GetCargoForQR(ctx, pgUUID(cargoID))
	if err != nil {
		return RecordCargoQRScanResponse{}, fmt.Errorf("get cargo by QR: %w", err)
	}
	eventType := strings.TrimSpace(request.EventType)
	if eventType == "" {
		eventType = "scanned"
	}
	if !validCargoScanEvent(eventType) {
		return RecordCargoQRScanResponse{}, fmt.Errorf("invalid QR scan event_type")
	}
	status := strings.TrimSpace(request.Status)
	if status != "" && !validCargoStatus(status) {
		return RecordCargoQRScanResponse{}, fmt.Errorf("invalid cargo status")
	}
	if (request.Latitude == nil) != (request.Longitude == nil) {
		return RecordCargoQRScanResponse{}, fmt.Errorf("latitude and longitude must be provided together")
	}
	if request.Latitude != nil && (*request.Latitude < -90 || *request.Latitude > 90 || *request.Longitude < -180 || *request.Longitude > 180) {
		return RecordCargoQRScanResponse{}, fmt.Errorf("invalid latitude or longitude")
	}
	notes := strings.TrimSpace(request.Notes)
	params := db.RecordCargoQRScanParams{
		CargoID:              pgUUID(cargoID),
		EventType:            eventType,
		StationID:            pgUUIDPtr(request.StationID),
		ScannedByPersonnelID: pgUUIDPtr(request.ScannedByPersonnelID),
		Latitude:             request.Latitude,
		Longitude:            request.Longitude,
	}
	if notes != "" {
		params.Notes = &notes
	}
	scan, err := s.queries.RecordCargoQRScan(ctx, params)
	if err != nil {
		return RecordCargoQRScanResponse{}, fmt.Errorf("record cargo QR scan: %w", err)
	}
	if status != "" {
		if _, err := s.queries.UpdateCargoStatusFromQR(ctx, db.UpdateCargoStatusFromQRParams{CargoID: pgUUID(cargoID), Status: status}); err != nil {
			return RecordCargoQRScanResponse{}, fmt.Errorf("update cargo status from QR scan: %w", err)
		}
	} else {
		status = cargo.Status
	}
	return RecordCargoQRScanResponse{
		Status: true, Message: "cargo QR scan recorded", ScanID: uuidFromPg(scan.ID).String(),
		CargoID: cargoID.String(), CargoStatus: status, ScannedAt: timeFromPg(scan.ScannedAt),
	}, nil
}

func validBatchStatus(value string) bool {
	switch value {
	case "draft", "planned", "packed", "dispatched", "received", "delayed", "cancelled":
		return true
	}
	return false
}
func validCargoStatus(value string) bool {
	switch value {
	case "draft", "packed", "dispatched", "in_transit", "received", "delayed", "cancelled":
		return true
	}
	return false
}
func validCargoScanEvent(value string) bool {
	switch value {
	case "created", "packed", "dispatched", "scanned", "received", "damaged", "missing":
		return true
	}
	return false
}
func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }
func pgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}
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
func timePtrFromPg(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
