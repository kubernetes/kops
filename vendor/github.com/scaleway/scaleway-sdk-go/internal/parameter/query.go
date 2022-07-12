package parameter

import (
	"fmt"
	"net/url"
	"reflect"
	"time"
)

// AddToQuery add a key/value pair to an URL query
func AddToQuery(query url.Values, key string, value interface{}) {
	elemValue := reflect.ValueOf(value)

	if elemValue.Kind() == reflect.Invalid || elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
		return
	}

	for elemValue.Kind() == reflect.Ptr {
		elemValue = reflect.ValueOf(value).Elem()
	}

	elemType := elemValue.Type()
	switch {
	case elemType.Kind() == reflect.Slice:
		for i := 0; i < elemValue.Len(); i++ {
			query.Add(key, fmt.Sprint(elemValue.Index(i).Interface()))
		}
	case elemType == reflect.TypeOf(time.Time{}):
		query.Add(key, value.(time.Time).Format(time.RFC3339))
	default:
		query.Add(key, fmt.Sprint(elemValue.Interface()))
	}
}
