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

// SalesOrderHandler handles sales order HTTP requests.
type SalesOrderHandler struct {
	orderSvc    *service.SalesOrderService
	shipmentSvc *service.ShipmentService
}

// NewSalesOrderHandler creates a SalesOrderHandler.
func NewSalesOrderHandler(orderSvc *service.SalesOrderService, shipmentSvc *service.ShipmentService) *SalesOrderHandler {
	return &SalesOrderHandler{orderSvc: orderSvc, shipmentSvc: shipmentSvc}
}

// List handles GET /api/v1/sales-orders.
func (h *SalesOrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	orders, total, err := h.orderSvc.List(c.Request.Context(), page, pageSize)
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

// Get handles GET /api/v1/sales-orders/:id.
func (h *SalesOrderHandler) Get(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.orderSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, order)
}

// Create handles POST /api/v1/sales-orders.
func (h *SalesOrderHandler) Create(c *gin.Context) {
	var input service.CreateSalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	order, err := h.orderSvc.Create(c.Request.Context(), input)
	if err != nil {
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// Update handles PUT /api/v1/sales-orders/:id.
func (h *SalesOrderHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateSalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	order, err := h.orderSvc.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// Delete handles DELETE /api/v1/sales-orders/:id.
func (h *SalesOrderHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.orderSvc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, nil)
}

// Confirm handles POST /api/v1/sales-orders/:id/confirm.
func (h *SalesOrderHandler) Confirm(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	order, err := h.orderSvc.Confirm(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		failServiceError(c, err)
		return
	}

	response.OK(c, order)
}

// ListShipments handles GET /api/v1/sales-orders/:id/shipments.
func (h *SalesOrderHandler) ListShipments(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	shipments, err := h.shipmentSvc.ListBySalesOrderID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "sales order not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, shipments)
}

// CreateShipment handles POST /api/v1/sales-orders/:id/shipments.
func (h *SalesOrderHandler) CreateShipment(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.CreateShipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	shipment, err := h.shipmentSvc.CreateFromSalesOrder(c.Request.Context(), id, input)
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
