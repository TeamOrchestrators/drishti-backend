package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const analyticsTimeout = 3 * time.Second

type analyticsClient struct {
	baseURL string
	client  *http.Client
}
type analyticsTransaction struct {
	TransactionType string  `json:"transaction_type"`
	QuantityDelta   float64 `json:"quantity_delta"`
	OccurredAt      string  `json:"occurred_at"`
}

// analyticsByInventoryID contains results returned by the service-wide endpoints.
// Those endpoints may return either one object or a list of inventory objects.
type analyticsByInventoryID map[string]*ItemAnalyticsResponse

func newAnalyticsClient(baseURL string) *analyticsClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &analyticsClient{baseURL: baseURL, client: &http.Client{Timeout: analyticsTimeout}}
}
func (c *analyticsClient) alerts(ctx context.Context) ([]InventoryAlertResponse, error) {
	var result []InventoryAlertResponse
	return result, c.get(ctx, "/inventory/alerts", &result)
}

func (c *analyticsClient) stockDays(ctx context.Context, inventoryID string, currentQuantity float64, transactions []analyticsTransaction) (*float64, error) {
	var stock struct {
		Days *float64 `json:"days_of_stock_remaining"`
	}
	if err := c.post(ctx, "/inventory/stock-days", map[string]any{"inventory_id": inventoryID, "current_quantity": currentQuantity, "inventory_transactions": transactions}, &stock); err != nil {
		return nil, err
	}
	return stock.Days, nil
}

func (c *analyticsClient) sharedAnalytics(ctx context.Context) (analyticsByInventoryID, error) {
	results := analyticsByInventoryID{}
	successful := 0
	for _, endpoint := range []string{"/inventory/depletion-date", "/inventory/anomalies", "/inventory/consumption/average"} {
		var payload json.RawMessage
		if err := c.get(ctx, endpoint, &payload); err != nil {
			continue
		}
		successful++
		mergeAnalyticsPayload(results, payload)
	}
	if successful == 0 {
		return results, fmt.Errorf("all shared analytics requests failed")
	}
	return results, nil
}

func mergeAnalyticsPayload(results analyticsByInventoryID, payload json.RawMessage) {
	var value any
	if json.Unmarshal(payload, &value) != nil {
		return
	}
	visitAnalyticsValue(results, value)
}

func visitAnalyticsValue(results analyticsByInventoryID, value any) {
	switch item := value.(type) {
	case []any:
		for _, child := range item {
			visitAnalyticsValue(results, child)
		}
	case map[string]any:
		if inventoryID, ok := item["inventory_id"].(string); ok && inventoryID != "" {
			analytics := results[inventoryID]
			if analytics == nil {
				analytics = &ItemAnalyticsResponse{}
				results[inventoryID] = analytics
			}
			if date, ok := item["predicted_depletion_date"].(string); ok {
				analytics.PredictedDepletionDate = &date
			}
			if detected, ok := item["anomaly_detected"].(bool); ok {
				analytics.AnomalyDetected = detected
			}
			if message, ok := item["message"].(string); ok {
				analytics.AnomalyMessage = message
			}
			if average, ok := number(item["average_daily_consumption"]); ok {
				analytics.AverageDailyConsumption = average
			}
		}
		for _, child := range item {
			visitAnalyticsValue(results, child)
		}
	}
}

func number(value any) (float64, bool) {
	result, ok := value.(float64)
	return result, ok
}

func (c *analyticsClient) get(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, target)
}
func (c *analyticsClient) post(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, target)
}
func (c *analyticsClient) do(req *http.Request, target any) error {
	req.Header.Set("ngrok-skip-browser-warning", "1")
	response, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("analytics service returned %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
