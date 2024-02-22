package prom

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/mmihic/golib/src/pkg/cli"
	"promlib/src/pkg/prom"
	"promlib/src/pkg/prom/promcli"
)

// BaseCommand is the base command for all prom CLI options.
type BaseCommand struct {
	cli.FormattedOutput
	cli.WithLogger
	promcli.ClientOptions
}

// WriteResult writes the results of a query.
func (cmd *BaseCommand) WriteResult(result *prom.Result) error {
	if cmd.Format != cli.FormatCSV {
		return cmd.WriteOutput(result)
	}

	var headers = []string{"metric", "timestamp", "value"}
	return cmd.WriteOutput(func(w io.Writer) error {
		csvw := csv.NewWriter(w)
		defer csvw.Flush()
		if err := csvw.Write(headers); err != nil {
			return err
		}

		iter := result.ValueIter()
		for iter.Next() {
			if err := csvw.Write([]string{
				iter.Metric().String(),
				iter.Timestamp().Format(time.RFC3339),
				iter.StringValue(),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}
