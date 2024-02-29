// Package prom contains a simplified client for Prometheus.
package prom

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/prometheus/common/model"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const (
	pathInstantQuery = "/api/v1/query"
	pathRangeQuery   = "/api/v1/query_range"
	pathLabelQuery   = "/api/v1/labels"
	pathSeriesQuery  = "/api/v1/series"
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

// A LabelQuery returns the label names that match a set of selectors.
type LabelQuery interface {
	Start(t time.Time) LabelQuery
	End(t time.Time) LabelQuery
	Selectors(sel []string) LabelQuery
	Do(ctx context.Context) ([]string, error)
}

// A SeriesQuery returns the set of series that match a set of selectors.
type SeriesQuery interface {
	Start(t time.Time) SeriesQuery
	End(t time.Time) SeriesQuery
	Selectors(sel []string) SeriesQuery
	Do(ctx context.Context) ([]model.LabelSet, error)
}

// Client is a Prometheus client for running queries.
type Client interface {
	RangeQuery(q string) RangeQuery
	InstantQuery(q string) InstantQuery
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
	qid := q.c.queryID.Add(1)
	startTime := q.c.clock.Now()

	p := url.Values{}
	p.Add("query", q.q)
	if !q.t.IsZero() {
		p.Add("time", strconv.FormatInt(q.t.UTC().Unix(), 10))
	}

	q.c.queryLog.Info("instant-query",
		zap.String("query", q.q),
		zap.Uint64("id", qid),
		zap.Time("time", q.t))

	var r Result
	if err := q.c.http.Post(ctx, pathInstantQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		q.c.queryLog.Error("instant-query",
			zap.Uint64("id", qid),
			zap.Error(err))
		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}
		return nil, err
	}

	endTime := q.c.clock.Now()
	q.c.queryLog.Info("instant-query",
		zap.Uint64("id", qid),
		zap.Strings("warnings", r.Warnings),
		zap.Duration("elapsed", endTime.Sub(startTime)))

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
	qid := q.c.queryID.Add(1)
	startTime := q.c.clock.Now()

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

	q.c.queryLog.Info("range-query",
		zap.String("query", q.q),
		zap.Uint64("id", qid),
		zap.Time("start", q.start),
		zap.Time("end", q.end),
		zap.Duration("step", time.Duration(q.step)))

	var r Result
	if err := q.c.http.Post(ctx, pathRangeQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		q.c.queryLog.Error("range-query",
			zap.Uint64("id", qid),
			zap.Error(err))

		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	endTime := q.c.clock.Now()
	q.c.queryLog.Info("range-query",
		zap.Uint64("id", qid),
		zap.Strings("warnings", r.Warnings),
		zap.Duration("elapsed", endTime.Sub(startTime)))

	return &r, nil
}

func (c *client) LabelQuery() LabelQuery {
	return labelQuery{
		c: c,
	}
}

type labelQuery struct {
	c     *client
	sels  []string
	start time.Time
	end   time.Time
}

func (q labelQuery) Selectors(sels []string) LabelQuery {
	q.sels = sels
	return q
}

func (q labelQuery) Start(t time.Time) LabelQuery {
	q.start = t
	return q
}

func (q labelQuery) End(t time.Time) LabelQuery {
	q.end = t
	return q
}

func (q labelQuery) Do(ctx context.Context) ([]string, error) {
	qid := q.c.queryID.Add(1)
	startTime := q.c.clock.Now()

	p := url.Values{}

	if !q.start.IsZero() {
		p.Add("start", strconv.FormatInt(q.start.Unix(), 10))
	}

	if !q.end.IsZero() {
		p.Add("end", strconv.FormatInt(q.end.Unix(), 10))
	}

	if len(q.sels) != 0 {
		for _, sel := range q.sels {
			p.Add("match[]", sel)
		}
	}

	q.c.queryLog.Info("labels-query",
		zap.Strings("sels", q.sels),
		zap.Uint64("id", qid),
		zap.Time("start", q.start),
		zap.Time("end", q.end))

	var r labelsResult
	if err := q.c.http.Post(ctx, pathLabelQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		q.c.queryLog.Error("labels-query",
			zap.Uint64("id", qid),
			zap.Error(err))

		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	endTime := q.c.clock.Now()
	q.c.queryLog.Info("range-query",
		zap.Uint64("id", qid),
		zap.Duration("elapsed", endTime.Sub(startTime)))

	return r.Data, nil
}

func (c *client) SeriesQuery() SeriesQuery {
	return seriesQuery{
		c: c,
	}
}

type seriesQuery struct {
	c     *client
	sels  []string
	start time.Time
	end   time.Time
}

func (q seriesQuery) Selectors(sels []string) SeriesQuery {
	q.sels = sels
	return q
}

func (q seriesQuery) Start(t time.Time) SeriesQuery {
	q.start = t
	return q
}

func (q seriesQuery) End(t time.Time) SeriesQuery {
	q.end = t
	return q
}

func (q seriesQuery) Do(ctx context.Context) ([]model.LabelSet, error) {
	qid := q.c.queryID.Add(1)
	startTime := q.c.clock.Now()

	p := url.Values{}

	if !q.start.IsZero() {
		p.Add("start", strconv.FormatInt(q.start.Unix(), 10))
	}

	if !q.end.IsZero() {
		p.Add("end", strconv.FormatInt(q.end.Unix(), 10))
	}

	if len(q.sels) != 0 {
		for _, sel := range q.sels {
			p.Add("match[]", sel)
		}
	}

	q.c.queryLog.Info("series-query",
		zap.Strings("sels", q.sels),
		zap.Uint64("id", qid),
		zap.Time("start", q.start),
		zap.Time("end", q.end))

	var r seriesResult
	if err := q.c.http.Post(ctx, pathSeriesQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		q.c.queryLog.Error("series-query",
			zap.Uint64("id", qid),
			zap.Error(err))

		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	endTime := q.c.clock.Now()
	q.c.queryLog.Info("series-query",
		zap.Uint64("id", qid),
		zap.Duration("elapsed", endTime.Sub(startTime)))

	return r.Data, nil
}
