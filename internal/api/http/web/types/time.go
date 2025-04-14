package types

import (
	"errors"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

// Time is a custom time type that implements JSON serialization with a specific format.
type Time time.Time

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+len(`""`))
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("Time.UnmarshalJSON: input is not a JSON string")
	}
	data = data[len(`"`) : len(data)-len(`"`)]
	parsed, err := time.Parse(timeFormat, string(data))
	if err != nil {
		return errors.New("Time.UnmarshalJSON: " + err.Error())
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
