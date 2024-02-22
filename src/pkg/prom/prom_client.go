// Package prom contains a simplified client for Prometheus.
package prom

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/mmihic/golib/src/pkg/httpclient"
	"github.com/prometheus/common/model"
)

const (
	pathInstantQuery = "/api/v1/query"
	pathRangeQuery   = "/api/v1/query_range"
)

// MetricsQuery is a query that returns metrics.
type MetricsQuery interface {
	Do(ctx context.Context) (*Result, error)
}

// An InstantQuery returns metrics at a specific point in time.
type InstantQuery interface {
	MetricsQuery
	Time(t time.Time) InstantQuery
}

// A RangeQuery returns metrics within a time range.
type RangeQuery interface {
	MetricsQuery
	Start(t time.Time) RangeQuery
	End(t time.Time) RangeQuery
	Step(n model.Duration) RangeQuery
}

// Client is a Prometheus client for running queries.
type Client interface {
	RangeQuery(q string) RangeQuery
	InstantQuery(q string) InstantQuery
}

// ClientOpt are options when creating a client.
type ClientOpt func(*client)

// WithHTTPClient sets the explicit HTTP client for talking to Prometheus.
func WithHTTPClient(httpClient httpclient.Client) ClientOpt {
	return func(c *client) {
		c.http = httpClient
	}
}

// WithHTTPOptions sets options for the HTTP client used to talk to Prometheus.
func WithHTTPOptions(opt ...httpclient.CallOption) ClientOpt {
	return func(c *client) {
		c.callOpts = append(c.callOpts, opt...)
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
		httpc, err := httpclient.NewClient(baseURL, c.callOpts...)
		if err != nil {
			return nil, err
		}

		c.http = httpc
	}
	c.callOpts = nil
	return c, nil
}

type client struct {
	http     httpclient.Client
	callOpts []httpclient.CallOption
}

func (c *client) InstantQuery(q string) InstantQuery {
	return instantQuery{
		q: q,
		c: c,
	}
}

type instantQuery struct {
	c *client
	q string
	t time.Time
}

func (q instantQuery) Time(t time.Time) InstantQuery {
	q.t = t
	return q
}

func (q instantQuery) Do(ctx context.Context) (*Result, error) {
	p := url.Values{}
	p.Add("query", q.q)
	if !q.t.IsZero() {
		p.Add("time", strconv.FormatInt(q.t.UTC().Unix(), 10))
	}

	var r Result
	if err := q.c.http.Post(ctx, pathInstantQuery, FormURLEncoded(p), httpclient.JSON(&r)); err != nil {
		if httperr, ok := httpclient.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}
		return nil, err
	}

	return &r, nil
}

func (c *client) RangeQuery(q string) RangeQuery {
	return rangeQuery{
		c:    c,
		q:    q,
		step: model.Duration(time.Minute),
	}
}

type rangeQuery struct {
	c          *client
	q          string
	start, end time.Time
	step       model.Duration
}

func (q rangeQuery) Start(t time.Time) RangeQuery {
	q.start = t
	return q
}

func (q rangeQuery) End(t time.Time) RangeQuery {
	q.end = t
	return q
}

func (q rangeQuery) Step(step model.Duration) RangeQuery {
	q.step = step
	return q
}

func (q rangeQuery) Do(ctx context.Context) (*Result, error) {
	if q.start.IsZero() {
		return nil, fmt.Errorf("'start' must be set for range queries")
	}

	if q.end.IsZero() {
		return nil, fmt.Errorf("'end' must be set for range queries")
	}

	p := url.Values{}
	p.Add("query", q.q)
	p.Add("start", strconv.FormatInt(q.start.Unix(), 10))
	p.Add("end", strconv.FormatInt(q.end.Unix(), 10))
	p.Add("step", q.step.String())

	var r Result
	if err := q.c.http.Post(ctx, pathRangeQuery, FormURLEncoded(p), httpclient.JSON(&r)); err != nil {
		if httperr, ok := httpclient.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	return &r, nil
}
