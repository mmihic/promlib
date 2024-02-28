package prom

import (
	"encoding/json"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_Vector(t *testing.T) {
	const asJSON = `
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {"cluster": "foosball", "foo": "bar", "namespace": "zen"},
        "value": [
          1708028165.516,
          "92905"
        ]
      }
    ]
  }
}`

	var result Result
	err := json.Unmarshal([]byte(asJSON), &result)
	require.NoError(t, err)

	assert.Equal(t, Result{
		Data: model.Vector{
			&model.Sample{
				Metric: model.Metric{
					"cluster":   "foosball",
					"foo":       "bar",
					"namespace": "zen",
				},
				Timestamp: 1708028165516,
				Value:     92905,
			},
		},
	}, result)
}
