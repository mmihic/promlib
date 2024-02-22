// Package promcli contains helpers for writing CLIs that use prom.
package promcli

import (
	"context"
	"errors"
	"os"

	"promlib/src/pkg/prom"
)

// ClientOptions are options for creating a client on the command line.
type ClientOptions struct {
	APITokenFile string `help:"file containing the API token"`
	ServerURL    string `name:"server-url" short:"s" required:"" help:"Prometheus server URL"`
}

// PromClient returns a prom PromClient.
func (opts *ClientOptions) PromClient(_ context.Context) (prom.Client, error) {
	var apiToken string

	if len(opts.APITokenFile) != 0 {
		apiTokenBytes, err := os.ReadFile(opts.APITokenFile)
		if err != nil {
			return nil, err
		}

		apiToken = string(apiTokenBytes)
	}

	if len(apiToken) == 0 {
		apiToken = os.Getenv("PROM_API_TOKEN")
	}

	if len(apiToken) == 0 {
		return nil, errors.New("neither --api-token-file nor PROM_API_TOKEN env var are set")
	}

	return prom.NewClient(opts.ServerURL,
		prom.WithHTTPOptions(prom.WithHeader("API-Token", apiToken)))
}
