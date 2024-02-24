package fakeprom

import (
	"fmt"

	"github.com/prometheus/prometheus/promql/parser"
)

// QueryEqual compares two queries for equality after normalization.
func QueryEqual(q1, q2 string) (bool, error) {
	expr1, err := parser.ParseExpr(q1)
	if err != nil {
		return false, fmt.Errorf("first query is invalid: %w", err)
	}

	expr2, err := parser.ParseExpr(q2)
	if err != nil {
		return false, fmt.Errorf("second query is invalid: %w", err)
	}

	return expr1.String() == expr2.String(), nil
}

