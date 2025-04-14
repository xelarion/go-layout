package types

import (
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

// Time is a custom time type that implements JSON serialization with a specific format.
type Time time.Time

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return nil
	}
	parsed, err := time.Parse(timeFormat, string(data[1:len(data)-1]))
	if err != nil {
		return err
	}
	*t = Time(parsed)
	return nil
}

// String returns the string representation of the time.
func (t Time) String() string {
	return time.Time(t).Format(timeFormat)
}

// Time returns the underlying time.Time value.
func (t Time) Time() time.Time {
	return time.Time(t)
}
