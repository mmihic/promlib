package prom

import (
	"context"

	"github.com/mmihic/promlib/src/pkg/prom/promcli"
)

// RangeQuery runs a range query.
type RangeQuery struct {
	BaseCommand

	Start promcli.Time     `help:"start date for the query"`
	End   promcli.Time     `help:"end date for the query"`
	Step  promcli.Duration `help:"step function"`
	Query string           `short:"q" help:"query to run" required:""`
}

// Run runs the command.
func (cmd *RangeQuery) Run(ctx context.Context) error {
	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := client.RangeQuery(cmd.Query).
		Start(cmd.Start.AsTime()).
		End(cmd.End.AsTime())

	if cmd.Step != 0 {
		q = q.Step(cmd.Step.AsDuration())
	}

	result, err := q.Do(ctx)
	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}
