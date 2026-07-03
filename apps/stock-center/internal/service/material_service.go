package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
)

var ErrDuplicateMaterialCode = errors.New("duplicate material code")

// MaterialService coordinates material business operations.
type MaterialService struct {
	repo *repository.MaterialRepository
}

// NewMaterialService creates a MaterialService.
func NewMaterialService(repo *repository.MaterialRepository) *MaterialService {
	return &MaterialService{repo: repo}
}

// ListMaterials returns paginated materials with optional qp filters.
func (s *MaterialService) ListMaterials(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.Material, int64, error) {
	return s.repo.List(ctx, page, pageSize, preds)
}

// GetMaterial returns a material by ID.
func (s *MaterialService) GetMaterial(ctx context.Context, id uint64) (*model.Material, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateMaterialInput is the payload for creating a material.
type CreateMaterialInput struct {
	MaterialCode string               `json:"materialCode"`
	Grade        string               `json:"grade"`
	Form         model.MaterialForm   `json:"form"`
	Spec         string               `json:"spec"`
	PrimaryUnit  model.PrimaryUnit    `json:"primaryUnit"`
	MaterialType model.MaterialType   `json:"materialType"`
	Status       *model.MaterialStatus `json:"status"`
}

// CreateMaterial creates a new material record.
func (s *MaterialService) CreateMaterial(ctx context.Context, input CreateMaterialInput) (*model.Material, error) {
	if input.MaterialCode == "" {
		return nil, errors.New("materialCode is required")
	}
	if input.Grade == "" {
		return nil, errors.New("grade is required")
	}
	if input.Form == "" {
		return nil, errors.New("form is required")
	}
	if input.Spec == "" {
		return nil, errors.New("spec is required")
	}
	if input.PrimaryUnit == "" {
		return nil, errors.New("primaryUnit is required")
	}
	if input.MaterialType == "" {
		return nil, errors.New("materialType is required")
	}

	exists, err := s.repo.ExistsByMaterialCode(ctx, input.MaterialCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateMaterialCode
	}

	status := model.StatusEnabled
	if input.Status != nil {
		status = *input.Status
	}

	material := &model.Material{
		MaterialCode: input.MaterialCode,
		Grade:        input.Grade,
		Form:         input.Form,
		Spec:         input.Spec,
		PrimaryUnit:  input.PrimaryUnit,
		MaterialType: input.MaterialType,
		Status:       status,
	}
	if err := s.repo.Create(ctx, material); err != nil {
		return nil, err
	}
	return material, nil
}

// UpdateMaterialInput is the payload for updating a material.
type UpdateMaterialInput struct {
	Grade        string               `json:"grade"`
	Form         model.MaterialForm   `json:"form"`
	Spec         string               `json:"spec"`
	PrimaryUnit  model.PrimaryUnit    `json:"primaryUnit"`
	MaterialType model.MaterialType   `json:"materialType"`
	Status       model.MaterialStatus `json:"status"`
}

// UpdateMaterial updates an existing material record.
func (s *MaterialService) UpdateMaterial(ctx context.Context, id uint64, input UpdateMaterialInput) (*model.Material, error) {
	material, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Grade != "" {
		material.Grade = input.Grade
	}
	if input.Form != "" {
		material.Form = input.Form
	}
	if input.Spec != "" {
		material.Spec = input.Spec
	}
	if input.PrimaryUnit != "" {
		material.PrimaryUnit = input.PrimaryUnit
	}
	if input.MaterialType != "" {
		material.MaterialType = input.MaterialType
	}
	if input.Status != "" {
		material.Status = input.Status
	}

	if err := s.repo.Update(ctx, material); err != nil {
		return nil, err
	}
	return material, nil
}
