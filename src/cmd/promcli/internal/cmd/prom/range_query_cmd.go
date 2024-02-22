package prom

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

// RangeQuery runs a range query.
type RangeQuery struct {
	BaseCommand

	Start string `short:"s" help:"start date for the query" required:""`
	End   string `short:"e" help:"end date for the query" required:""`
	Step  string `short:"p" help:"step function"`
	Query string `short:"q" required:"" help:"query to run"`
}

// Run runs the command.
func (cmd *RangeQuery) Run(ctx context.Context) error {
	start, err := parseTime(cmd.Start)
	if err != nil {
		return fmt.Errorf("cannot parse 'start' as absolute or relative time: %w", err)
	}

	end, err := parseTime(cmd.End)
	if err != nil {
		return fmt.Errorf("cannot parse 'start' as absolute or relative time: %w", err)
	}

	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := client.RangeQuery(cmd.Query).
		Start(start).
		End(end)

	if cmd.Step != "" {
		step, err := model.ParseDuration(cmd.Step)
		if err != nil {
			return fmt.Errorf("cannot parse 'step': %w", err)
		}

		q.Step(step)
	}

	result, err := q.Do(ctx)
	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	backDuration, err := model.ParseDuration(s)
	if err == nil {
		return time.Now().Add(-time.Duration(backDuration)), nil
	}

	return time.Time{}, err
}
