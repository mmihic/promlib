package prom

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/mmihic/golib/src/pkg/cli"
)

// LabelQuery pulls labels matching a set of selectors.
type LabelQuery struct {
	BaseCommand
	Start string   `help:"start date for the query"`
	End   string   `help:"end date for the query"`
	Sel   []string `help:"query to run"`
}

func (cmd *LabelQuery) Run(ctx context.Context) error {
	c, err := cmd.PromClient(ctx)
	if err != nil {
		return err
	}

	q := c.LabelQuery()
	if len(cmd.Start) != 0 {
		start, err := parseTime(cmd.Start)
		if err != nil {
			return fmt.Errorf("invalid start: %w", err)
		}

		q = q.Start(start)
	}

	if len(cmd.End) != 0 {
		end, err := parseTime(cmd.End)
		if err != nil {
			return fmt.Errorf("invalid end: %w", err)
		}

		q = q.End(end)
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
