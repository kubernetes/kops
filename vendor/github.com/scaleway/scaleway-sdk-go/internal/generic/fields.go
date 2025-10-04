package generic

import "reflect"

// HasField returns true if given struct has a field with given name
// Also allow a slice, it will use the underlying type
func HasField(i any, fieldName string) bool {
	value := reflect.Indirect(reflect.ValueOf(i))
	typ := value.Type()

	if value.Kind() == reflect.Slice {
		typ = indirectType(typ.Elem())
	}

	_, fieldExists := typ.FieldByName(fieldName)
	return fieldExists
}
