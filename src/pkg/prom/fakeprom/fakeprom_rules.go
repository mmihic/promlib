package fakeprom

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"

	"promlib/src/pkg/prom"
)

// A QueryMatcher matches other queries.
type QueryMatcher[T any] interface {
	Matches(other T) bool
}

// A Rule contains a target query and a result - either an error or JSON.
type Rule[T QueryMatcher[T], R any] struct {
	Name   string
	Target T
	Err    error
	Result *R
}

// Matches returns true if this rule matches a given query.
func (r Rule[T, R]) Matches(other T) bool {
	return r.Target.Matches(other)
}

// UnmarshalJSON unmarshals the rule as JSON.
func (r *Rule[T, R]) UnmarshalJSON(b []byte) error {
	var repr ruleRepr[T, R]
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
func (r *Rule[T, R]) MarshalJSON() ([]byte, error) {
	repr := ruleRepr[T, R]{
		Name:   r.Name,
		Target: r.Target,
	}

	if r.Result != nil {
		resultJSON, err := json.MarshalIndent(r.Result, "", "  ")
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
func (r *Rule[T, R]) UnmarshalYAML(n *yaml.Node) error {
	var repr ruleRepr[T, R]
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

// A ruleRepr is the external (JSON, YAML) representation of a rule
type ruleRepr[T QueryMatcher[T], R any] struct {
	Name   string `json:"name" yaml:"name"`
	Target T      `json:"target" yaml:"target"`
	Err    string `json:"err,omitempty" yaml:"err,omitempty"`
	Result string `json:"result,omitempty" yaml:"result,omitempty"`
}

func (repr ruleRepr[T, R]) toRule() (Rule[T, R], error) {
	r := Rule[T, R]{
		Name:   repr.Name,
		Target: repr.Target,
	}

	if repr.Result != "" {
		var result R
		if err := json.Unmarshal([]byte(repr.Result), &result); err != nil {
			return r, err
		}

		r.Result = &result
	}

	if repr.Err != "" {
		r.Err = errors.New(repr.Err)
	}

	return r, nil
}

// Rules are fakeprom rules.
type Rules struct {
	InstantQueries InstantQueryRules `json:"instant_queries" yaml:"instant_queries"`
	RangeQueries   RangeQueryRules   `json:"range_queries" yaml:"range_queries"`
	LabelQueries   LabelQueryRules   `json:"label_queries" yaml:"label_queries"`
	SeriesQueries  SeriesQueryRules  `json:"series_queries" yaml:"series_queries"`
}

// Rule type aliases.
type (
	InstantQueryRule  = Rule[InstantQuery, prom.Result]
	InstantQueryRules = []Rule[InstantQuery, prom.Result]
	RangeQueryRule    = Rule[RangeQuery, prom.Result]
	RangeQueryRules   = []Rule[RangeQuery, prom.Result]
	LabelQueryRule    = Rule[LabelQuery, LabelResults]
	LabelQueryRules   = []Rule[LabelQuery, LabelResults]
	SeriesQueryRule   = Rule[SeriesQuery, SeriesResults]
	SeriesQueryRules  = []Rule[SeriesQuery, SeriesResults]
)

// LabelResults are the results of a labels query.
type LabelResults struct {
	Labels []string `json:"data" yaml:"data"`
}

// SeriesResults are the results of a series query.
type SeriesResults struct {
	Series []model.LabelSet `json:"data" yaml:"data"`
}

// FindMatchingResult finds the result that matches a given query from a set of rules.
func FindMatchingResult[T QueryMatcher[T], R any](rules []Rule[T, R], query T) (*R, error) {
	for _, rule := range rules {
		if rule.Matches(query) {
			if rule.Err != nil {
				return nil, rule.Err
			}

			return rule.Result, nil
		}
	}

	return nil, prom.NewErrorf(http.StatusNotFound, "matcher not found")
}

var (
	_ json.Unmarshaler = &RangeQueryRule{}
	_ yaml.Unmarshaler = &RangeQueryRule{}

	_ yaml.Unmarshaler = &InstantQueryRule{}
	_ json.Unmarshaler = &InstantQueryRule{}
)
