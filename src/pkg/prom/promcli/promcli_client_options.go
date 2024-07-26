// Package promcli contains helpers for writing CLIs that use prom.
package promcli

import (
	"context"
	"errors"
	"fmt"
	"github.com/mmihic/golib/src/pkg/cli"
	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/mmihic/promlib/src/pkg/prom"
	"go.uber.org/zap"
)

const (
	chronoPrometheusURL = "https://%s.chronosphere.io/data/m3/"
)

// ClientOptions are options for creating a client on the command line.
type ClientOptions struct {
	PromAPITokenFile string `name:"prom-api-token-file" help:"file containing the API token"`
	PromServerURL    string `name:"prom-server-url" help:"Prometheus server URL"`
	SourceTenant     string `name:"source-tenant" help:"name of the Chronosphere tenant to query" default:"meta"`
	LogQueries       bool   `help:"set to log queries"`
	LogResponses     bool   `help:"set to log request/response bodies"`
}

// PromClient returns a prom PromClient.
func (opts *ClientOptions) PromClient(_ context.Context) (prom.Client, error) {
	apiToken, err := cli.ReadAPIToken(opts.PromAPITokenFile, "PROM_API_TOKEN")
	if err != nil {
		return nil, err
	}

	baseURL, err := opts.getBaseURL()
	if err != nil {
		return nil, err
	}

	clientOpts := []prom.ClientOpt{
		prom.WithHTTPOptions(httplib.SetHeader("API-Token", apiToken)),
	}

	if opts.LogQueries || opts.LogResponses {
		log, err := zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("unable to create production logger: %w", err)
		}

		clientOpts = append(clientOpts, prom.WithQueryLog(log, opts.LogResponses))
	}

	return prom.NewClient(baseURL, clientOpts...)
}

func (opts *ClientOptions) getBaseURL() (string, error) {
	var baseURL string
	switch {
	case len(opts.PromServerURL) != 0 && len(opts.SourceTenant) != 0:
		return "", errors.New("only one of --source-tenant or --prom-server-url must be specified")
	case len(opts.PromServerURL) != 0:
		baseURL = opts.PromServerURL
	case len(opts.SourceTenant) != 0:
		baseURL = fmt.Sprintf(chronoPrometheusURL, opts.SourceTenant)
	}

	if len(baseURL) == 0 {
		return "", errors.New("one of --tenant or --server-url must be specified")
	}
	return baseURL, nil
}
