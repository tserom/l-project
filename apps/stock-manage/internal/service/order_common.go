package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
)

var (
	ErrOperatorRequired        = errors.New("operator is required")
	ErrCustomerRequired        = errors.New("customerName is required")
	ErrLinesRequired           = errors.New("at least one line is required")
	ErrAlreadyConfirmed        = errors.New("document is already confirmed")
	ErrSalesOrderNotConfirmed  = errors.New("sales order must be confirmed")
	ErrWarehouseRequired       = errors.New("warehouse is required when copying all order lines")
)

// CenterError wraps stock-center business errors for HTTP 400 mapping.
type CenterError struct {
	Message string
}

func (e *CenterError) Error() string {
	return e.Message
}

// StockMover calls stock-center inbound/outbound APIs.
type StockMover interface {
	InboundStock(ctx context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error)
	OutboundStock(ctx context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error)
}

// OrderLineInput is a line item on create/update requests.
type OrderLineInput struct {
	MaterialID uint64          `json:"materialId"`
	BatchID    uint64          `json:"batchId"`
	Warehouse  string          `json:"warehouse"`
	WeightKg   decimal.Decimal `json:"weightKg"`
	Quantity   decimal.Decimal `json:"quantity"`
}

// SalesOrderLineInput is a sales order line without warehouse.
type SalesOrderLineInput struct {
	MaterialID uint64           `json:"materialId"`
	BatchID    uint64           `json:"batchId"`
	WeightKg   decimal.Decimal  `json:"weightKg"`
	Quantity   decimal.Decimal  `json:"quantity"`
	UnitPrice  *decimal.Decimal `json:"unitPrice,omitempty"`
}

func validateOrderInput(operator string, lines []OrderLineInput) error {
	if strings.TrimSpace(operator) == "" {
		return ErrOperatorRequired
	}
	if len(lines) == 0 {
		return ErrLinesRequired
	}
	for i, line := range lines {
		if line.MaterialID == 0 {
			return errors.New("materialId is required on line " + strconv.Itoa(i+1))
		}
		if line.BatchID == 0 {
			return errors.New("batchId is required on line " + strconv.Itoa(i+1))
		}
		if strings.TrimSpace(line.Warehouse) == "" {
			return errors.New("warehouse is required on line " + strconv.Itoa(i+1))
		}
	}
	return nil
}

func parseDocDate(raw string) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return dateOnly(time.Now()), nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, errors.New("docDate must be YYYY-MM-DD")
	}
	return dateOnly(t), nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func toInboundLines(inputs []OrderLineInput) []model.InboundOrderLine {
	lines := make([]model.InboundOrderLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.InboundOrderLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			Warehouse:  in.Warehouse,
			WeightKg:   in.WeightKg,
			Quantity:   in.Quantity,
		}
	}
	return lines
}

func toOutboundLines(inputs []OrderLineInput) []model.OutboundOrderLine {
	lines := make([]model.OutboundOrderLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.OutboundOrderLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			Warehouse:  in.Warehouse,
			WeightKg:   in.WeightKg,
			Quantity:   in.Quantity,
		}
	}
	return lines
}

func validateSalesOrderInput(customerName, operator string, lines []SalesOrderLineInput) error {
	if strings.TrimSpace(customerName) == "" {
		return ErrCustomerRequired
	}
	if strings.TrimSpace(operator) == "" {
		return ErrOperatorRequired
	}
	if len(lines) == 0 {
		return ErrLinesRequired
	}
	for i, line := range lines {
		if line.MaterialID == 0 {
			return errors.New("materialId is required on line " + strconv.Itoa(i+1))
		}
		if line.BatchID == 0 {
			return errors.New("batchId is required on line " + strconv.Itoa(i+1))
		}
	}
	return nil
}

func validateShipmentLines(lines []OrderLineInput) error {
	if len(lines) == 0 {
		return ErrLinesRequired
	}
	for i, line := range lines {
		if line.MaterialID == 0 {
			return errors.New("materialId is required on line " + strconv.Itoa(i+1))
		}
		if line.BatchID == 0 {
			return errors.New("batchId is required on line " + strconv.Itoa(i+1))
		}
		if strings.TrimSpace(line.Warehouse) == "" {
			return errors.New("warehouse is required on line " + strconv.Itoa(i+1))
		}
	}
	return nil
}

func toSalesOrderLines(inputs []SalesOrderLineInput) []model.SalesOrderLine {
	lines := make([]model.SalesOrderLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.SalesOrderLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			WeightKg:   in.WeightKg,
			Quantity:   in.Quantity,
			UnitPrice:  in.UnitPrice,
		}
	}
	return lines
}

func toSalesShipmentLines(inputs []OrderLineInput) []model.SalesShipmentLine {
	lines := make([]model.SalesShipmentLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.SalesShipmentLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			Warehouse:  in.Warehouse,
			WeightKg:   in.WeightKg,
			Quantity:   in.Quantity,
		}
	}
	return lines
}

func validateProcessingOrderInput(
	operator string,
	pickLines []ProcessingPickLineInput,
	finishLines []ProcessingFinishLineInput,
) error {
	if strings.TrimSpace(operator) == "" {
		return ErrOperatorRequired
	}
	if len(pickLines) == 0 || len(finishLines) == 0 {
		return ErrLinesRequired
	}
	for i, line := range pickLines {
		if line.MaterialID == 0 {
			return errors.New("materialId is required on pick line " + strconv.Itoa(i+1))
		}
		if line.BatchID == 0 {
			return errors.New("batchId is required on pick line " + strconv.Itoa(i+1))
		}
		if strings.TrimSpace(line.Warehouse) == "" {
			return errors.New("warehouse is required on pick line " + strconv.Itoa(i+1))
		}
	}
	for i, line := range finishLines {
		if line.MaterialID == 0 {
			return errors.New("materialId is required on finish line " + strconv.Itoa(i+1))
		}
		if line.BatchID == 0 {
			return errors.New("batchId is required on finish line " + strconv.Itoa(i+1))
		}
		if strings.TrimSpace(line.Warehouse) == "" {
			return errors.New("warehouse is required on finish line " + strconv.Itoa(i+1))
		}
	}
	return nil
}

func toProcessingPickLines(inputs []ProcessingPickLineInput) []model.ProcessingPickLine {
	lines := make([]model.ProcessingPickLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.ProcessingPickLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			Warehouse:  in.Warehouse,
			WeightKg:   in.WeightKg,
		}
	}
	return lines
}

func toProcessingFinishLines(inputs []ProcessingFinishLineInput) []model.ProcessingFinishLine {
	lines := make([]model.ProcessingFinishLine, len(inputs))
	for i, in := range inputs {
		lines[i] = model.ProcessingFinishLine{
			MaterialID: in.MaterialID,
			BatchID:    in.BatchID,
			Warehouse:  in.Warehouse,
			Quantity:   in.Quantity,
			WeightKg:   in.WeightKg,
		}
	}
	return lines
}

func copySalesOrderLinesToShipment(orderLines []model.SalesOrderLine, warehouse string) []model.SalesShipmentLine {
	lines := make([]model.SalesShipmentLine, len(orderLines))
	for i, ol := range orderLines {
		lines[i] = model.SalesShipmentLine{
			MaterialID: ol.MaterialID,
			BatchID:    ol.BatchID,
			Warehouse:  warehouse,
			WeightKg:   ol.WeightKg,
			Quantity:   ol.Quantity,
		}
	}
	return lines
}

func wrapCenterErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	const prefix = "stock-center error: "
	if strings.HasPrefix(msg, prefix) {
		return &CenterError{Message: strings.TrimPrefix(msg, prefix)}
	}
	return err
}
