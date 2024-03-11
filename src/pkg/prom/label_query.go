package prom

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"go.uber.org/zap"
)

// A LabelQuery returns the label names that match a set of selectors.
type LabelQuery interface {
	Start(t time.Time) LabelQuery
	End(t time.Time) LabelQuery
	Selectors(sel []string) LabelQuery
	Do(ctx context.Context) ([]string, error)
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

	log := q.c.queryLog.BeginQuery("labels-query",
		zap.Strings("sels", q.sels),
		zap.Time("start", q.start),
		zap.Time("end", q.end))

	var r labelsResult
	if err := q.c.http.Post(ctx, pathLabelQuery,
		httplib.FormURLEncoded(p), httplib.JSON(&r)); err != nil {
		log.QueryFailed(err)

		if httperr, ok := httplib.UnwrapError(err); ok {
			return nil, NewError(httperr.StatusCode, httperr.Body.String())
		}

		return nil, err
	}

	log.QueryComplete(&r)

	return r.Data, nil
}
