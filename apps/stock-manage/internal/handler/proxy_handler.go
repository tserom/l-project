package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/pkg/response"
)

// ProxyHandler forwards material, batch, stock, and ledger requests to stock-center.
type ProxyHandler struct {
	client *stockcenter.Client
}

// NewProxyHandler creates a ProxyHandler.
func NewProxyHandler(client *stockcenter.Client) *ProxyHandler {
	return &ProxyHandler{client: client}
}

func (h *ProxyHandler) forward(c *gin.Context, path string) {
	var body []byte
	if c.Request.Body != nil {
		var err error
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, 40000, err.Error())
			return
		}
	}

	statusCode, respBody, err := h.client.Forward(
		c.Request.Context(),
		c.Request.Method,
		path,
		c.Request.URL.RawQuery,
		body,
	)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, 50200, err.Error())
		return
	}

	c.Data(statusCode, "application/json", respBody)
}

// ListMaterials handles GET /api/v1/materials.
func (h *ProxyHandler) ListMaterials(c *gin.Context) {
	h.forward(c, "/api/v1/materials")
}

// CreateMaterial handles POST /api/v1/materials.
func (h *ProxyHandler) CreateMaterial(c *gin.Context) {
	h.forward(c, "/api/v1/materials")
}

// GetMaterial handles GET /api/v1/materials/:id.
func (h *ProxyHandler) GetMaterial(c *gin.Context) {
	h.forward(c, "/api/v1/materials/"+c.Param("id"))
}

// UpdateMaterial handles PUT /api/v1/materials/:id.
func (h *ProxyHandler) UpdateMaterial(c *gin.Context) {
	h.forward(c, "/api/v1/materials/"+c.Param("id"))
}

// ListBatches handles GET /api/v1/batches.
func (h *ProxyHandler) ListBatches(c *gin.Context) {
	h.forward(c, "/api/v1/batches")
}

// CreateBatch handles POST /api/v1/batches.
func (h *ProxyHandler) CreateBatch(c *gin.Context) {
	h.forward(c, "/api/v1/batches")
}

// GetBatch handles GET /api/v1/batches/:id.
func (h *ProxyHandler) GetBatch(c *gin.Context) {
	h.forward(c, "/api/v1/batches/"+c.Param("id"))
}

// UpdateBatch handles PUT /api/v1/batches/:id.
func (h *ProxyHandler) UpdateBatch(c *gin.Context) {
	h.forward(c, "/api/v1/batches/"+c.Param("id"))
}

// ListStocks handles GET /api/v1/stocks.
func (h *ProxyHandler) ListStocks(c *gin.Context) {
	h.forward(c, "/api/v1/stocks")
}

// QueryStock handles GET /api/v1/stocks/query.
func (h *ProxyHandler) QueryStock(c *gin.Context) {
	h.forward(c, "/api/v1/stocks/query")
}

// ListLedger handles GET /api/v1/ledger.
func (h *ProxyHandler) ListLedger(c *gin.Context) {
	h.forward(c, "/api/v1/ledger")
}
