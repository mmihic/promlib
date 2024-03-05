package prom

import (
	"context"
	"time"

	"github.com/prometheus/common/model"

	"github.com/mmihic/promlib/src/pkg/prom/promcli"
)

// RangeQuery runs a range query.
type RangeQuery struct {
	BaseCommand

	Start promcli.Time     `help:"start date for the query" required:""`
	End   promcli.Time     `help:"end date for the query" required:""`
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
