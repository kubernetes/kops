package schema

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
)

// IDOrName can be used in API requests where either a resource id or name can be
// specified.
type IDOrName struct {
	ID   int64
	Name string
}

var _ json.Unmarshaler = (*IDOrName)(nil)
var _ json.Marshaler = (*IDOrName)(nil)

func (o IDOrName) MarshalJSON() ([]byte, error) {
	if o.ID != 0 {
		return json.Marshal(o.ID)
	}
	if o.Name != "" {
		return json.Marshal(o.Name)
	}

	// We want to preserve the behavior of an empty interface{} to prevent breaking
	// changes (marshaled to null when empty).
	return json.Marshal(nil)
}

func (o *IDOrName) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	// This ensures we won't lose precision on large IDs, see json.Number below
	d.UseNumber()

	var v any
	if err := d.Decode(&v); err != nil {
		return err
	}

	switch typed := v.(type) {
	case string:
		id, err := strconv.ParseInt(typed, 10, 64)
		if err == nil {
			o.ID = id
		} else if typed != "" {
			o.Name = typed
		}
	case json.Number:
		id, err := typed.Int64()
		if err != nil {
			return &json.UnmarshalTypeError{
				Value: string(data),
				Type:  reflect.TypeOf(*o),
			}
		}
		o.ID = id
	default:
		return &json.UnmarshalTypeError{
			Value: string(data),
			Type:  reflect.TypeOf(*o),
		}
	}

	return nil
}
