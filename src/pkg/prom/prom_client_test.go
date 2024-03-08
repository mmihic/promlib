package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/mmihic/httplib/src/pkg/httplib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const skipTest = true

func TestPromClient_RangeQuery(t *testing.T) {
	if skipTest {
		return
	}
	c := requireClient(t)

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

func TestPromClient_MonthlyQuery(t *testing.T) {
	if skipTest {
		return
	}
	c := requireClient(t)

	results, err := c.MonthlyQuery("up").
		Start(timex.MustParseMonthYear("2023-12")).
		End(timex.MustParseMonthYear("2024-01")).
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

func requireClient(t *testing.T) Client {
	/*
		apiToken := os.Getenv("API_TOKEN")
		baseURL := os.Getenv("BASE_URL")
	*/
	apiToken := "cpt-1-c619c1774ee247fa70800f43f00ede66760ea2c7c4b726228ff0ea81fe42d206"
	baseURL := "https://meta.chronosphere.io/data/m3/"
	c, err := NewClient(baseURL, WithHTTPOptions(httplib.SetHeader("API-Token", apiToken)))
	require.NoError(t, err)
	return c
}
