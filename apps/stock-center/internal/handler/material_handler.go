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

// MaterialHandler handles material HTTP requests.
type MaterialHandler struct {
	svc *service.MaterialService
}

// NewMaterialHandler creates a MaterialHandler.
func NewMaterialHandler(svc *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{svc: svc}
}

// List handles GET /api/v1/materials.
func (h *MaterialHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	preds, err := qp.MaterialPredicates(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	materials, total, err := h.svc.ListMaterials(c.Request.Context(), page, pageSize, preds)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, gin.H{
		"list":     materials,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/materials/:id.
func (h *MaterialHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	material, err := h.svc.GetMaterial(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMaterialNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "material not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.OK(c, material)
}

// Create handles POST /api/v1/materials.
func (h *MaterialHandler) Create(c *gin.Context) {
	var input service.CreateMaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	material, err := h.svc.CreateMaterial(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateMaterialCode) {
			response.Fail(c, http.StatusBadRequest, 40000, err.Error())
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, material)
}

// Update handles PUT /api/v1/materials/:id.
func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var input service.UpdateMaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	material, err := h.svc.UpdateMaterial(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrMaterialNotFound) {
			response.Fail(c, http.StatusNotFound, 40400, "material not found")
			return
		}
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.OK(c, material)
}
