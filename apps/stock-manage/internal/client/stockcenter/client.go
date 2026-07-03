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
	"strconv"
	"time"
)

// ErrMaterialNotFound indicates the material does not exist in stock-center.
var ErrMaterialNotFound = errors.New("material not found")

// ErrBatchNotFound indicates the batch does not exist in stock-center.
var ErrBatchNotFound = errors.New("batch not found")

// ErrStockBalanceNotFound indicates the stock balance does not exist in stock-center.
var ErrStockBalanceNotFound = errors.New("stock balance not found")

// Client is an HTTP client for stock-center APIs.
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

// Material is the master data record for stainless steel items.
type Material struct {
	ID           uint64    `json:"id"`
	MaterialCode string    `json:"materialCode"`
	Grade        string    `json:"grade"`
	Form         string    `json:"form"`
	Spec         string    `json:"spec"`
	PrimaryUnit  string    `json:"primaryUnit"`
	MaterialType string    `json:"materialType"`
	Status       string    `json:"status"`
	OrgID        uint64    `json:"orgId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// MaterialBatch tracks heat/lot numbers per material.
type MaterialBatch struct {
	ID         uint64    `json:"id"`
	MaterialID uint64    `json:"materialId"`
	HeatNo     string    `json:"heatNo"`
	Remark     string    `json:"remark"`
	OrgID      uint64    `json:"orgId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// StockBalance holds dual-measure inventory for material + batch + warehouse.
type StockBalance struct {
	ID         uint64    `json:"id"`
	MaterialID uint64    `json:"materialId"`
	BatchID    uint64    `json:"batchId"`
	Warehouse  string    `json:"warehouse"`
	WeightKg   string    `json:"weightKg"`
	Quantity   string    `json:"quantity"`
	OrgID      uint64    `json:"orgId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// StockLedger is an immutable inventory movement record.
type StockLedger struct {
	ID            uint64    `json:"id"`
	MaterialID    uint64    `json:"materialId"`
	BatchID       uint64    `json:"batchId"`
	Warehouse     string    `json:"warehouse"`
	DeltaWeightKg string    `json:"deltaWeightKg"`
	DeltaQuantity string    `json:"deltaQuantity"`
	RefType       string    `json:"refType"`
	RefNo         string    `json:"refNo"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"createdAt"`
}

// MaterialListData is the paginated material list payload.
type MaterialListData struct {
	List     []Material `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// BatchListData is the paginated batch list payload.
type BatchListData struct {
	List     []MaterialBatch `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// StockBalanceListData is the paginated stock balance list payload.
type StockBalanceListData struct {
	List     []StockBalance `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// StockLedgerListData is the paginated ledger list payload.
type StockLedgerListData struct {
	List     []StockLedger `json:"list"`
	Total    int64         `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// CreateMaterialInput is the payload for creating a material.
type CreateMaterialInput struct {
	MaterialCode string  `json:"materialCode"`
	Grade        string  `json:"grade"`
	Form         string  `json:"form"`
	Spec         string  `json:"spec"`
	PrimaryUnit  string  `json:"primaryUnit"`
	MaterialType string  `json:"materialType"`
	Status       *string `json:"status,omitempty"`
}

// UpdateMaterialInput is the payload for updating a material.
type UpdateMaterialInput struct {
	Grade        string `json:"grade"`
	Form         string `json:"form"`
	Spec         string `json:"spec"`
	PrimaryUnit  string `json:"primaryUnit"`
	MaterialType string `json:"materialType"`
	Status       string `json:"status"`
}

// CreateBatchInput is the payload for creating a batch.
type CreateBatchInput struct {
	MaterialID uint64 `json:"materialId"`
	HeatNo     string `json:"heatNo"`
	Remark     string `json:"remark"`
}

// UpdateBatchInput is the payload for updating a batch.
type UpdateBatchInput struct {
	HeatNo string `json:"heatNo"`
	Remark string `json:"remark"`
}

// StockMovementInput is the payload for inbound and outbound stock operations.
type StockMovementInput struct {
	MaterialID uint64 `json:"materialId"`
	BatchID    uint64 `json:"batchId"`
	Warehouse  string `json:"warehouse"`
	WeightKg   string `json:"weightKg"`
	Quantity   string `json:"quantity"`
	RefType    string `json:"refType"`
	RefNo      string `json:"refNo"`
	Remark     string `json:"remark"`
}

// ListMaterials queries paginated materials from stock-center.
func (c *Client) ListMaterials(ctx context.Context, page, pageSize int, query url.Values) (*MaterialListData, error) {
	endpoint := c.listURL("/api/v1/materials", page, pageSize, query)
	var data MaterialListData
	if err := c.doGet(ctx, endpoint, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetMaterial returns a material by ID.
func (c *Client) GetMaterial(ctx context.Context, id uint64) (*Material, error) {
	endpoint := fmt.Sprintf("%s/api/v1/materials/%d", c.baseURL, id)
	var material Material
	if err := c.doGetWithNotFound(ctx, endpoint, 40400, ErrMaterialNotFound, &material); err != nil {
		return nil, err
	}
	return &material, nil
}

// CreateMaterial creates a material via stock-center.
func (c *Client) CreateMaterial(ctx context.Context, input CreateMaterialInput) (*Material, error) {
	endpoint := fmt.Sprintf("%s/api/v1/materials", c.baseURL)
	var material Material
	if err := c.doJSON(ctx, http.MethodPost, endpoint, input, &material); err != nil {
		return nil, err
	}
	return &material, nil
}

// UpdateMaterial updates a material via stock-center.
func (c *Client) UpdateMaterial(ctx context.Context, id uint64, input UpdateMaterialInput) (*Material, error) {
	endpoint := fmt.Sprintf("%s/api/v1/materials/%d", c.baseURL, id)
	var material Material
	if err := c.doJSONWithNotFound(ctx, http.MethodPut, endpoint, 40400, ErrMaterialNotFound, input, &material); err != nil {
		return nil, err
	}
	return &material, nil
}

// ListBatches queries paginated batches from stock-center.
func (c *Client) ListBatches(ctx context.Context, page, pageSize int, query url.Values) (*BatchListData, error) {
	endpoint := c.listURL("/api/v1/batches", page, pageSize, query)
	var data BatchListData
	if err := c.doGet(ctx, endpoint, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetBatch returns a batch by ID.
func (c *Client) GetBatch(ctx context.Context, id uint64) (*MaterialBatch, error) {
	endpoint := fmt.Sprintf("%s/api/v1/batches/%d", c.baseURL, id)
	var batch MaterialBatch
	if err := c.doGetWithNotFound(ctx, endpoint, 40400, ErrBatchNotFound, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

// CreateBatch creates a batch via stock-center.
func (c *Client) CreateBatch(ctx context.Context, input CreateBatchInput) (*MaterialBatch, error) {
	endpoint := fmt.Sprintf("%s/api/v1/batches", c.baseURL)
	var batch MaterialBatch
	if err := c.doJSON(ctx, http.MethodPost, endpoint, input, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

// UpdateBatch updates a batch via stock-center.
func (c *Client) UpdateBatch(ctx context.Context, id uint64, input UpdateBatchInput) (*MaterialBatch, error) {
	endpoint := fmt.Sprintf("%s/api/v1/batches/%d", c.baseURL, id)
	var batch MaterialBatch
	if err := c.doJSONWithNotFound(ctx, http.MethodPut, endpoint, 40400, ErrBatchNotFound, input, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

// ListStocks queries paginated stock balances from stock-center.
func (c *Client) ListStocks(ctx context.Context, page, pageSize int, query url.Values) (*StockBalanceListData, error) {
	endpoint := c.listURL("/api/v1/stocks", page, pageSize, query)
	var data StockBalanceListData
	if err := c.doGet(ctx, endpoint, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// QueryStock returns a balance for material, batch, and warehouse.
func (c *Client) QueryStock(ctx context.Context, materialID, batchID uint64, warehouse string) (*StockBalance, error) {
	q := url.Values{}
	q.Set("materialId", strconv.FormatUint(materialID, 10))
	q.Set("batchId", strconv.FormatUint(batchID, 10))
	if warehouse != "" {
		q.Set("warehouse", warehouse)
	}
	endpoint := fmt.Sprintf("%s/api/v1/stocks/query?%s", c.baseURL, q.Encode())

	var balance StockBalance
	if err := c.doGetWithNotFound(ctx, endpoint, 40400, ErrStockBalanceNotFound, &balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// InboundStock increases stock balance via stock-center.
func (c *Client) InboundStock(ctx context.Context, input StockMovementInput) (*StockBalance, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks/inbound", c.baseURL)
	var balance StockBalance
	if err := c.doJSON(ctx, http.MethodPost, endpoint, input, &balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// OutboundStock decreases stock balance via stock-center.
func (c *Client) OutboundStock(ctx context.Context, input StockMovementInput) (*StockBalance, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stocks/outbound", c.baseURL)
	var balance StockBalance
	if err := c.doJSON(ctx, http.MethodPost, endpoint, input, &balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// ListLedger queries paginated ledger entries from stock-center.
func (c *Client) ListLedger(ctx context.Context, page, pageSize int, query url.Values) (*StockLedgerListData, error) {
	endpoint := c.listURL("/api/v1/ledger", page, pageSize, query)
	var data StockLedgerListData
	if err := c.doGet(ctx, endpoint, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Client) listURL(path string, page, pageSize int, query url.Values) string {
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	for key, values := range query {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	return fmt.Sprintf("%s%s?%s", c.baseURL, path, q.Encode())
}

func (c *Client) doGet(ctx context.Context, endpoint string, out interface{}) error {
	return c.doRequest(ctx, http.MethodGet, endpoint, nil, 0, nil, out)
}

func (c *Client) doGetWithNotFound(ctx context.Context, endpoint string, notFoundCode int, notFoundErr error, out interface{}) error {
	return c.doRequest(ctx, http.MethodGet, endpoint, nil, notFoundCode, notFoundErr, out)
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, payload, out interface{}) error {
	return c.doRequest(ctx, method, endpoint, payload, 0, nil, out)
}

func (c *Client) doJSONWithNotFound(ctx context.Context, method, endpoint string, notFoundCode int, notFoundErr error, payload, out interface{}) error {
	return c.doRequest(ctx, method, endpoint, payload, notFoundCode, notFoundErr, out)
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, payload interface{}, notFoundCode int, notFoundErr error, out interface{}) error {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return err
	}
	if apiResp.Code != 0 {
		if notFoundErr != nil && (resp.StatusCode == http.StatusNotFound || apiResp.Code == notFoundCode) {
			return notFoundErr
		}
		return fmt.Errorf("stock-center error: %s", apiResp.Message)
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(apiResp.Data, out)
}

// Forward proxies an HTTP request to stock-center and returns the raw response.
func (c *Client) Forward(ctx context.Context, method, path, rawQuery string, body []byte) (statusCode int, respBody []byte, err error) {
	endpoint := c.baseURL + path
	if rawQuery != "" {
		endpoint += "?" + rawQuery
	}

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, respBody, nil
}
