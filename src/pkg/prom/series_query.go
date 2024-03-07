package prom

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// A SeriesQuery returns the set of series that match a set of selectors.
type SeriesQuery interface {
	Start(t time.Time) SeriesQuery
	End(t time.Time) SeriesQuery
	Selectors(sel []string) SeriesQuery
	Do(ctx context.Context) ([]model.LabelSet, error)
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
