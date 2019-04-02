package database

import (
	"github.com/pkg/errors"
)

type OperatorType int

const (
	OpEquals = iota
	OpGreaterThan
	OpGreaterThanOrEqual
	OpLessThan
	OpLessThanOrEqual
	OpContains
)

var (
	operatorTypes = map[string]OperatorType{
		"eq":       OpEquals,
		"gt":       OpGreaterThan,
		"gte":      OpGreaterThanOrEqual,
		"lt":       OpLessThan,
		"lte":      OpLessThanOrEqual,
		"contains": OpContains,
	}
)

type Operator struct {
	Type    OperatorType
	Field   string
	Operand interface{}
}

type QueryType int

const (
	QueryAnd = iota
	QueryNot
	QueryOr
)

type Query struct {
	Type       QueryType
	Conditions []Operator
}

// queryFromJson parses JSON query into Query object.
func QueryFromJson(rawQuery string) (*Query, error) {
	q := make(map[string]interface{})
	json.UnmarshalFromString(rawQuery, &q)

	query := &Query{
		Type:       QueryAnd,
		Conditions: []Operator{},
	}

	// TODO: support or / not condition
	for field, operator := range q {
		op := Operator{
			Field: field,
		}
		if rawOp, ok := operator.(map[string]interface{}); ok {
			var opType string
			opType, op.Operand = getFirstItem(rawOp)

			if op.Type, ok = operatorTypes[opType]; !ok {
				return nil, errors.Errorf("unknown operator: %s", opType)
			}
		} else {
			// an abbreviation of OpEquals: {"fieldName": value}
			op.Type = OpEquals
			op.Operand = operator
		}
		query.Conditions = append(query.Conditions, op)
	}
	return query, nil
}

func getFirstItem(m map[string]interface{}) (string, interface{}) {
	for k, v := range m {
		return k, v
	}
	return "", nil
}
