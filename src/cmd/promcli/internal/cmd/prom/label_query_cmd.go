package prom

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/mmihic/golib/src/pkg/cli"

	"github.com/mmihic/promlib/src/pkg/prom/promcli"
)

// LabelQuery pulls labels matching a set of selectors.
type LabelQuery struct {
	BaseCommand
	Start promcli.Time `help:"start date for the query"`
	End   promcli.Time `help:"end date for the query"`
	Sel   []string     `help:"query to run"`
}

func (cmd *LabelQuery) Run(ctx context.Context) error {
	c, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := c.LabelQuery()
	if !cmd.Start.AsTime().IsZero() {
		q = q.Start(cmd.Start.AsTime())
	}

	if !cmd.End.AsTime().IsZero() {
		q = q.End(cmd.End.AsTime())
	}

	q = q.Selectors(cmd.Sel)
	results, err := q.Do(ctx)
	if err != nil {
		return err
	}

	if cmd.Format != cli.FormatCSV {
		return cmd.WriteOutput(results)
	}

	var headers = []string{"label"}
	return cmd.WriteOutput(func(w io.Writer) error {
		csvw := csv.NewWriter(w)
		defer csvw.Flush()
		if err := csvw.Write(headers); err != nil {
			return err
		}

		for _, label := range results {
			if err := csvw.Write([]string{label}); err != nil {
				return err
			}
		}

		return nil
	})
}
