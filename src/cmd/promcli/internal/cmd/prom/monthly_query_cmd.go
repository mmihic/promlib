package prom

import (
	"context"

	"github.com/mmihic/golib/src/pkg/timex"
)

// MonthlyQuery runs a range query over months.
type MonthlyQuery struct {
	BaseCommand

	Start timex.MonthYear `help:"start date for the query"`
	End   timex.MonthYear `help:"end date for the query"`
	Query string          `short:"q" help:"query to run" required:""`
}

// Run runs the command.
func (cmd *MonthlyQuery) Run(ctx context.Context) error {
	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.MonthlyQuery(cmd.Query).
		Start(cmd.Start).
		End(cmd.End).
		Do(ctx)
	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}
