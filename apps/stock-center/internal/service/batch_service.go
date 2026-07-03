package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
)

var ErrDuplicateBatch = errors.New("duplicate batch for material and heat number")

// BatchService coordinates batch business operations.
type BatchService struct {
	batchRepo    *repository.BatchRepository
	materialRepo *repository.MaterialRepository
}

// NewBatchService creates a BatchService.
func NewBatchService(batchRepo *repository.BatchRepository, materialRepo *repository.MaterialRepository) *BatchService {
	return &BatchService{
		batchRepo:    batchRepo,
		materialRepo: materialRepo,
	}
}

// ListBatches returns paginated batches with optional qp filters.
func (s *BatchService) ListBatches(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.MaterialBatch, int64, error) {
	return s.batchRepo.List(ctx, page, pageSize, preds)
}

// GetBatch returns a batch by ID.
func (s *BatchService) GetBatch(ctx context.Context, id uint64) (*model.MaterialBatch, error) {
	return s.batchRepo.GetByID(ctx, id)
}

// CreateBatchInput is the payload for creating a batch.
type CreateBatchInput struct {
	MaterialID uint64 `json:"materialId"`
	HeatNo     string `json:"heatNo"`
	Remark     string `json:"remark"`
}

// CreateBatch creates a new batch linked to a material.
func (s *BatchService) CreateBatch(ctx context.Context, input CreateBatchInput) (*model.MaterialBatch, error) {
	if input.MaterialID == 0 {
		return nil, errors.New("materialId is required")
	}
	if input.HeatNo == "" {
		return nil, errors.New("heatNo is required")
	}

	if _, err := s.materialRepo.GetByID(ctx, input.MaterialID); err != nil {
		return nil, err
	}

	exists, err := s.batchRepo.ExistsByMaterialHeatOrg(ctx, input.MaterialID, input.HeatNo, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateBatch
	}

	batch := &model.MaterialBatch{
		MaterialID: input.MaterialID,
		HeatNo:     input.HeatNo,
		Remark:     input.Remark,
	}
	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, err
	}
	return batch, nil
}

// UpdateBatchInput is the payload for updating a batch.
type UpdateBatchInput struct {
	HeatNo string `json:"heatNo"`
	Remark string `json:"remark"`
}

// UpdateBatch updates an existing batch record.
func (s *BatchService) UpdateBatch(ctx context.Context, id uint64, input UpdateBatchInput) (*model.MaterialBatch, error) {
	batch, err := s.batchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newHeatNo := batch.HeatNo
	if input.HeatNo != "" {
		newHeatNo = input.HeatNo
	}
	if newHeatNo != batch.HeatNo {
		exists, err := s.batchRepo.ExistsByMaterialHeatOrg(ctx, batch.MaterialID, newHeatNo, batch.OrgID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrDuplicateBatch
		}
		batch.HeatNo = newHeatNo
	}

	if input.Remark != "" || batch.Remark != input.Remark {
		batch.Remark = input.Remark
	}

	if err := s.batchRepo.Update(ctx, batch); err != nil {
		return nil, err
	}
	return batch, nil
}
