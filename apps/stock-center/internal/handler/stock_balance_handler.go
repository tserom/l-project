package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"github.com/tserom/l-project/apps/stock-center/internal/service"
	"github.com/tserom/l-project/apps/stock-center/pkg/response"
)

// StockBalanceHandler handles stock balance and ledger HTTP requests.
type StockBalanceHandler struct {
	svc *service.StockBalanceService
}

// NewStockBalanceHandler creates a StockBalanceHandler.
func NewStockBalanceHandler(svc *service.StockBalanceService) *StockBalanceHandler {
	return &StockBalanceHandler{svc: svc}
}

// List handles GET /api/v1/stocks.
func (h *StockBalanceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	preds, err := qp.StockPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	balances, total, err := h.svc.ListStocks(c.Request.Context(), page, pageSize, preds)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     balances,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Query handles GET /api/v1/stocks/query.
func (h *StockBalanceHandler) Query(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Query("materialId"), 10, 64)
	if err != nil || materialID == 0 {
		response.Fail(c, http.StatusBadRequest, 40000, "materialId is required")
		return
	}

	batchID, err := strconv.ParseUint(c.Query("batchId"), 10, 64)
	if err != nil || batchID == 0 {
		response.Fail(c, http.StatusBadRequest, 40000, "batchId is required")
		return
	}

	warehouse := c.Query("warehouse")

	balance, err := h.svc.QueryStock(c.Request.Context(), materialID, batchID, warehouse)
	if err != nil {
		if errors.Is(err, repository.ErrStockBalanceNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "stock balance not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, balance)
}

// Inbound handles POST /api/v1/stocks/inbound.
func (h *StockBalanceHandler) Inbound(c *gin.Context) {
	var input service.StockMovementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	balance, err := h.svc.Inbound(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, balance)
}

// Outbound handles POST /api/v1/stocks/outbound.
func (h *StockBalanceHandler) Outbound(c *gin.Context) {
	var input service.StockMovementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	balance, err := h.svc.Outbound(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientWeight) || errors.Is(err, repository.ErrInsufficientQuantity) || errors.Is(err, repository.ErrStockBalanceNotFound) {
			response.Fail(c, http.StatusBadRequest, 40000, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, balance)
}

// ListLedger handles GET /api/v1/ledger.
func (h *StockBalanceHandler) ListLedger(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	preds, err := qp.LedgerPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	entries, total, err := h.svc.ListLedger(c.Request.Context(), page, pageSize, preds)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     entries,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
