// Package fakeprom has a fake prometheus client, returning
// canned responses to queries.
package fakeprom

import (
	"context"
	"time"

	"github.com/mmihic/golib/src/pkg/set"
	"github.com/prometheus/common/model"

	"promlib/src/pkg/prom"
)

// Client is a fake prom client.
type Client interface {
	AddInstantQueryRules(rules ...InstantQueryRule)
	AddRangeQueryRules(rules ...RangeQueryRule)
	AddLabelQueryRules(rules ...LabelQueryRule)
	AddSeriesQueryRules(rules ...SeriesQueryRule)
	prom.Client
}

// NewClient returns a new fake prom.Client which returns canned responses.
func NewClient() Client {
	return &client{}
}

// NewClientWithRules returns a new fake prom.Client initialized with a set
// of query rules.
func NewClientWithRules(rules *Rules) Client {
	return &client{
		instants: rules.InstantQueries,
		ranges:   rules.RangeQueries,
		series:   rules.SeriesQueries,
		labels:   rules.LabelQueries,
	}
}

type client struct {
	instants InstantQueryRules
	ranges   RangeQueryRules
	labels   LabelQueryRules
	series   SeriesQueryRules
}

func (c *client) AddInstantQueryRules(rules ...InstantQueryRule) {
	c.instants = append(c.instants, rules...)
}

func (c *client) AddRangeQueryRules(rules ...RangeQueryRule) {
	c.ranges = append(c.ranges, rules...)
}

func (c *client) AddLabelQueryRules(rules ...LabelQueryRule) {
	c.labels = append(c.labels, rules...)
}

func (c *client) AddSeriesQueryRules(rules ...SeriesQueryRule) {
	c.series = append(c.series, rules...)
}

func (c *client) RangeQuery(q string) prom.RangeQuery {
	return RangeQuery{
		c:     c,
		Query: q,
	}
}

// RangeQuery contains the parameters for a range query.
type RangeQuery struct {
	c          *client
	Query      string         `json:"query,omitempty" yaml:"query"`
	StartTime  time.Time      `json:"start_time" yaml:"start_time"`
	EndTime    time.Time      `json:"end_time" yaml:"end_time"`
	StepPeriod model.Duration `json:"step_period" yaml:"step"`
}

// Start sets the start time for the query.
func (q RangeQuery) Start(t time.Time) prom.RangeQuery {
	q.StartTime = t
	return q
}

// End sets the end time for the query.
func (q RangeQuery) End(t time.Time) prom.RangeQuery {
	q.EndTime = t
	return q
}

// Step sets the step period for the query.
func (q RangeQuery) Step(n model.Duration) prom.RangeQuery {
	q.StepPeriod = n
	return q
}

// Matches returns true if this query matches another range query.
func (q RangeQuery) Matches(other RangeQuery) bool {
	if q.StepPeriod != 0 && q.StepPeriod != other.StepPeriod {
		return false
	}

	if !q.StartTime.IsZero() && !q.StartTime.Equal(other.StartTime) {
		return false
	}

	if !q.EndTime.IsZero() && !q.EndTime.Equal(other.EndTime) {
		return false
	}

	eq, err := QueryEqual(q.Query, other.Query)
	if err != nil {
		panic(err)
	}

	return eq
}

// Do executes the range query.
func (q RangeQuery) Do(_ context.Context) (*prom.Result, error) {
	return FindMatchingResult(q.c.ranges, q)
}

// InstantQuery returns a new instant query.
func (c *client) InstantQuery(q string) prom.InstantQuery {
	return InstantQuery{
		c:     c,
		Query: q,
	}
}

// InstantQuery is an instant query.
type InstantQuery struct {
	c     *client
	Query string    `json:"query,omitempty" yaml:"query"`
	When  time.Time `json:"when" yaml:"when"`
}

// Do executes the instant query.
func (q InstantQuery) Do(_ context.Context) (*prom.Result, error) {
	return FindMatchingResult(q.c.instants, q)
}

// Time sets the time for the instant query.
func (q InstantQuery) Time(t time.Time) prom.InstantQuery {
	q.When = t
	return q
}

// Matches checks whether this query matches another query.
func (q InstantQuery) Matches(other InstantQuery) bool {
	if !q.When.IsZero() && !q.When.Equal(other.When) {
		return false
	}

	eq, err := QueryEqual(q.Query, other.Query)
	if err != nil {
		panic(err)
	}

	return eq
}

func (c *client) LabelQuery() prom.LabelQuery {
	return LabelQuery{
		c: c,
	}
}

// LabelQuery is a fake query for labels.
type LabelQuery struct {
	c *client

	StartTime time.Time       `json:"start_time" yaml:"start_time"`
	EndTime   time.Time       `json:"end_time" yaml:"end_time"`
	Sels      set.Set[string] `json:"selectors" yaml:"selectors"`
}

func (q LabelQuery) Start(t time.Time) prom.LabelQuery {
	q.StartTime = t
	return q
}

func (q LabelQuery) End(t time.Time) prom.LabelQuery {
	q.EndTime = t
	return q
}

func (q LabelQuery) Selectors(sels []string) prom.LabelQuery {
	q.Sels = set.NewSet(sels...)
	return q
}

func (q LabelQuery) Matches(other LabelQuery) bool {
	if !q.StartTime.IsZero() && !q.StartTime.Equal(other.StartTime) {
		return false
	}

	if !q.EndTime.IsZero() && !q.EndTime.Equal(other.EndTime) {
		return false
	}

	return q.Sels.Equal(other.Sels)
}

func (q LabelQuery) Do(_ context.Context) ([]string, error) {
	r, err := FindMatchingResult(q.c.labels, q)
	if err != nil {
		return nil, err
	}

	return r.Labels, nil
}

func (c *client) SeriesQuery() prom.SeriesQuery {
	return SeriesQuery{
		c: c,
	}
}

// SeriesQuery is a fake query for labels.
type SeriesQuery struct {
	c *client

	StartTime time.Time       `json:"start_time" yaml:"start_time"`
	EndTime   time.Time       `json:"end_time" yaml:"end_time"`
	Sels      set.Set[string] `json:"selectors" yaml:"selectors"`
}

func (q SeriesQuery) Start(t time.Time) prom.SeriesQuery {
	q.StartTime = t
	return q
}

func (q SeriesQuery) End(t time.Time) prom.SeriesQuery {
	q.EndTime = t
	return q
}

func (q SeriesQuery) Selectors(sels []string) prom.SeriesQuery {
	q.Sels = set.NewSet(sels...)
	return q
}

func (q SeriesQuery) Matches(other SeriesQuery) bool {
	if !q.StartTime.IsZero() && !q.StartTime.Equal(other.StartTime) {
		return false
	}

	if !q.EndTime.IsZero() && !q.EndTime.Equal(other.EndTime) {
		return false
	}

	return q.Sels.Equal(other.Sels)
}

func (q SeriesQuery) Do(_ context.Context) ([]model.LabelSet, error) {
	r, err := FindMatchingResult(q.c.series, q)
	if err != nil {
		return nil, err
	}

	return r.Series, nil
}
