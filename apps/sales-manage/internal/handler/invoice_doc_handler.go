package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"github.com/tserom/l-project/apps/sales-manage/internal/service"
	"github.com/tserom/l-project/apps/sales-manage/pkg/response"
)

// InvoiceDocHandler handles /api/v1/sale/invoice-docs.
type InvoiceDocHandler struct {
	svc *service.InvoiceDocService
}

func NewInvoiceDocHandler(svc *service.InvoiceDocService) *InvoiceDocHandler {
	return &InvoiceDocHandler{svc: svc}
}

func (h *InvoiceDocHandler) List(c *gin.Context) {
	preds, err := qp.InvoiceDocPredicates(c)
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

func (h *InvoiceDocHandler) Get(c *gin.Context) {
	doc, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, doc)
}

func (h *InvoiceDocHandler) Create(c *gin.Context) {
	var input service.CreateInvoiceDocInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}
	doc, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, doc)
}

func (h *InvoiceDocHandler) Void(c *gin.Context) {
	doc, err := h.svc.Void(c.Request.Context(), c.Param("id"))
	if err != nil {
		failService(c, err)
		return
	}
	response.OK(c, doc)
}
