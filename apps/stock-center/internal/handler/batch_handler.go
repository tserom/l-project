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

// BatchHandler handles batch HTTP requests.
type BatchHandler struct {
	svc *service.BatchService
}

// NewBatchHandler creates a BatchHandler.
func NewBatchHandler(svc *service.BatchService) *BatchHandler {
	return &BatchHandler{svc: svc}
}

// List handles GET /api/v1/batches.
func (h *BatchHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	preds, err := qp.BatchPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	batches, total, err := h.svc.ListBatches(c.Request.Context(), page, pageSize, preds)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     batches,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/batches/:id.
func (h *BatchHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	batch, err := h.svc.GetBatch(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrBatchNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "batch not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, batch)
}

// Create handles POST /api/v1/batches.
func (h *BatchHandler) Create(c *gin.Context) {
	var input service.CreateBatchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	batch, err := h.svc.CreateBatch(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateBatch) {
			response.Fail(c, http.StatusBadRequest, 40000, err.Error())
			return
		}
		if errors.Is(err, repository.ErrMaterialNotFound) {
			response.Fail(c, http.StatusBadRequest, 40000, "material not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, batch)
}

// Update handles PUT /api/v1/batches/:id.
func (h *BatchHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateBatchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	batch, err := h.svc.UpdateBatch(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrBatchNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "batch not found")
			return
		}
		if errors.Is(err, service.ErrDuplicateBatch) {
			response.Fail(c, http.StatusBadRequest, 40000, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, batch)
}
