// Package promcli contains helpers for writing CLIs that use prom.
package promcli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"go.uber.org/zap"

	"promlib/src/pkg/prom"
)

const (
	chronoPrometheusURL = "https://%s.chronosphere.io/data/m3/"
)

// ClientOptions are options for creating a client on the command line.
type ClientOptions struct {
	APITokenFile string `help:"file containing the API token"`
	ServerURL    string `name:"server-url" short:"s" help:"Prometheus server URL"`
	Tenant       string `name:"tenant" short:"t" help:"name of the Chronosphere tenant to query"`
	LogQueries   bool   `help:"set to log queries"`
}

// PromClient returns a prom PromClient.
func (opts *ClientOptions) PromClient(_ context.Context) (prom.Client, error) {
	apiToken, client, err := opts.getAPIToken()
	if err != nil {
		return client, err
	}

	baseURL, err := opts.getBaseURL()
	if err != nil {
		return nil, err
	}

	clientOpts := []prom.ClientOpt{
		prom.WithHTTPOptions(httplib.SetHeader("API-Token", apiToken)),
	}

	if opts.LogQueries {
		log, err := zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("unable to create production logger: %w", err)
		}
		clientOpts = append(clientOpts, prom.WithQueryLog(log))
	}

	return prom.NewClient(baseURL, clientOpts...)
}

func (opts *ClientOptions) getBaseURL() (string, error) {
	var baseURL string
	switch {
	case len(opts.ServerURL) != 0 && len(opts.Tenant) != 0:
		return "", errors.New("only one of --tenant or --server-url must be specified")
	case len(opts.ServerURL) != 0:
		baseURL = opts.ServerURL
	case len(opts.Tenant) != 0:
		baseURL = fmt.Sprintf(chronoPrometheusURL, opts.Tenant)
	}

	if len(baseURL) == 0 {
		return "", errors.New("one of --tenant or --server-url must be specified")
	}
	return baseURL, nil
}

func (opts *ClientOptions) getAPIToken() (string, prom.Client, error) {
	var apiToken string

	if len(opts.APITokenFile) != 0 {
		apiTokenBytes, err := os.ReadFile(opts.APITokenFile)
		if err != nil {
			return "", nil, err
		}

		apiToken = string(apiTokenBytes)
	}

	if len(apiToken) == 0 {
		apiToken = os.Getenv("PROM_API_TOKEN")
	}

	if len(apiToken) == 0 {
		return "", nil, errors.New("neither --api-token-file nor PROM_API_TOKEN env var are set")
	}
	return apiToken, nil, nil
}
