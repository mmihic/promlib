package prom

import (
	"testing"
	"time"

	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

type value struct {
	name      string
	timestamp string
	value     string
}

func collect(iter ValueIter) []value {
	var values []value
	for iter.Next() {
		values = append(values, value{
			name:      iter.Metric().String(),
			timestamp: iter.Timestamp().UTC().Format(time.RFC3339),
			value:     iter.StringValue(),
		})
	}

	return values
}

func TestValueIter_Matrix_Empty(t *testing.T) {
	iter := NewValueIter(model.Matrix{})
	require.False(t, iter.Next())
}

func TestValueIter_Matrix_NoSamples(t *testing.T) {
	iter := NewValueIter(model.Matrix{
		{
			Metric: model.Metric{
				"foo": "bar",
			},
		},
	})
	require.False(t, iter.Next())
}

func TestValueIter_Matrix(t *testing.T) {
	iter := NewValueIter(model.Matrix{
		{
			Metric: model.Metric{
				"foo":     "bar",
				"cluster": "muster",
			},
			Values: []model.SamplePair{
				{
					Timestamp: mustParseTime("2022-05-19T13:45:16Z"),
					Value:     3.145,
				},
				{
					Timestamp: mustParseTime("2022-05-19T13:46:10Z"),
					Value:     4.26,
				},
			},
		},
		{
			Metric: model.Metric{
				"foo":       "bar",
				"cluster":   "foosball",
				"namespace": "zen",
			},
			Values: []model.SamplePair{
				{
					Timestamp: mustParseTime("2022-05-19T13:45:16Z"),
					Value:     19.7434,
				},
				{
					Timestamp: mustParseTime("2022-05-19T13:46:10Z"),
					Value:     21.561,
				},
				{
					Timestamp: mustParseTime("2022-05-19T13:47:14Z"),
					Value:     17.891,
				},
			},
		},
	})

	actual := collect(iter)
	require.Equal(t, actual, []value{
		{
			name:      `{cluster="muster", foo="bar"}`,
			value:     "3.145",
			timestamp: "2022-05-19T13:45:16Z",
		},
		{
			name:      `{cluster="muster", foo="bar"}`,
			value:     "4.26",
			timestamp: "2022-05-19T13:46:10Z",
		},
		{
			name:      `{cluster="foosball", foo="bar", namespace="zen"}`,
			value:     "19.7434",
			timestamp: "2022-05-19T13:45:16Z",
		},
		{
			name:      `{cluster="foosball", foo="bar", namespace="zen"}`,
			value:     "21.561",
			timestamp: "2022-05-19T13:46:10Z",
		},
		{
			name:      `{cluster="foosball", foo="bar", namespace="zen"}`,
			value:     "17.891",
			timestamp: "2022-05-19T13:47:14Z",
		},
	})
}

func TestValueIter_Vector_Empty(t *testing.T) {
	iter := NewValueIter(model.Vector{})
	require.False(t, iter.Next())
}

func TestValueIter_Vector(t *testing.T) {
	iter := NewValueIter(model.Vector{
		{
			Metric: model.Metric{
				"foo":       "bar",
				"cluster":   "foosball",
				"namespace": "zen",
			},
			Timestamp: mustParseTime("2022-05-19T13:45:16Z"),
			Value:     20438.45,
		},
		{
			Metric: model.Metric{
				"foo":     "bar",
				"cluster": "muster",
			},
			Timestamp: mustParseTime("2022-05-19T13:46:16Z"),
			Value:     3.145,
		},
	})

	actual := collect(iter)
	require.Equal(t, actual, []value{
		{
			name:      `{cluster="foosball", foo="bar", namespace="zen"}`,
			timestamp: "2022-05-19T13:45:16Z",
			value:     "20438.45",
		},
		{
			name:      `{cluster="muster", foo="bar"}`,
			timestamp: "2022-05-19T13:46:16Z",
			value:     "3.145",
		},
	})
}

func TestValueIter_Scalar(t *testing.T) {
	iter := NewValueIter(&model.Scalar{
		Timestamp: mustParseTime("2022-05-19T13:45:16Z"),
		Value:     32994.9034,
	})

	values := collect(iter)
	require.Equal(t, values, []value{
		{
			name:      "{}",
			timestamp: "2022-05-19T13:45:16Z",
			value:     "32994.9034",
		},
	})
}

func TestValueIter_String(t *testing.T) {
	iter := NewValueIter(&model.String{
		Timestamp: mustParseTime("2022-05-19T13:45:16Z"),
		Value:     "muzak",
	})

	values := collect(iter)
	require.Equal(t, values, []value{
		{
			name:      "{}",
			timestamp: "2022-05-19T13:45:16Z",
			value:     "muzak",
		},
	})
}

func TestValueIter_Empty(t *testing.T) {
	iter := NewValueIter(nil)
	require.False(t, iter.Next())
}

func mustParseTime(s string) model.Time {
	return model.Time(timex.MustParseTime(time.RFC3339, s).UnixMilli())
}
