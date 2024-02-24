package fakeprom

import (
	"encoding/json"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"

	"promlib/src/pkg/prom"
)

// A QueryMatcher matches other queries.
type QueryMatcher[T any] interface {
	Matches(other T) bool
}

// A Rule contains a target query and a result - either an error or JSON.
type Rule[T QueryMatcher[T]] struct {
	Name   string
	Target T
	Err    error
	Result *prom.Result
}

// Matches returns true if this rule matches a given query.
func (r Rule[T]) Matches(other T) bool {
	return r.Target.Matches(other)
}

// UnmarshalJSON unmarshals the rule as JSON.
func (r *Rule[T]) UnmarshalJSON(b []byte) error {
	var repr ruleRepr[T]
	if err := json.Unmarshal(b, &repr); err != nil {
		return err
	}

	asRule, err := repr.toRule()
	if err != nil {
		return err
	}

	*r = asRule
	return nil
}

// MarshalJSON marshals the rule as JSON.
func (r *Rule[T]) MarshalJSON() ([]byte, error) {
	repr := ruleRepr[T]{
		Name:   r.Name,
		Target: r.Target,
	}

	if r.Result != nil {
		resultJSON, err := r.Result.MarshalJSON()
		if err != nil {
			return nil, err
		}

		repr.Result = string(resultJSON)
	}

	if r.Err != nil {
		repr.Err = r.Err.Error()
	}

	return json.MarshalIndent(repr, "", "  ")
}

// UnmarshalYAML unmarshals the rule as YAML.
func (r *Rule[T]) UnmarshalYAML(n *yaml.Node) error {
	var repr ruleRepr[T]
	if err := n.Decode(&repr); err != nil {
		return err
	}

	asRule, err := repr.toRule()
	if err != nil {
		return err
	}

	*r = asRule
	return nil
}

// MarshalYAML marshals the rule as YAML.

// A ruleRepr is the external (JSON, YAML) representation of a rule
type ruleRepr[T QueryMatcher[T]] struct {
	Name   string `json:"name" yaml:"name"`
	Target T      `json:"target" yaml:"target"`
	Err    string `json:"err,omitempty" yaml:"err,omitempty"`
	Result string `json:"result,omitempty" yaml:"result,omitempty"`
}

func (repr ruleRepr[T]) toRule() (Rule[T], error) {
	r := Rule[T]{
		Name:   repr.Name,
		Target: repr.Target,
	}

	if repr.Result != "" {
		result, err := prom.ParseResult(strings.NewReader(repr.Result))
		if err != nil {
			return r, err
		}

		r.Result = result
	}

	if repr.Err != "" {
		r.Err = errors.New(repr.Err)
	}

	return r, nil
}

// Rules are fakeprom rules.
type Rules struct {
	InstantQueries []Rule[InstantQuery] `json:"instant_queries" yaml:"instant_queries"`
	RangeQueries   []Rule[RangeQuery]   `json:"range_queries" yaml:"range_queries"`
}

var (
	_ json.Unmarshaler = &Rule[RangeQuery]{}
	_ yaml.Unmarshaler = &Rule[RangeQuery]{}

	_ yaml.Unmarshaler = &Rule[InstantQuery]{}
	_ json.Unmarshaler = &Rule[InstantQuery]{}
)
