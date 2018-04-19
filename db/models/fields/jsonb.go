package fields

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

// JSONB ...
type JSONB []byte

// Value ...
func (j JSONB) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

// Scan ...
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("Scan source was not string")
	}
	// I think I need to make a copy of the bytes.
	// It seems the byte slice passed in is re-used
	*j = append((*j)[0:0], s...)

	return nil
}

// MarshalJSON returns *j as the JSON encoding of j.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// IsNull ...
func (j JSONB) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

// Equals ...
func (j JSONB) Equals(j1 JSONB) bool {
	return bytes.Equal([]byte(j), []byte(j1))
}
