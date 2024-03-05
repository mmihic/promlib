package prom

import (
	"context"

	"github.com/mmihic/promlib/src/pkg/prom/promcli"
)

// InstantQuery runs an instant query.
type InstantQuery struct {
	BaseCommand
	Time  promcli.Time `help:"the time to query, defaults to now"`
	Query string       `short:"q" required:"" help:"query to run"`
}

// Run runs the command.
func (cmd *InstantQuery) Run(ctx context.Context) error {
	client, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := client.InstantQuery(cmd.Query)
	if !cmd.Time.AsTime().IsZero() {
		q = q.Time(cmd.Time.AsTime())
	}

	result, err := q.Do(ctx)
	if err != nil {
		return err
	}

	return cmd.WriteResult(result)
}
