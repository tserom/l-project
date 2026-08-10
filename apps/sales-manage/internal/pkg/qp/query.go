package qp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const qpPrefix = "qp-"

// Operator is a whitelisted query predicate operator.
type Operator string

const (
	OpEq   Operator = "eq"
	OpLike Operator = "like"
	OpIn   Operator = "in"
	OpGte  Operator = "gte"
	OpLte  Operator = "lte"
)

// Predicate is a parsed qp-<field>-<operator> filter.
type Predicate struct {
	Field    string
	Operator Operator
	Value    string
}

var (
	ErrUnknownField    = errors.New("unknown query field")
	ErrUnknownOperator = errors.New("unknown query operator")
)

// FieldSpec maps an API field to a DB column and allowed operators.
type FieldSpec struct {
	Column string
	Ops    map[Operator]bool
}

func salesOrderFields() map[string]FieldSpec {
	return map[string]FieldSpec{
		"customerName": {
			Column: "customer_name",
			Ops:    map[Operator]bool{OpLike: true},
		},
		"orderNo": {
			Column: "order_no",
			Ops:    map[Operator]bool{OpEq: true, OpIn: true, OpLike: true},
		},
		"orderDate": {
			Column: "order_date",
			Ops:    map[Operator]bool{OpGte: true, OpLte: true},
		},
		"warehouseName": {
			Column: "warehouse_name",
			Ops:    map[Operator]bool{OpEq: true},
		},
	}
}

func invoiceDocFields() map[string]FieldSpec {
	return map[string]FieldSpec{
		"status": {
			Column: "status",
			Ops:    map[Operator]bool{OpEq: true},
		},
		"invoiceNo": {
			Column: "invoice_no",
			Ops:    map[Operator]bool{OpLike: true},
		},
	}
}

// SalesOrderPredicates parses whitelisted qp-* for sales order lists / candidates.
func SalesOrderPredicates(c *gin.Context) ([]Predicate, error) {
	return parsePredicates(c, salesOrderFields())
}

// InvoiceDocPredicates parses whitelisted qp-* for invoice doc lists.
func InvoiceDocPredicates(c *gin.Context) ([]Predicate, error) {
	return parsePredicates(c, invoiceDocFields())
}

func parsePredicates(c *gin.Context, fields map[string]FieldSpec) ([]Predicate, error) {
	var preds []Predicate
	for key, values := range c.Request.URL.Query() {
		if !strings.HasPrefix(key, qpPrefix) {
			continue
		}
		if len(values) == 0 {
			continue
		}

		rest := strings.TrimPrefix(key, qpPrefix)
		idx := strings.LastIndex(rest, "-")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid query parameter: %s", key)
		}

		apiField := rest[:idx]
		opStr := rest[idx+1:]
		op := Operator(opStr)

		spec, ok := fields[apiField]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownField, apiField)
		}
		if !spec.Ops[op] {
			return nil, fmt.Errorf("%w: %s for field %s", ErrUnknownOperator, opStr, apiField)
		}

		preds = append(preds, Predicate{
			Field:    spec.Column,
			Operator: op,
			Value:    values[0],
		})
	}
	return preds, nil
}

// Apply scopes a GORM query with the given predicates (parameter-bound).
func Apply(db *gorm.DB, preds []Predicate) *gorm.DB {
	for _, p := range preds {
		switch p.Operator {
		case OpEq:
			db = db.Where(p.Field+" = ?", p.Value)
		case OpLike:
			db = db.Where(p.Field+" LIKE ?", "%"+p.Value+"%")
		case OpIn:
			parts := strings.Split(p.Value, ",")
			db = db.Where(p.Field+" IN ?", parts)
		case OpGte:
			db = db.Where(p.Field+" >= ?", p.Value)
		case OpLte:
			db = db.Where(p.Field+" <= ?", p.Value)
		}
	}
	return db
}
