package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/fireeye/gocrack/shared"
)

// CLDevices is a list of OpenCL device ID's that a task will use on a specific host
// Note: One might ask "Schmitt, why is the type saving the integer array as a json array and not a postgres integer[]?"
// The main reason behind this is time. While the pg driver does have a type for an []int64, the support for it in DBR as well as the overall sql
// interface in go doesn't play well with it.
type CLDevices []int

func (s CLDevices) String() string {
	return shared.IntSliceToString(s)
}

// Value implements driver.Value
func (s CLDevices) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(&s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements the sql.Scanner interface,
func (s *CLDevices) Scan(src interface{}) (err error) {
	switch src.(type) {
	case []byte:
		return json.Unmarshal(src.([]byte), &s)
	case nil:
		return nil
	default:
		return errors.New("bad type assertion")
	}
}
