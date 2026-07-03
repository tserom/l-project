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

// InboundOrderHandler handles inbound order HTTP requests.
type InboundOrderHandler struct {
	svc *service.InboundOrderService
}

// NewInboundOrderHandler creates an InboundOrderHandler.
func NewInboundOrderHandler(svc *service.InboundOrderService) *InboundOrderHandler {
	return &InboundOrderHandler{svc: svc}
}

// List handles GET /api/v1/inbound-orders.
func (h *InboundOrderHandler) List(c *gin.Context) {
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

// Get handles GET /api/v1/inbound-orders/:id.
func (h *InboundOrderHandler) Get(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrInboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "inbound order not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, order)
}

// Create handles POST /api/v1/inbound-orders.
func (h *InboundOrderHandler) Create(c *gin.Context) {
	var input service.CreateInboundOrderInput
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

// Update handles PUT /api/v1/inbound-orders/:id.
func (h *InboundOrderHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateInboundOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	order, err := h.svc.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrInboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "inbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// Delete handles DELETE /api/v1/inbound-orders/:id.
func (h *InboundOrderHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrInboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "inbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, nil)
}

// Confirm handles POST /api/v1/inbound-orders/:id/confirm.
func (h *InboundOrderHandler) Confirm(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.svc.Confirm(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrInboundOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "inbound order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}
