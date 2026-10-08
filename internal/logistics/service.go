package logistics

import (
	"context"
	"encoding/json"
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
		cargoResponse := CargoResponse{ID: uuidFromPg(row.ID).String(), CargoCode: row.CargoCode, OriginStationName: origin, DestinationStationName: row.DestinationStationName, ExpeditionName: row.ExpeditionName, LogisticsBatchCode: row.LogisticsBatchCode, LogisticsShipmentCode: row.LogisticsShipmentCode, TransportMode: row.TransportMode, Priority: row.Priority, Status: row.Status, Notes: row.Notes}
		if row.LogisticsShipmentID.Valid {
			id := uuidFromPg(row.LogisticsShipmentID).String()
			cargoResponse.LogisticsShipmentID = &id
		}
		response.Cargo = append(response.Cargo, cargoResponse)
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
	expeditionWindow, err := s.queries.GetLogisticsExpeditionWindow(ctx, pgUUID(request.ExpeditionID))
	if err != nil {
		return MutationResponse{}, fmt.Errorf("get expedition schedule: %w", err)
	}
	if expeditionWindow.PlannedStartAt.Valid && request.PlannedDispatchAt.Before(expeditionWindow.PlannedStartAt.Time) {
		return MutationResponse{}, fmt.Errorf("planned_dispatch_at must not be before the expedition start")
	}
	if expeditionWindow.PlannedEndAt.Valid && request.EstimatedArrivalAt.After(expeditionWindow.PlannedEndAt.Time) {
		return MutationResponse{}, fmt.Errorf("estimated_arrival_at must not be after the expedition end")
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
	batch, err := s.queries.CreateLogisticsBatch(ctx, params)
	if err != nil {
		return MutationResponse{}, fmt.Errorf("create logistics batch: %w", err)
	}
	if _, err := s.queries.CreateDefaultLogisticsShipment(ctx, batch.ID); err != nil {
		return MutationResponse{}, fmt.Errorf("create default shipment: %w", err)
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

func (s *Service) CreateShipment(ctx context.Context, batchID uuid.UUID, request CreateShipmentRequest) (MutationResponse, error) {
	if batchID == uuid.Nil || strings.TrimSpace(request.ShipmentCode) == "" || strings.TrimSpace(request.TransportMode) == "" {
		return MutationResponse{}, fmt.Errorf("batch id, shipment_code, and transport_mode are required")
	}
	if !validTransportMode(request.TransportMode) || (request.Priority != "" && !validCargoPriority(request.Priority)) || (request.Status != "" && !validBatchStatus(request.Status)) {
		return MutationResponse{}, fmt.Errorf("invalid shipment transport_mode, priority, or status")
	}
	if request.PlannedDispatchAt.IsZero() || request.EstimatedArrivalAt.IsZero() || request.EstimatedArrivalAt.Before(request.PlannedDispatchAt) {
		return MutationResponse{}, fmt.Errorf("estimated_arrival_at must not be before planned_dispatch_at")
	}
	// A split route may leave before the parent batch's planned dispatch, but
	// must still fit inside the parent expedition's operational window.
	expeditionWindow, err := s.queries.GetLogisticsShipmentWindow(ctx, pgUUID(batchID))
	if err != nil {
		return MutationResponse{}, fmt.Errorf("get parent expedition schedule: %w", err)
	}
	if expeditionWindow.PlannedStartAt.Valid && request.PlannedDispatchAt.Before(expeditionWindow.PlannedStartAt.Time) {
		return MutationResponse{}, fmt.Errorf("planned_dispatch_at must not be before the expedition start")
	}
	if expeditionWindow.PlannedEndAt.Valid && request.EstimatedArrivalAt.After(expeditionWindow.PlannedEndAt.Time) {
		return MutationResponse{}, fmt.Errorf("estimated_arrival_at must not be after the expedition end")
	}
	priority, status := request.Priority, request.Status
	if priority == "" {
		priority = "standard"
	}
	if status == "" {
		status = "planned"
	}
	params := db.CreateLogisticsShipmentParams{ShipmentCode: strings.TrimSpace(request.ShipmentCode), LogisticsBatchID: pgUUID(batchID), TransportMode: request.TransportMode, Priority: priority, Status: status, PlannedDispatchAt: pgTime(request.PlannedDispatchAt), EstimatedArrivalAt: pgTime(request.EstimatedArrivalAt)}
	if notes := strings.TrimSpace(request.Notes); notes != "" {
		params.Notes = &notes
	}
	if _, err := s.queries.CreateLogisticsShipment(ctx, params); err != nil {
		return MutationResponse{}, fmt.Errorf("create logistics shipment: %w", err)
	}
	return MutationResponse{Status: true, Message: "logistics shipment created"}, nil
}

func (s *Service) AssignCargoToShipment(ctx context.Context, cargoID uuid.UUID, shipmentID *uuid.UUID) (MutationResponse, error) {
	if cargoID == uuid.Nil {
		return MutationResponse{}, fmt.Errorf("cargo_id is required")
	}
	if shipmentID == nil || *shipmentID == uuid.Nil {
		if _, err := s.queries.ClearCargoLogisticsShipment(ctx, pgUUID(cargoID)); err != nil {
			return MutationResponse{}, fmt.Errorf("clear cargo shipment: %w", err)
		}
		return MutationResponse{Status: true, Message: "cargo removed from logistics shipment"}, nil
	}
	if _, err := s.queries.AssignCargoToLogisticsShipment(ctx, db.AssignCargoToLogisticsShipmentParams{CargoID: pgUUID(cargoID), LogisticsShipmentID: pgUUID(*shipmentID)}); err != nil {
		return MutationResponse{}, fmt.Errorf("assign cargo to shipment: %w", err)
	}
	return MutationResponse{Status: true, Message: "cargo assigned to logistics shipment"}, nil
}

func (s *Service) SplitCargo(ctx context.Context, sourceCargoID uuid.UUID, request SplitCargoRequest) (SplitCargoResponse, error) {
	if sourceCargoID == uuid.Nil || strings.TrimSpace(request.NewCargoCode) == "" || len(request.Items) == 0 {
		return SplitCargoResponse{}, fmt.Errorf("source cargo id, new_cargo_code, and at least one item are required")
	}
	seen := make(map[uuid.UUID]struct{}, len(request.Items))
	for _, item := range request.Items {
		if item.ItemID == uuid.Nil || item.Quantity <= 0 {
			return SplitCargoResponse{}, fmt.Errorf("every split item requires an item_id and positive quantity")
		}
		if _, exists := seen[item.ItemID]; exists {
			return SplitCargoResponse{}, fmt.Errorf("split items cannot contain duplicate item_id values")
		}
		seen[item.ItemID] = struct{}{}
	}
	priority := request.Priority
	if priority == "" {
		priority = "standard"
	}
	if !validCargoPriority(priority) {
		return SplitCargoResponse{}, fmt.Errorf("invalid cargo priority")
	}
	itemsJSON, err := json.Marshal(request.Items)
	if err != nil {
		return SplitCargoResponse{}, fmt.Errorf("encode split items: %w", err)
	}
	params := db.SplitCargoManifestParams{SourceCargoID: pgUUID(sourceCargoID), ItemsJson: itemsJSON, NewCargoCode: strings.TrimSpace(request.NewCargoCode), Priority: priority, LogisticsShipmentID: pgUUIDPtr(request.ShipmentID)}
	if notes := strings.TrimSpace(request.Notes); notes != "" {
		params.Notes = &notes
	}
	childID, err := s.queries.SplitCargoManifest(ctx, params)
	if err != nil {
		return SplitCargoResponse{}, fmt.Errorf("split cargo manifest: %w", err)
	}
	if !childID.Valid {
		return SplitCargoResponse{}, fmt.Errorf("split cargo manifest: requested quantities or shipment are not valid")
	}
	id := uuidFromPg(childID).String()
	return SplitCargoResponse{Status: true, Message: "cargo manifest split", CargoID: id, QRToken: id, SourceCargoID: sourceCargoID.String()}, nil
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
		LogisticsShipmentCode:  cargo.LogisticsShipmentCode,
		TransportMode:          cargo.TransportMode,
		Priority:               cargo.Priority,
		Status:                 cargo.Status,
		Notes:                  cargo.Notes,
		DispatchedAt:           timePtrFromPg(cargo.DispatchedAt),
		ReceivedAt:             timePtrFromPg(cargo.ReceivedAt),
		CreatedAt:              timeFromPg(cargo.CreatedAt),
		Items:                  make([]CargoQRItemResponse, 0, len(items)),
		ScanHistory:            make([]CargoQRScanResponse, 0, len(scans)),
	}
	if cargo.LogisticsBatchID.Valid {
		batchID := uuidFromPg(cargo.LogisticsBatchID).String()
		response.LogisticsBatchID = &batchID
	}
	if cargo.LogisticsShipmentID.Valid {
		shipmentID := uuidFromPg(cargo.LogisticsShipmentID).String()
		response.LogisticsShipmentID = &shipmentID
	}
	if cargo.SourceCargoID.Valid {
		sourceID := uuidFromPg(cargo.SourceCargoID).String()
		response.SourceCargoID = &sourceID
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
	if status == "" {
		status = cargoStatusForScanEvent(eventType)
	}
	if request.Latitude == nil || request.Longitude == nil {
		return RecordCargoQRScanResponse{}, fmt.Errorf("latitude and longitude are required for a cargo checkpoint")
	}
	if *request.Latitude < -90 || *request.Latitude > 90 || *request.Longitude < -180 || *request.Longitude > 180 {
		return RecordCargoQRScanResponse{}, fmt.Errorf("invalid latitude or longitude")
	}
	scope := strings.TrimSpace(request.Scope)
	if scope == "" {
		if request.ApplyToBatch {
			scope = "batch"
		} else {
			scope = "cargo"
		}
	}
	if scope != "cargo" && scope != "shipment" && scope != "batch" {
		return RecordCargoQRScanResponse{}, fmt.Errorf("scope must be cargo, shipment, or batch")
	}
	if scope == "batch" && !cargo.LogisticsBatchID.Valid {
		return RecordCargoQRScanResponse{}, fmt.Errorf("cargo is not assigned to a logistics batch")
	}
	if scope == "shipment" && !cargo.LogisticsShipmentID.Valid {
		return RecordCargoQRScanResponse{}, fmt.Errorf("cargo is not assigned to a logistics shipment")
	}
	notes := strings.TrimSpace(request.Notes)
	params := db.RecordCargoQRCheckpointParams{
		CargoID:              pgUUID(cargoID),
		PropagationScope:     scope,
		EventType:            eventType,
		StationID:            pgUUIDPtr(request.StationID),
		ScannedByPersonnelID: pgUUIDPtr(request.ScannedByPersonnelID),
		Latitude:             *request.Latitude,
		Longitude:            *request.Longitude,
	}
	if status != "" {
		params.Status = &status
	}
	if notes != "" {
		params.Notes = &notes
	}
	scan, err := s.queries.RecordCargoQRCheckpoint(ctx, params)
	if err != nil {
		return RecordCargoQRScanResponse{}, fmt.Errorf("record cargo checkpoint: %w", err)
	}
	responseStatus := status
	if responseStatus == "" || cargo.Status == "received" || cargo.Status == "cancelled" {
		responseStatus = cargo.Status
	}
	response := RecordCargoQRScanResponse{
		Status: true, Message: "cargo QR scan recorded", ScanID: uuidFromPg(scan.ScanID).String(),
		CargoID: cargoID.String(), CargoStatus: responseStatus, ScannedAt: timeFromPg(scan.ScannedAt),
		AppliedToBatch: scope == "batch", PropagationScope: scope, AffectedCargoCount: int(scan.AffectedCargoCount),
		StatusUpdatedCount: int(scan.StatusUpdatedCount), LogisticsBatchCode: cargo.LogisticsBatchCode,
	}
	if cargo.LogisticsBatchID.Valid {
		batchID := uuidFromPg(cargo.LogisticsBatchID).String()
		response.LogisticsBatchID = &batchID
	}
	return response, nil
}

func (s *Service) GetBatchTracking(ctx context.Context, batchID uuid.UUID) (LogisticsBatchTrackingResponse, error) {
	batch, err := s.queries.GetLogisticsBatchTracking(ctx, pgUUID(batchID))
	if err != nil {
		return LogisticsBatchTrackingResponse{}, fmt.Errorf("get logistics batch tracking: %w", err)
	}
	points, err := s.queries.ListLogisticsBatchCheckpointPoints(ctx, pgUUID(batchID))
	if err != nil {
		return LogisticsBatchTrackingResponse{}, fmt.Errorf("list logistics batch checkpoints: %w", err)
	}
	shipments, err := s.queries.ListLogisticsShipmentsForBatch(ctx, pgUUID(batchID))
	if err != nil {
		return LogisticsBatchTrackingResponse{}, fmt.Errorf("list logistics shipments: %w", err)
	}
	response := LogisticsBatchTrackingResponse{
		ID: uuidFromPg(batch.ID).String(), BatchCode: batch.BatchCode, Status: batch.Status,
		ExpeditionCode: batch.ExpeditionCode, ExpeditionName: batch.ExpeditionName,
		OriginStationName: batch.OriginStationName, DestinationStationName: batch.DestinationStationName,
		CargoCount: int(batch.CargoCount), Checkpoints: make([]LogisticsBatchCheckpointPoint, 0, len(points)), Shipments: make([]LogisticsShipmentTrackingResponse, 0, len(shipments)),
	}
	for _, point := range points {
		response.Checkpoints = append(response.Checkpoints, LogisticsBatchCheckpointPoint{
			ID: uuidFromPg(point.ID).String(), EventType: point.EventType, Latitude: point.Latitude,
			Longitude: point.Longitude, ScannedAt: timeFromPg(point.ScannedAt), Notes: point.Notes,
			AffectedCargoCount: int(point.AffectedCargoCount),
		})
	}
	shipmentStatuses := make(map[string]int, len(shipments))
	for _, shipment := range shipments {
		checkpointRows, err := s.queries.ListLogisticsShipmentCheckpointPoints(ctx, shipment.ID)
		if err != nil {
			return LogisticsBatchTrackingResponse{}, fmt.Errorf("list shipment checkpoints: %w", err)
		}
		shipmentResponse := LogisticsShipmentTrackingResponse{ID: uuidFromPg(shipment.ID).String(), ShipmentCode: shipment.ShipmentCode, TransportMode: shipment.TransportMode, Priority: shipment.Priority, Status: shipment.Status, PlannedDispatchAt: timePtrFromPg(shipment.PlannedDispatchAt), EstimatedArrivalAt: timePtrFromPg(shipment.EstimatedArrivalAt), CargoCount: int(shipment.CargoCount), TotalQuantity: shipment.TotalQuantity, Checkpoints: make([]LogisticsShipmentCheckpointPoint, 0, len(checkpointRows))}
		shipmentStatuses[shipment.Status]++
		for _, point := range checkpointRows {
			shipmentResponse.Checkpoints = append(shipmentResponse.Checkpoints, LogisticsShipmentCheckpointPoint{ID: uuidFromPg(point.ID).String(), CargoID: uuidFromPg(point.CargoID).String(), CargoCode: point.CargoCode, EventType: point.EventType, Latitude: point.Latitude, Longitude: point.Longitude, ScannedAt: timeFromPg(point.ScannedAt), Notes: point.Notes})
		}
		response.Shipments = append(response.Shipments, shipmentResponse)
	}
	response.ProgressStatus = deriveBatchProgress(batch.Status, shipmentStatuses, len(shipments))
	return response, nil
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
func validCargoPriority(value string) bool {
	return value == "standard" || value == "high" || value == "critical"
}
func validTransportMode(value string) bool {
	return value == "ground" || value == "airlift" || value == "sea_lift" || value == "mixed" || value == "other"
}
func deriveBatchProgress(batchStatus string, statuses map[string]int, shipmentCount int) string {
	if shipmentCount == 0 {
		return batchStatus
	}
	if statuses["received"] == shipmentCount {
		return "received"
	}
	if statuses["received"] > 0 {
		return "partially_received"
	}
	if statuses["dispatched"]+statuses["in_transit"] > 0 && statuses["dispatched"]+statuses["in_transit"] < shipmentCount {
		return "partially_dispatched"
	}
	if statuses["dispatched"]+statuses["in_transit"] == shipmentCount {
		return "dispatched"
	}
	return batchStatus
}
func validCargoScanEvent(value string) bool {
	switch value {
	case "created", "packed", "dispatched", "scanned", "received", "damaged", "missing":
		return true
	}
	return false
}
func cargoStatusForScanEvent(eventType string) string {
	switch eventType {
	case "created":
		return "draft"
	case "packed":
		return "packed"
	case "dispatched":
		return "dispatched"
	case "received":
		return "received"
	case "damaged", "missing":
		return "delayed"
	default:
		return ""
	}
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
