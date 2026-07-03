package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"github.com/tserom/l-project/apps/stock-manage/pkg/response"
)

// OutboundOrderHandler handles outbound order HTTP requests.
type OutboundOrderHandler struct {
	svc *service.OutboundOrderService
}

// NewOutboundOrderHandler creates an OutboundOrderHandler.
func NewOutboundOrderHandler(svc *service.OutboundOrderService) *OutboundOrderHandler {
	return &OutboundOrderHandler{svc: svc}
}

// List handles GET /api/v1/outbound-orders.
func (h *OutboundOrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	orders, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     orders,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/outbound-orders/:id.
func (h *OutboundOrderHandler) Get(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOutboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "outbound order not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, order)
}

// Create handles POST /api/v1/outbound-orders.
func (h *OutboundOrderHandler) Create(c *gin.Context) {
	var input service.CreateOutboundOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	order, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// Update handles PUT /api/v1/outbound-orders/:id.
func (h *OutboundOrderHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateOutboundOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	order, err := h.svc.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrOutboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "outbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// Delete handles DELETE /api/v1/outbound-orders/:id.
func (h *OutboundOrderHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrOutboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "outbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, nil)
}

// Confirm handles POST /api/v1/outbound-orders/:id/confirm.
func (h *OutboundOrderHandler) Confirm(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.svc.Confirm(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOutboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "outbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}
