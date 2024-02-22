// Package fakeprom has a fake prometheus client, returning
// canned responses to queries.
package fakeprom

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/model"

	"promlib/src/pkg/prom"
)

// CannedInstantQuery is a matcher and query results for instant queries.
type CannedInstantQuery struct {
	Matches InstantQueryMatcher
	Err     error
	Result  string
}

// CannedRangeQuery is a matcher and query results for range queries.
type CannedRangeQuery struct {
	Matches RangeQueryMatcher
	Err     error
	Result  string
}

// InstantQueryMatcher matches instant queries.
type InstantQueryMatcher func(InstantQuery) bool

// MatchInstantQueryText is a matcher that matches an instant query
// through the query text.
func MatchInstantQueryText(query string) InstantQueryMatcher {
	return func(q InstantQuery) bool {
		return q.Query == query
	}
}

// RangeQueryMatcher matches range queries.
type RangeQueryMatcher func(RangeQuery) bool

// MatchRangeQueryText is a matcher that matches a range query
// through the query text.
func MatchRangeQueryText(query string) RangeQueryMatcher {
	return func(q RangeQuery) bool {
		return q.Query == query
	}
}

// Client is a fake prom client.
type Client interface {
	prom.Client
}

// NewClient returns a new fake prom.Client which returns canned responses.
func NewClient(
	instants []CannedInstantQuery,
	ranges []CannedRangeQuery,
) Client {
	return &client{
		instants: instants,
		ranges:   ranges,
	}
}

type client struct {
	instants []CannedInstantQuery
	ranges   []CannedRangeQuery
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
	Query      string
	StartTime  time.Time
	EndTime    time.Time
	StepPeriod model.Duration
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

// Do executes the range query.
func (q RangeQuery) Do(_ context.Context) (*prom.Result, error) {
	for _, canned := range q.c.ranges {
		if !canned.Matches(q) {
			continue
		}

		if canned.Err != nil {
			return nil, canned.Err
		}

		return prom.ParseResult(strings.NewReader(canned.Result))
	}

	return nil, prom.NewErrorf(http.StatusNotFound, "matcher for query '%s' not found", q.Query)
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
	Query string
	When  time.Time
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

		return prom.ParseResult(strings.NewReader(canned.Result))
	}

	return nil, prom.NewErrorf(http.StatusNotFound, "matcher for query '%s' not found", q.Query)
}

// Time sets the time for the instant query.
func (q InstantQuery) Time(t time.Time) prom.InstantQuery {
	q.When = t
	return q
}
