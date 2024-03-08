// Package prom contains a simplified client for Prometheus.
package prom

import (
	"github.com/jonboulle/clockwork"
	"github.com/mmihic/httplib/src/pkg/httplib"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const (
	pathInstantQuery = "/api/v1/query"
	pathRangeQuery   = "/api/v1/query_range"
	pathLabelQuery   = "/api/v1/labels"
	pathSeriesQuery  = "/api/v1/series"
)

// Client is a Prometheus client for running queries.
type Client interface {
	RangeQuery(q string) RangeQuery
	InstantQuery(q string) InstantQuery
	MonthlyQuery(q string) MonthlyQuery
	LabelQuery() LabelQuery
	SeriesQuery() SeriesQuery
}

// ClientOpt are options when creating a client.
type ClientOpt func(*client)

// WithHTTPClient sets the explicit HTTP client for talking to Prometheus.
func WithHTTPClient(httpClient httplib.Client) ClientOpt {
	return func(c *client) {
		c.http = httpClient
	}
}

// WithHTTPOptions sets options for the HTTP client used to talk to Prometheus.
func WithHTTPOptions(opt ...httplib.CallOption) ClientOpt {
	return func(c *client) {
		c.callOpts = append(c.callOpts, opt...)
	}
}

// WithQueryLog sets a Logger for queries and query timers issued by the client.
func WithQueryLog(log *zap.Logger) ClientOpt {
	return func(c *client) {
		c.queryLog = log
	}
}

// WithClock sets the clock used by the client.
func WithClock(clock clockwork.Clock) ClientOpt {
	return func(c *client) {
		c.clock = clock
	}
}

// NewClient creates a new Prometheus client against a base URL and a set
// of client options.
func NewClient(baseURL string, opts ...ClientOpt) (Client, error) {
	c := &client{}

	for _, opt := range opts {
		opt(c)
	}

	if c.http == nil {
		httpc, err := httplib.NewClient(baseURL, httplib.WithDefaultCallOptions(c.callOpts...))
		if err != nil {
			return nil, err
		}

		c.http = httpc
	}
	c.callOpts = nil

	if c.queryLog == nil {
		c.queryLog = zap.NewNop()
	}
	if c.clock == nil {
		c.clock = clockwork.NewRealClock()
	}

	return c, nil
}

type client struct {
	http     httplib.Client
	callOpts []httplib.CallOption
	queryLog *zap.Logger
	queryID  atomic.Uint64
	clock    clockwork.Clock
}
