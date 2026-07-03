package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"github.com/tserom/l-project/apps/stock-manage/pkg/response"
)

// InventoryHandler handles business inventory endpoints.
type InventoryHandler struct {
	svc *service.InventoryService
}

// NewInventoryHandler creates an InventoryHandler.
func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// Query handles GET /api/v1/inventory/query.
func (h *InventoryHandler) Query(c *gin.Context) {
	var input service.QueryInventoryInput
	if err := c.ShouldBindQuery(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	stock, err := h.svc.QueryInventory(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, stock)
}

// List handles GET /api/v1/inventory.
func (h *InventoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	data, err := h.svc.ListInventory(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, 50200, err.Error())
		return
	}

	response.OK(c, data)
}

// Inbound handles POST /api/v1/inventory/inbound.
func (h *InventoryHandler) Inbound(c *gin.Context) {
	var input service.InboundStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	stock, err := h.svc.InboundStock(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, stock)
}
