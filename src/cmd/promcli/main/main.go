package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/alecthomas/kong"

	"promlib/src/cmd/promcli/internal/cmd/prom"
)

type Commands struct {
	InstantQuery prom.InstantQuery `cmd:"" help:"runs an instant query"`
	RangeQuery   prom.RangeQuery   `cmd:"" help:"runs a range query"`
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