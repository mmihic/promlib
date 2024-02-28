package fakeprom

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mmihic/golib/src/pkg/set"
	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"promlib/src/pkg/prom"
)

func TestFakeProm_RangeQuery(t *testing.T) {
	c := requireTestClient(t)

	for _, tt := range []struct {
		name        string
		q           RangeQuery
		expected    *prom.Result
		expectedErr string
	}{
		{
			"matches all properties",
			RangeQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Query:     "sum ( up )", // matches but not identical
			},
			&prom.Result{
				Data: model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"cluster":   "zed",
							"namespace": "bar",
						},
						Timestamp: 1708028165516,
						Value:     2930,
					},
				},
			}, "",
		},

		{
			"matches query only",
			RangeQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-07T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-13T00:36:15Z"),
				Query:     "sum ( up )", // matches but not identical
			},
			&prom.Result{
				Data: model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"cluster":   "zed",
							"namespace": "mork",
						},
						Timestamp: 1708028165516,
						Value:     92905,
					},
				},
			}, "",
		},

		{
			"returns an error",
			RangeQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Query:     "avg ( up )",
			},
			nil, "this was an error",
		},

		{
			"does not match any rules",
			RangeQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Query:     "min ( up )",
			},
			nil, "status_code: 404, msg=matcher not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := c.RangeQuery(tt.q.Query).
				Start(tt.q.StartTime).
				End(tt.q.EndTime).
				Step(model.Duration(time.Minute * 1)).
				Do(context.TODO())

			if tt.expectedErr != "" {
				if !assert.Error(t, err) {
					return
				}

				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}

func TestFakeProm_InstantQuery(t *testing.T) {
	c := requireTestClient(t)

	for _, tt := range []struct {
		name        string
		q           InstantQuery
		expected    *prom.Result
		expectedErr string
	}{
		{
			"matches all properties",
			InstantQuery{
				When:  timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				Query: "sum ( up )", // matches but not identical
			},
			&prom.Result{
				Data: model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"cluster":   "zed",
							"namespace": "bar",
						},
						Timestamp: 1708028165516,
						Value:     2930,
					},
				},
			}, "",
		},

		{
			"matches query only",
			InstantQuery{
				When:  timex.MustParseTime(time.RFC3339, "2023-04-07T00:35:15Z"),
				Query: "sum ( up )", // matches but not identical
			},
			&prom.Result{
				Data: model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"cluster":   "zed",
							"namespace": "mork",
						},
						Timestamp: 1708028165516,
						Value:     92905,
					},
				},
			}, "",
		},

		{
			"returns an error",
			InstantQuery{
				When:  timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				Query: "avg ( up )",
			},
			nil, "this is an error",
		},

		{
			"does not match any rules",
			InstantQuery{
				When:  timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Query: "min ( up )",
			},
			nil, "status_code: 404, msg=matcher not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := c.InstantQuery(tt.q.Query).
				Time(tt.q.When).
				Do(context.TODO())

			if tt.expectedErr != "" {
				if !assert.Error(t, err) {
					return
				}

				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}

func TestFakeProm_LabelQuery(t *testing.T) {
	c := requireTestClient(t)

	for _, tt := range []struct {
		name        string
		q           LabelQuery
		expected    []string
		expectedErr string
	}{
		{
			"matching start and end time",
			LabelQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Sels:      set.NewSet("up", "down"),
			},
			[]string{"foo", "bar"},
			"",
		},

		{
			"non-matching start and end time",
			LabelQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:40:15Z"),
				Sels:      set.NewSet("up", "down"),
			},
			[]string{"zed", "med"},
			"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := c.LabelQuery().
				Start(tt.q.StartTime).
				End(tt.q.EndTime).
				Selectors(tt.q.Sels.All()).
				Do(context.TODO())

			if tt.expectedErr != "" {
				if !assert.Error(t, err) {
					return
				}

				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}

func TestFakeProm_SeriesQuery(t *testing.T) {
	c := requireTestClient(t)

	for _, tt := range []struct {
		name        string
		q           SeriesQuery
		expected    []model.LabelSet
		expectedErr string
	}{
		{
			"matching start and end time",
			SeriesQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:36:15Z"),
				Sels:      set.NewSet("up", "down"),
			},
			[]model.LabelSet{
				{"cluster": "foo", "namespace": "bar"},
				{"cluster": "boo"},
			},
			"",
		},

		{
			"non-matching start and end time",
			SeriesQuery{
				StartTime: timex.MustParseTime(time.RFC3339, "2023-04-06T00:35:15Z"),
				EndTime:   timex.MustParseTime(time.RFC3339, "2023-04-06T00:40:15Z"),
				Sels:      set.NewSet("up", "down"),
			},
			[]model.LabelSet{
				{"cluster": "ken", "namespace": "blend"},
			},
			"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := c.SeriesQuery().
				Start(tt.q.StartTime).
				End(tt.q.EndTime).
				Selectors(tt.q.Sels.All()).
				Do(context.TODO())

			if tt.expectedErr != "" {
				if !assert.Error(t, err) {
					return
				}

				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}

func requireTestClient(t *testing.T) Client {
	f, err := os.Open("testdata/test-rules.yaml")
	require.NoError(t, err)

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var rules Rules
	err = dec.Decode(&rules)
	require.NoError(t, err)

	c := NewClientWithRules(&rules)
	return c
}
