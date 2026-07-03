package stockcenter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ErrStockNotFound indicates the stock record does not exist in stock-center.
var ErrStockNotFound = errors.New("stock not found")

// Client is an HTTP client for stock-center internal APIs.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a stock-center client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// APIResponse is the unified response envelope from stock-center.
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Stock is the stock record returned by stock-center.
type Stock struct {
	ID        uint64 `json:"id"`
	SKU       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int64  `json:"quantity"`
}

// StockListData is the paginated list payload from stock-center.
type StockListData struct {
	List     []Stock `json:"list"`
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
}

// GetStockBySKU queries stock-center by SKU and warehouse.
func (c *Client) GetStockBySKU(ctx context.Context, sku, warehouse string) (*Stock, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks/by-sku?sku=%s&warehouse=%s",
		c.baseURL,
		url.QueryEscape(sku),
		url.QueryEscape(warehouse),
	)
	return c.getStock(ctx, endpoint)
}

// GetStockByID queries stock-center by primary key.
func (c *Client) GetStockByID(ctx context.Context, id uint64) (*Stock, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks/%d", c.baseURL, id)
	return c.getStock(ctx, endpoint)
}

// ListStocks queries paginated stock records from stock-center.
func (c *Client) ListStocks(ctx context.Context, page, pageSize int) (*StockListData, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks?page=%d&pageSize=%d", c.baseURL, page, pageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("stock-center error: %s", apiResp.Message)
	}

	var data StockListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// CreateStockInput is the payload for creating stock via stock-center.
type CreateStockInput struct {
	SKU       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int64  `json:"quantity"`
}

// CreateStock creates a stock record via stock-center.
func (c *Client) CreateStock(ctx context.Context, input CreateStockInput) (*Stock, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks", c.baseURL)
	return c.postStock(ctx, endpoint, input)
}

// AdjustStockInput is the payload for adjusting stock quantity.
type AdjustStockInput struct {
	Quantity int64 `json:"quantity"`
}

// AdjustStock updates stock quantity via stock-center.
func (c *Client) AdjustStock(ctx context.Context, id uint64, input AdjustStockInput) (*Stock, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks/%d/quantity", c.baseURL, id)
	return c.putStock(ctx, endpoint, input)
}

func (c *Client) getStock(ctx context.Context, endpoint string) (*Stock, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		if resp.StatusCode == http.StatusNotFound || apiResp.Code == 40400 {
			return nil, ErrStockNotFound
		}
		return nil, fmt.Errorf("stock-center error: %s", apiResp.Message)
	}

	var stock Stock
	if err := json.Unmarshal(apiResp.Data, &stock); err != nil {
		return nil, err
	}
	return &stock, nil
}

func (c *Client) postStock(ctx context.Context, endpoint string, payload interface{}) (*Stock, error) {
	return c.writeStock(ctx, http.MethodPost, endpoint, payload)
}

func (c *Client) putStock(ctx context.Context, endpoint string, payload interface{}) (*Stock, error) {
	return c.writeStock(ctx, http.MethodPut, endpoint, payload)
}

func (c *Client) writeStock(ctx context.Context, method, endpoint string, payload interface{}) (*Stock, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("stock-center error: %s", apiResp.Message)
	}

	var stock Stock
	if err := json.Unmarshal(apiResp.Data, &stock); err != nil {
		return nil, err
	}
	return &stock, nil
}
