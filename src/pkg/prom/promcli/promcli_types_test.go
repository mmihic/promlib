package promcli

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime_UnmarshalText_Absolute(t *testing.T) {
	var value struct {
		MyTime Time `json:"my-time"`
	}

	err := json.Unmarshal([]byte(`{"my-time": "2023-05-06T12:35:46Z"}`), &value)
	require.NoError(t, err)

	assert.Equal(t, value.MyTime, Time(timex.MustParseTime(time.RFC3339, "2023-05-06T12:35:46Z")))
}

func TestTime_UnmarshalText_Relative(t *testing.T) {
	var value struct {
		MyTime Time `json:"my-time"`
	}

	err := json.Unmarshal([]byte(`{"my-time": "30m"}`), &value)
	require.NoError(t, err)

	halfHourAgo := time.Now().Add(-time.Minute * 30)
	assert.True(t, value.MyTime.AsTime().Equal(halfHourAgo) || value.MyTime.AsTime().Before(halfHourAgo))
}

func TestTime_UnmarshalText_Invalid(t *testing.T) {
	var value struct {
		MyTime Time `json:"my-time"`
	}

	err := json.Unmarshal([]byte(`{"my-time": "blerg"}`), &value)
	require.Error(t, err)
	require.Contains(t, "invalid time blerg", err.Error())
}

func TestDuration_UnmarshalText(t *testing.T) {
	var value struct {
		MyDuration Duration `json:"my-duration"`
	}

	err := json.Unmarshal([]byte(`{"my-duration": "30m"}`), &value)
	require.NoError(t, err)
	require.Equal(t, Duration(time.Minute*30), value.MyDuration)
}
