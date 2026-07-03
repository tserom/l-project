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

// SalesShipmentHandler handles sales shipment HTTP requests.
type SalesShipmentHandler struct {
	svc *service.ShipmentService
}

// NewSalesShipmentHandler creates a SalesShipmentHandler.
func NewSalesShipmentHandler(svc *service.ShipmentService) *SalesShipmentHandler {
	return &SalesShipmentHandler{svc: svc}
}

// List handles GET /api/v1/sales-shipments.
func (h *SalesShipmentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	shipments, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     shipments,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/sales-shipments/:id.
func (h *SalesShipmentHandler) Get(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	shipment, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSalesShipmentNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales shipment not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, shipment)
}

// Create handles POST /api/v1/sales-shipments.
func (h *SalesShipmentHandler) Create(c *gin.Context) {
	var input service.CreateShipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	shipment, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, shipment)
}

// Update handles PUT /api/v1/sales-shipments/:id.
func (h *SalesShipmentHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateShipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	shipment, err := h.svc.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrSalesShipmentNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales shipment not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, shipment)
}

// Delete handles DELETE /api/v1/sales-shipments/:id.
func (h *SalesShipmentHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrSalesShipmentNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales shipment not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, nil)
}

// Confirm handles POST /api/v1/sales-shipments/:id/confirm.
func (h *SalesShipmentHandler) Confirm(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	shipment, err := h.svc.Confirm(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSalesShipmentNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales shipment not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, shipment)
}
