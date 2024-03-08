package prom

import (
	"context"

	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/prometheus/common/model"
	"golang.org/x/sync/errgroup"
)

// A MonthlyQuery is a RangeQuery that operates on monthly data,
// handling the fact that this means the "step" function is
// dynamic based on the length of the month.
type MonthlyQuery interface {
	MetricsQuery
	Start(t timex.MonthYear) MonthlyQuery
	End(t timex.MonthYear) MonthlyQuery
	MaxParallel(n int) MonthlyQuery
}

func (c *client) MonthlyQuery(q string) MonthlyQuery {
	return monthlyQuery{
		c: c,
		q: q,
	}
}

type monthlyQuery struct {
	q           string
	c           *client
	start       timex.MonthYear
	end         timex.MonthYear
	maxParallel int
}

func (q monthlyQuery) Start(t timex.MonthYear) MonthlyQuery {
	q.start = t
	return q
}

func (q monthlyQuery) End(t timex.MonthYear) MonthlyQuery {
	q.end = t
	return q
}

func (q monthlyQuery) MaxParallel(n int) MonthlyQuery {
	q.maxParallel = n
	return q
}

func (q monthlyQuery) Do(ctx context.Context) (*Result, error) {
	var (
		month     = q.start
		numMonths = monthsBetween(q.start, q.end)
		idx       int

		eg errgroup.Group
	)

	maxParallel := q.maxParallel
	if maxParallel == 0 {
		maxParallel = numMonths
	}

	eg.SetLimit(maxParallel)

	monthlyResults := make([]*Result, numMonths)
	for !month.After(q.end) {
		var (
			idxForResults = idx
			monthStart    = month.MonthStart().DayStart()
			monthEnd      = month.MonthEnd().DayEnd()
			daysBetween   = monthEnd.Sub(monthStart)
		)

		eg.Go(func() error {
			rq, err := q.c.RangeQuery(q.q).
				Start(monthStart).
				End(monthEnd).
				Step(model.Duration(daysBetween)).
				Do(ctx)
			if err != nil {
				return err
			}

			monthlyResults[idxForResults] = rq
			return nil
		})

		idx = idx + 1
		month = month.NextMonth()
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var (
		results  model.Matrix
		warnings []string
	)
	for _, r := range monthlyResults {
		results = append(results, r.Data.(model.Matrix)...)
		if len(r.Warnings) != 0 {
			warnings = append(warnings, r.Warnings...)
		}
	}

	return &Result{
		Data:     results,
		Warnings: warnings,
	}, nil
}

func monthsBetween(start, end timex.MonthYear) int {
	if end.Less(start) {
		end, start = start, end
	}

	var (
		startAsMonths = (start.Year * 12) + int(start.Month) - 1
		fromAsMonths  = (end.Year * 12) + int(end.Month) - 1
	)

	numMonths := startAsMonths - fromAsMonths
	if numMonths < 0 {
		numMonths = -numMonths
	}
	numMonths = numMonths + 1
	return numMonths
}
