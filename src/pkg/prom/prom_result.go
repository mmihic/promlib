package prom

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prometheus/common/model"
)

// A Result is a result from a Prometheus query.
type Result struct {
	Data     model.Value
	Error    string
	Warnings []string
}

// ParseResult parses a Result.
func ParseResult(r io.Reader) (*Result, error) {
	var result Result
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UnmarshalJSON unmarshals a result from JSON.
func (r *Result) UnmarshalJSON(b []byte) error {
	var wire wireResult
	if err := json.Unmarshal(b, &wire); err != nil {
		return fmt.Errorf("unable to unmarshal wire format: %w", err)
	}

	result, err := wire.ToResult()
	if err != nil {
		return err
	}

	*r = *result
	return nil
}

// MarshalJSON marshals a result to JSON.
func (r *Result) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(wireResult{
		Data: data{
			Type:   r.Data.Type(),
			Result: value,
		},
		Error:    r.Error,
		Warnings: r.Warnings,
	})
}

type wireResult struct {
	Data     data     `json:"data"`
	Error    string   `json:"error"`
	Warnings []string `json:"warnings,omitempty"`
}

func (r wireResult) ToResult() (*Result, error) {
	v, err := r.Data.ToValue()
	if err != nil {
		return nil, fmt.Errorf("unable to convert data: %w", err)
	}
	return &Result{
		Data:     v,
		Error:    r.Error,
		Warnings: r.Warnings,
	}, nil
}

type data struct {
	Type   model.ValueType `json:"resultType"`
	Result json.RawMessage `json:"result"`
}

func (d data) ToValue() (model.Value, error) {
	switch d.Type {
	case model.ValString:
		var sv model.String
		if err := json.Unmarshal(d.Result, &sv); err != nil {
			return nil, err
		}
		return &sv, nil

	case model.ValScalar:
		var sv model.Scalar
		if err := json.Unmarshal(d.Result, &sv); err != nil {
			return nil, err
		}
		return &sv, nil

	case model.ValVector:
		var vv model.Vector
		if err := json.Unmarshal(d.Result, &vv); err != nil {
			return nil, err
		}
		return vv, nil

	case model.ValMatrix:
		var mv model.Matrix
		if err := json.Unmarshal(d.Result, &mv); err != nil {
			return nil, err
		}
		return mv, nil

	default:
		return nil, fmt.Errorf("unexpected value type %q", d.Type)
	}
}
