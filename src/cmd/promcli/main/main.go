package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/alecthomas/kong"

	"github.com/mmihic/promlib/src/cmd/promcli/internal/cmd/prom"
)

type Commands struct {
	Instant prom.InstantQuery `cmd:"" help:"runs an instant query"`
	Range   prom.RangeQuery   `cmd:"" help:"runs a range query"`
	Monthly prom.MonthlyQuery `cmd:"" help:"runs a range query over months"`
	Series  prom.SeriesQuery  `cmd:"" help:"pulls series matching an optional set of selectors"`
	Labels  prom.LabelQuery   `cmd:"" help:"pulls label names matching an optional set of selectors"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	var cli Commands
	var k = kong.Parse(&cli,
		kong.Bind(logger),
		kong.BindTo(context.Background(), (*context.Context)(nil)))
	if err := k.Run(); err != nil {
		logger.Fatal("unable to run command", zap.Error(err))
	}
}
