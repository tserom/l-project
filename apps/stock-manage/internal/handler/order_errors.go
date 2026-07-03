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

func parseID(raw string) (uint64, error) {
	return strconv.ParseUint(raw, 10, 64)
}

func failServiceError(c *gin.Context, err error) {
	var centerErr *service.CenterError
	if errors.As(err, &centerErr) {
		response.Fail(c, http.StatusBadRequest, 40000, centerErr.Message)
		return
	}

	switch {
	case errors.Is(err, service.ErrOperatorRequired),
		errors.Is(err, service.ErrCustomerRequired),
		errors.Is(err, service.ErrLinesRequired),
		errors.Is(err, service.ErrAlreadyConfirmed),
		errors.Is(err, service.ErrSalesOrderNotConfirmed),
		errors.Is(err, service.ErrWarehouseRequired),
		errors.Is(err, repository.ErrNotDraft):
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
	default:
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
	}
}
