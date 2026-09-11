package inventory

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries   *db.Queries
	analytics *analyticsClient
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries, analytics: newAnalyticsClient(os.Getenv("INVENTORY_MICROSERVICE_URL"))}
}

func (s *Service) List(ctx context.Context, stationID uuid.UUID) (StationInventoryResponse, error) {
	items, err := s.queries.ListStationInventory(ctx, pgUUID(stationID))
	if err != nil {
		return StationInventoryResponse{}, fmt.Errorf("list station inventory: %w", err)
	}
	history, err := s.queries.ListStationInventoryHistory(ctx, db.ListStationInventoryHistoryParams{StationID: pgUUID(stationID), HistoryLimit: 20})
	if err != nil {
		return StationInventoryResponse{}, fmt.Errorf("list inventory history: %w", err)
	}
	response := StationInventoryResponse{Items: make([]ItemResponse, 0, len(items)), History: make([]HistoryResponse, 0, len(history)), Alerts: []InventoryAlertResponse{}}
	for _, item := range items {
		var maximum *float64
		if item.MaximumQuantity > 0 {
			value := item.MaximumQuantity
			maximum = &value
		}
		response.Items = append(response.Items, ItemResponse{InventoryID: uuidFromPg(item.ID).String(), ItemID: uuidFromPg(item.ItemID).String(), ItemCode: item.ItemCode, Name: item.Name, Category: item.Category, Unit: item.Unit, IsCritical: item.IsCritical, OnHandQuantity: item.OnHandQuantity, ReservedQuantity: item.ReservedQuantity, AvailableQuantity: item.AvailableQuantity, MinimumQuantity: item.MinimumQuantity, MaximumQuantity: maximum, StockLevel: stockLevel(item.AvailableQuantity, item.MinimumQuantity, item.IsCritical)})
	}
	for _, entry := range history {
		response.History = append(response.History, HistoryResponse{ID: uuidFromPg(entry.ID).String(), ItemName: entry.ItemName, Unit: entry.Unit, TransactionType: entry.TransactionType, OnHandDelta: entry.OnHandDelta, ReservedDelta: entry.ReservedDelta, OccurredAt: timeFromPg(entry.OccurredAt), Notes: entry.Notes})
	}
	if s.analytics != nil {
		alerts, err := s.analytics.alerts(ctx)
		if err == nil {
			response.AnalyticsAvailable = true
			inventoryIDs := make(map[string]struct{}, len(response.Items))
			for _, item := range response.Items {
				inventoryIDs[item.InventoryID] = struct{}{}
			}
			for _, alert := range alerts {
				if _, found := inventoryIDs[alert.InventoryID]; found {
					response.Alerts = append(response.Alerts, alert)
				}
			}
		}
		sharedAnalytics, err := s.analytics.sharedAnalytics(ctx)
		if err == nil {
			response.AnalyticsAvailable = true
		}
		for index := range response.Items {
			inventoryID, err := uuid.Parse(response.Items[index].InventoryID)
			if err != nil {
				continue
			}
			transactions, err := s.queries.ListInventoryTransactionsForAnalytics(ctx, pgUUID(inventoryID))
			if err != nil {
				continue
			}
			payload := make([]analyticsTransaction, 0, len(transactions))
			for _, transaction := range transactions {
				payload = append(payload, analyticsTransaction{TransactionType: transaction.TransactionType, QuantityDelta: transaction.QuantityDelta, OccurredAt: timeFromPg(transaction.OccurredAt).Format(time.RFC3339)})
			}
			analytics := sharedAnalytics[response.Items[index].InventoryID]
			if analytics == nil {
				analytics = &ItemAnalyticsResponse{}
			}
			days, err := s.analytics.stockDays(ctx, response.Items[index].InventoryID, response.Items[index].AvailableQuantity, payload)
			if err == nil {
				analytics.DaysOfStockRemaining = days
				response.AnalyticsAvailable = true
			}
			if analytics.DaysOfStockRemaining != nil || analytics.PredictedDepletionDate != nil || analytics.AnomalyMessage != "" || analytics.AverageDailyConsumption != 0 {
				response.Items[index].Analytics = analytics
			}
		}
	}
	return response, nil
}

func (s *Service) CreateItem(ctx context.Context, stationID uuid.UUID, request CreateItemRequest) (MutationResponse, error) {
	if strings.TrimSpace(request.ItemCode) == "" || strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Category) == "" || strings.TrimSpace(request.Unit) == "" {
		return MutationResponse{}, fmt.Errorf("item_code, name, category, and unit are required")
	}
	if request.OpeningQuantity < 0 || request.MinimumQuantity < 0 || (request.MaximumQuantity != nil && *request.MaximumQuantity < request.MinimumQuantity) {
		return MutationResponse{}, fmt.Errorf("invalid inventory quantities")
	}
	params := db.CreateStationInventoryItemParams{ItemCode: strings.TrimSpace(request.ItemCode), Name: strings.TrimSpace(request.Name), Category: strings.TrimSpace(request.Category), Unit: strings.TrimSpace(request.Unit), IsCritical: request.IsCritical, StationID: pgUUID(stationID), OpeningQuantity: numeric(request.OpeningQuantity), MinimumQuantity: numeric(request.MinimumQuantity)}
	if request.MaximumQuantity != nil {
		params.MaximumQuantity = numeric(*request.MaximumQuantity)
	}
	if notes := strings.TrimSpace(request.Notes); notes != "" {
		params.Notes = &notes
	}
	if _, err := s.queries.CreateStationInventoryItem(ctx, params); err != nil {
		return MutationResponse{}, fmt.Errorf("create inventory item: %w", err)
	}
	return MutationResponse{Status: true, Message: "inventory item created"}, nil
}

func (s *Service) AdjustStock(ctx context.Context, stationID, inventoryID uuid.UUID, request StockAdjustmentRequest) (MutationResponse, error) {
	if request.Quantity <= 0 {
		return MutationResponse{}, fmt.Errorf("quantity must be greater than zero")
	}
	delta, transactionType := request.Quantity, "receipt"
	if request.Operation == "remove" {
		delta, transactionType = -request.Quantity, "consumption"
	} else if request.Operation != "add" {
		return MutationResponse{}, fmt.Errorf("operation must be add or remove")
	}
	params := db.AdjustStationInventoryStockParams{OnHandDelta: numeric(delta), StationInventoryID: pgUUID(inventoryID), StationID: pgUUID(stationID), TransactionType: transactionType}
	if notes := strings.TrimSpace(request.Notes); notes != "" {
		params.Notes = &notes
	}
	if _, err := s.queries.AdjustStationInventoryStock(ctx, params); err != nil {
		return MutationResponse{}, fmt.Errorf("adjust inventory stock: %w", err)
	}
	return MutationResponse{Status: true, Message: "inventory stock updated"}, nil
}

func stockLevel(available, minimum float64, critical bool) string {
	if available < minimum {
		if critical {
			return "critical"
		}
		return "low"
	}
	return "normal"
}
func numeric(value float64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(int64(math.Round(value * 1000))), Exp: -3, Valid: true}
}
func pgUUID(value uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: value, Valid: true} }
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
