package parameter

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"time"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// AddToQuery add a key/value pair to an URL query
func AddToQuery(query url.Values, key string, value any) {
	elemValue := reflect.ValueOf(value)

	if elemValue.Kind() == reflect.Invalid || elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
		return
	}

	for elemValue.Kind() == reflect.Ptr {
		elemValue = reflect.ValueOf(value).Elem()
	}

	elemType := elemValue.Type()
	switch {
	case elemType == reflect.TypeOf(net.IP{}):
		query.Add(key, value.(*net.IP).String())
	case elemType == reflect.TypeOf(net.IPNet{}):
		query.Add(key, value.(*net.IPNet).String())
	case elemType == reflect.TypeOf(scw.IPNet{}):
		query.Add(key, value.(*scw.IPNet).String())
	case elemType.Kind() == reflect.Slice:
		for i := 0; i < elemValue.Len(); i++ {
			query.Add(key, fmt.Sprint(elemValue.Index(i).Interface()))
		}
	case elemType == reflect.TypeOf(time.Time{}):
		query.Add(key, value.(*time.Time).Format(time.RFC3339))
	default:
		query.Add(key, fmt.Sprint(elemValue.Interface()))
	}
}
