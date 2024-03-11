package prom

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"go.uber.org/zap"
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

	log := q.c.queryLog.BeginQuery("instant-query",
		zap.String("query", q.q),
		zap.Time("time", q.t))

	var r Result
	if err := q.c.http.Post(ctx, pathInstantQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		log.QueryFailed(err)
		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}
		return nil, err
	}

	log.QueryComplete(&r)
	return &r, nil
}
