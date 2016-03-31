package fi

import (
	"fmt"
	"reflect"
)

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func String(s string) *string {
	return &s
}

func Bool(v bool) *bool {
	return &v
}

func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func Int(v int) *int {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func DebugPrint(o interface{}) string {
	if o == nil {
		return "<nil>"
	}
	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return "<?>"
	}
	o = v.Interface()
	stringer, ok := o.(fmt.Stringer)
	if ok {
		return stringer.String()
	}
	return fmt.Sprint(o)
}
