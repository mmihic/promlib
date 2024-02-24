package fakeprom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryEqual(t *testing.T) {
	for _, tt := range []struct {
		name        string
		q1, q2      string
		expected    bool
		expectedErr string
	}{
		{
			"simple query",
			"sum ( up )", "sum(up)",
			true, "",
		},
		{
			"simple non-equal query",
			"sum(up)", "avg(up)",
			false, "",
		},
		{
			"invalid first query",
			"sum up", "sum(up)",
			false, "first query is invalid: 1:5: parse error",
		},
		{
			"invalid first query",
			"sum(up)", "avg up",
			false, "second query is invalid: 1:5: parse error",
		},
		{
			"complex equal query",
			`
sum by(cluster,namespace) (1 - (((0 * container_memory_usage_bytes{
        cluster=~"(production-a|production-b)", pod=~".*cluster.*rep.*", namespace=~"foo"
    })
    + on(instance) group_left node_memory_MemAvailable_bytes{job="node-exporter"}) /
       ((0 * container_memory_usage_bytes{
            cluster=~"(production-a|production-b)",
            pod=~".*cluster.*rep.*",
            namespace=~"foo"
       }) + on(instance) group_left node_memory_MemTotal_bytes{job="node-exporter"})
    ))`,
			`
sum by(cluster,namespace) (1 - (
    ((0 * container_memory_usage_bytes{
        cluster=~"(production-a|production-b)",
        pod=~".*cluster.*rep.*",
        namespace=~"foo"
    })
    +
    on(instance) group_left node_memory_MemAvailable_bytes{job="node-exporter"})
        /
       ((0 * container_memory_usage_bytes{
            cluster=~"(production-a|production-b)",
            pod=~".*cluster.*rep.*",
            namespace=~"foo"
       }) + on(instance) group_left node_memory_MemTotal_bytes{job="node-exporter"})
    ))`,
			true, "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := QueryEqual(tt.q1, tt.q2)
			if tt.expectedErr != "" {
				if !assert.Error(t, actualErr) {
					return
				}

				assert.Contains(t, actualErr.Error(), tt.expectedErr)
				return
			}

			if !assert.NoError(t, actualErr) {
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
