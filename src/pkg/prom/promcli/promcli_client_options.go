// Package promcli contains helpers for writing CLIs that use prom.
package promcli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"promlib/src/pkg/prom"
)

const (
	baseURL = "https://%s.chronosphere.io/data/m3/"
)

// ClientOptions are options for creating a client on the command line.
type ClientOptions struct {
	APITokenFile string `help:"file containing the API token"`
	Domain       string `short:"d" default:"meta" required:"" help:"tenant to query"`
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
		apiToken = os.Getenv("CHRONO_API_TOKEN")
	}

	if len(apiToken) == 0 {
		return nil, errors.New("neither --api-token-file nor CHRONO_API_TOKEN env var are set")
	}

	return prom.NewClient(fmt.Sprintf(baseURL, opts.Domain),
		prom.WithHTTPOptions(prom.WithHeader("API-Token", apiToken)))
}
