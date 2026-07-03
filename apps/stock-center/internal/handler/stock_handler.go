package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"github.com/tserom/l-project/apps/stock-center/internal/service"
	"github.com/tserom/l-project/apps/stock-center/pkg/response"
)

// StockHandler handles stock-related HTTP requests.
type StockHandler struct {
	svc *service.StockService
}

// NewStockHandler creates a StockHandler.
func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

// List handles GET /api/v1/stocks.
func (h *StockHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	stocks, total, err := h.svc.ListStocks(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     stocks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/stocks/:id.
func (h *StockHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	stock, err := h.svc.GetStock(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrStockNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "stock not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, stock)
}

// GetBySKU handles GET /api/v1/stocks/by-sku.
func (h *StockHandler) GetBySKU(c *gin.Context) {
	sku := c.Query("sku")
	warehouse := c.Query("warehouse")
	if sku == "" || warehouse == "" {
		response.Fail(c, http.StatusBadRequest, 40000, "sku and warehouse are required")
		return
	}

	stock, err := h.svc.GetStockBySKU(c.Request.Context(), sku, warehouse)
	if err != nil {
		if errors.Is(err, repository.ErrStockNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "stock not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, stock)
}

// Create handles POST /api/v1/stocks.
func (h *StockHandler) Create(c *gin.Context) {
	var input service.CreateStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	stock, err := h.svc.CreateStock(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, stock)
}

// Adjust handles PUT /api/v1/stocks/:id/quantity.
func (h *StockHandler) Adjust(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.AdjustStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	stock, err := h.svc.AdjustStock(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrStockNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "stock not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, stock)
}
