package prom

import (
	"context"
	"fmt"
)

// InstantQuery runs an instant query.
type InstantQuery struct {
	BaseCommand
	Time  string `help:"the time to query, defaults to now"`
	Query string `short:"q" required:"" help:"query to run"`
}

// Run runs the command.
func (cmd *InstantQuery) Run(ctx context.Context) error {
	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := client.InstantQuery(cmd.Query)
	if len(cmd.Time) != 0 {
		tm, err := parseTime(cmd.Time)
		if err != nil {
			return fmt.Errorf("invalid time: %w", err)
		}

		q = q.Time(tm)
	}

	result, err := q.Do(ctx)
	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}
