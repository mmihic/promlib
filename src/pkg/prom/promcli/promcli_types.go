package promcli

import (
	"encoding"
	"errors"
	"fmt"
	"time"

	"github.com/mmihic/golib/src/pkg/timex"
	"github.com/prometheus/common/model"
)

// Time is the CLI representation of Time, allowing for absolute and relative times.
type Time time.Time

// UnmarshalText unmarshals the text format of a time.
func (tm *Time) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return errors.New("invalid time ''")
	}

	txt := string(b)

	// Parse as absolute time
	t, err := time.Parse(time.RFC3339, txt)
	if err == nil {
		*tm = Time(t)
		return nil
	}

	// Parse as a date
	date, err := timex.ParseDate(txt)
	if err == nil {
		*tm = Time(date.DayStart())
		return nil
	}

	// Parse as month
	month, err := timex.ParseMonthYear(txt)
	if err == nil {
		*tm = Time(month.MonthStart().DayStart())
		return nil
	}

	// Parse as relative time
	multiplier := 1
	switch txt[0] {
	case '-':
		multiplier = -1
		txt = txt[1:]
	case '+':
		txt = txt[1:]
	}

	d, err := model.ParseDuration(txt)
	if err == nil {
		*tm = Time(time.Now().Add(time.Duration(d) * time.Duration(multiplier)))
		return nil
	}

	return fmt.Errorf("invalid time '%s'", txt)
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
