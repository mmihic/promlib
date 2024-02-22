package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mmihic/golib/src/pkg/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromClient(t *testing.T) {
	c, err := NewClient("https://meta.chronosphere.io/data/m3/",
		WithHTTPOptions(WithHeader("API-Token", os.Getenv("CHRONO_API_TOKEN"))))

	require.NoError(t, err)

	results, err := c.RangeQuery("up").
		Start(time.Now().Add(-time.Minute * 30)).
		End(time.Now()).
		Do(context.TODO())
	if !assert.NoError(t, err) {
		if httperr, ok := httpclient.UnwrapError(err); ok {
			fmt.Println(httperr.Body.String())
		}
	}

	b, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(b))
}
