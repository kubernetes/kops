package table

import "time"

// This is just a bunch of data type checks, so... no linting here
//
//nolint:cyclop
func asInt(data interface{}) (int64, bool) {
	switch val := data.(type) {
	case int:
		return int64(val), true

	case int8:
		return int64(val), true

	case int16:
		return int64(val), true

	case int32:
		return int64(val), true

	case int64:
		return val, true

	case uint:
		return int64(val), true

	case uint8:
		return int64(val), true

	case uint16:
		return int64(val), true

	case uint32:
		return int64(val), true

	case uint64:
		return int64(val), true

	case time.Duration:
		return int64(val), true

	case StyledCell:
		return asInt(val.Data)
	}

	return 0, false
}

func asNumber(data interface{}) (float64, bool) {
	switch val := data.(type) {
	case float32:
		return float64(val), true

	case float64:
		return val, true

	case StyledCell:
		return asNumber(val.Data)
	}

	intVal, isInt := asInt(data)

	return float64(intVal), isInt
}
