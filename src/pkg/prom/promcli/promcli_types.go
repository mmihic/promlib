package promcli

import (
	"encoding"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

// Time is the CLI representation of Time, allowing for absolute and relative times.
type Time time.Time

// UnmarshalText unmarshals the text format of a time.
func (tm *Time) UnmarshalText(b []byte) error {
	txt := string(b)

	// Parse as absolute time
	t, err := time.Parse(time.RFC3339, txt)
	if err == nil {
		*tm = Time(t)
		return nil
	}

	// Parse as relative time
	d, err := model.ParseDuration(txt)
	if err == nil {
		*tm = Time(time.Now().Add(-time.Duration(d)))
		return nil
	}

	return fmt.Errorf("invalid time %s", txt)
}

// AsTime converts the Time to time.Time
func (tm Time) AsTime() time.Time {
	return time.Time(tm)
}

// Duration is the CLI representation of Duration.
type Duration model.Duration

// UnmarshalText unmarshals the text format of a time.
func (d *Duration) UnmarshalText(b []byte) error {
	txt := string(b)

	parsed, err := model.ParseDuration(txt)
	if err != nil {
		return fmt.Errorf("invalid duration %s: %w", txt, err)
	}

	*d = Duration(parsed)
	return nil
}

// AsDuration converts the Duration to model.Duration.
func (d Duration) AsDuration() model.Duration {
	return model.Duration(d)
}

var (
	_ encoding.TextUnmarshaler = &Time{}
	_ encoding.TextUnmarshaler = (*Duration)(nil)
)
