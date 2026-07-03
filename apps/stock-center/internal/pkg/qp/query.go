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

var materialAPIFields = map[string]string{
	"grade":        "grade",
	"form":         "form",
	"materialType": "material_type",
	"status":       "status",
}

var materialOperators = map[Operator]bool{
	OpEq:   true,
	OpLike: true,
	OpIn:   true,
}

var batchAPIFields = map[string]string{
	"materialId": "material_id",
}

var batchOperators = map[Operator]bool{
	OpEq: true,
}

// MaterialPredicates parses whitelisted qp-* filters for material list queries.
func MaterialPredicates(c *gin.Context) ([]Predicate, error) {
	return parsePredicates(c, materialAPIFields, materialOperators)
}

// BatchPredicates parses whitelisted qp-* filters for batch list queries.
func BatchPredicates(c *gin.Context) ([]Predicate, error) {
	return parsePredicates(c, batchAPIFields, batchOperators)
}

func parsePredicates(
	c *gin.Context,
	fields map[string]string,
	operators map[Operator]bool,
) ([]Predicate, error) {
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
		if !operators[op] {
			return nil, fmt.Errorf("%w: %s", ErrUnknownOperator, opStr)
		}

		column, ok := fields[apiField]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownField, apiField)
		}

		preds = append(preds, Predicate{
			Field:    column,
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
		}
	}
	return db
}
