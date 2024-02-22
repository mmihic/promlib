package prom

import (
	"context"
	"time"
)

// InstantQuery runs an instant query.
type InstantQuery struct {
	BaseCommand
	Time  time.Time `help:"the time to query, defaults to now"`
	Query string    `short:"q" required:"" help:"query to run"`
}

// Run runs the command.
func (cmd *InstantQuery) Run(ctx context.Context) error {
	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.InstantQuery(cmd.Query).
		Time(cmd.Time).
		Do(ctx)

	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}
