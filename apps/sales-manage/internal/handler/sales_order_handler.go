package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"github.com/tserom/l-project/apps/sales-manage/internal/service"
	"github.com/tserom/l-project/apps/sales-manage/pkg/response"
)

// SalesOrderHandler handles /api/v1/sale/orders.
type SalesOrderHandler struct {
	svc *service.SalesOrderService
}

func NewSalesOrderHandler(svc *service.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{svc: svc}
}

func (h *SalesOrderHandler) List(c *gin.Context) {
	preds, err := qp.SalesOrderPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.svc.List(c.Request.Context(), preds, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.OK(c, gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *SalesOrderHandler) Get(c *gin.Context) {
	order, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, order)
}

func (h *SalesOrderHandler) Create(c *gin.Context) {
	var input service.SalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}
	order, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, order)
}

func (h *SalesOrderHandler) Update(c *gin.Context) {
	var input service.SalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}
	order, err := h.svc.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, order)
}

func (h *SalesOrderHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		failService(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *SalesOrderHandler) ListInvoiceCandidates(c *gin.Context) {
	preds, err := qp.SalesOrderPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}
	list, err := h.svc.ListInvoiceCandidates(c.Request.Context(), preds)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func failService(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOrderNotFound), errors.Is(err, service.ErrInvoiceNotFound):
		response.Fail(c, http.StatusNotFound, 40400, err.Error())
	case errors.Is(err, service.ErrValidation),
		errors.Is(err, service.ErrOrderHasInvoice),
		errors.Is(err, service.ErrInvoiceAlreadyVoid),
		errors.Is(err, service.ErrInvoiceLinesEmpty):
		response.Fail(c, http.StatusBadRequest, 40000, unwrapMsg(err))
	default:
		msg := err.Error()
		if strings.Contains(msg, "已开票") ||
			strings.Contains(msg, "不存在") ||
			strings.Contains(msg, "未勾选") {
			response.Fail(c, http.StatusBadRequest, 40000, msg)
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, msg)
	}
}

func unwrapMsg(err error) string {
	msg := err.Error()
	// fmt.Errorf("%w: xxx") → prefer the human part after ": "
	if i := strings.Index(msg, ": "); i >= 0 && i+2 < len(msg) {
		return msg[i+2:]
	}
	return msg
}
