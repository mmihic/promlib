package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const skipTest = true

func TestPromClient(t *testing.T) {
	if skipTest {
		return
	}
	apiToken := os.Getenv("API_TOKEN")
	baseURL := os.Getenv("BASE_URL")

	c, err := NewClient(baseURL, WithHTTPOptions(httplib.SetHeader("API-Token", apiToken)))

	require.NoError(t, err)

	results, err := c.RangeQuery("up").
		Start(time.Now().Add(-time.Minute * 30)).
		End(time.Now()).
		Do(context.TODO())
	if !assert.NoError(t, err) {
		if httperr, ok := httplib.UnwrapError(err); ok {
			fmt.Println(httperr.Body.String())
		}
	}

	b, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(b))
}
