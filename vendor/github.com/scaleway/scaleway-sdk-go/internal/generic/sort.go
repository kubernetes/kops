package generic

import (
	"reflect"
	"sort"
)

// SortSliceByField sorts given slice of struct by passing the specified field to given compare function
// given slice must be a slice of Ptr
func SortSliceByField(list interface{}, field string, compare func(interface{}, interface{}) bool) {
	listValue := reflect.ValueOf(list)
	sort.SliceStable(list, func(i, j int) bool {
		field1 := listValue.Index(i).Elem().FieldByName(field).Interface()
		field2 := listValue.Index(j).Elem().FieldByName(field).Interface()
		return compare(field1, field2)
	})
}
