package prom

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// A RangeQuery returns metrics within a time range.
type RangeQuery interface {
	MetricsQuery
	Start(t time.Time) RangeQuery
	End(t time.Time) RangeQuery
	Step(n model.Duration) RangeQuery
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
	log := q.c.queryLog.BeginQuery("range-query",
		zap.String("query", q.q),
		zap.Time("start", q.start),
		zap.Time("end", q.end),
		zap.Duration("step", time.Duration(q.step)))

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
	if err := q.c.http.Post(ctx, pathRangeQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		log.QueryFailed(err)

		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	log.QueryComplete(r)
	return &r, nil
}
