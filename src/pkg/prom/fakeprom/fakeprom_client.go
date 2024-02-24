// Package fakeprom has a fake prometheus client, returning
// canned responses to queries.
package fakeprom

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/common/model"

	"promlib/src/pkg/prom"
)

// Client is a fake prom client.
type Client interface {
	AddInstantQueryRules(rules ...Rule[InstantQuery])
	AddRangeQueryRules(rules ...Rule[RangeQuery])
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
	}
}

type client struct {
	instants []Rule[InstantQuery]
	ranges   []Rule[RangeQuery]
}

func (c *client) AddInstantQueryRules(rules ...Rule[InstantQuery]) {
	c.instants = append(c.instants, rules...)
}

func (c *client) AddRangeQueryRules(rules ...Rule[RangeQuery]) {
	c.ranges = append(c.ranges, rules...)
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
	for _, canned := range q.c.ranges {
		if !canned.Matches(q) {
			continue
		}

		if canned.Err != nil {
			return nil, canned.Err
		}

		return canned.Result, nil
	}

	return nil, prom.NewErrorf(http.StatusNotFound,
		"matcher for query '%s' not found", q.Query)
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
	for _, canned := range q.c.instants {
		if !canned.Matches(q) {
			continue
		}

		if canned.Err != nil {
			return nil, canned.Err
		}

		return canned.Result, nil
	}

	return nil, prom.NewErrorf(http.StatusNotFound, "matcher for query '%s' not found", q.Query)
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
