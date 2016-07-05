// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"fmt"
	"reflect"
	"time"

	pb "google.golang.org/cloud/datastore/internal/proto"
)

var (
	typeOfByteSlice = reflect.TypeOf([]byte(nil))
	typeOfTime      = reflect.TypeOf(time.Time{})
	typeOfGeoPoint  = reflect.TypeOf(GeoPoint{})
)

// typeMismatchReason returns a string explaining why the property p could not
// be stored in an entity field of type v.Type().
func typeMismatchReason(p Property, v reflect.Value) string {
	entityType := "empty"
	switch p.Value.(type) {
	case int64:
		entityType = "int"
	case bool:
		entityType = "bool"
	case string:
		entityType = "string"
	case float64:
		entityType = "float"
	case *Key:
		entityType = "*datastore.Key"
	case GeoPoint:
		entityType = "GeoPoint"
	case time.Time:
		entityType = "time.Time"
	case []byte:
		entityType = "[]byte"
	}

	return fmt.Sprintf("type mismatch: %s versus %v", entityType, v.Type())
}

type propertyLoader struct {
	// m holds the number of times a substruct field like "Foo.Bar.Baz" has
	// been seen so far. The map is constructed lazily.
	m map[string]int
}

func (l *propertyLoader) load(codec *structCodec, structValue reflect.Value, p Property, prev map[string]struct{}) string {
	sl, ok := p.Value.([]interface{})
	if !ok {
		return l.loadOneElement(codec, structValue, p, prev)
	}
	for _, val := range sl {
		p.Value = val
		if errStr := l.loadOneElement(codec, structValue, p, prev); errStr != "" {
			return errStr
		}
	}
	return ""
}

func (l *propertyLoader) loadOneElement(codec *structCodec, structValue reflect.Value, p Property, prev map[string]struct{}) string {
	var sliceOk bool
	var v reflect.Value
	// Traverse a struct's struct-typed fields.
	for name := p.Name; ; {
		decoder, ok := codec.byName[name]
		if !ok {
			return "no such struct field"
		}
		v = structValue.Field(decoder.index)
		if !v.IsValid() {
			return "no such struct field"
		}
		if !v.CanSet() {
			return "cannot set struct field"
		}

		if decoder.substructCodec == nil {
			break
		}

		// If the element is a slice, we need to accommodate it.
		if v.Kind() == reflect.Slice {
			if l.m == nil {
				l.m = make(map[string]int)
			}
			index := l.m[p.Name]
			l.m[p.Name] = index + 1
			for v.Len() <= index {
				v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
			}
			structValue = v.Index(index)
			sliceOk = true
		} else {
			structValue = v
		}
		// Strip the "I." from "I.X".
		name = name[len(codec.byIndex[decoder.index].name):]
		codec = decoder.substructCodec
	}

	var slice reflect.Value
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 {
		slice = v
		v = reflect.New(v.Type().Elem()).Elem()
	} else if _, ok := prev[p.Name]; ok && !sliceOk {
		// Zero the field back out that was set previously, turns out
		// it's a slice and we don't know what to do with it
		v.Set(reflect.Zero(v.Type()))
		return "multiple-valued property requires a slice field type"
	}

	prev[p.Name] = struct{}{}

	pValue := p.Value
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, ok := pValue.(int64)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		if v.OverflowInt(x) {
			return fmt.Sprintf("value %v overflows struct field of type %v", x, v.Type())
		}
		v.SetInt(x)
	case reflect.Bool:
		x, ok := pValue.(bool)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		v.SetBool(x)
	case reflect.String:
		x, ok := pValue.(string)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		v.SetString(x)
	case reflect.Float32, reflect.Float64:
		x, ok := pValue.(float64)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		if v.OverflowFloat(x) {
			return fmt.Sprintf("value %v overflows struct field of type %v", x, v.Type())
		}
		v.SetFloat(x)
	case reflect.Ptr:
		x, ok := pValue.(*Key)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		if _, ok := v.Interface().(*Key); !ok {
			return typeMismatchReason(p, v)
		}
		v.Set(reflect.ValueOf(x))
	case reflect.Struct:
		switch v.Type() {
		case typeOfTime:
			x, ok := pValue.(time.Time)
			if !ok && pValue != nil {
				return typeMismatchReason(p, v)
			}
			v.Set(reflect.ValueOf(x))
		case typeOfGeoPoint:
			x, ok := pValue.(GeoPoint)
			if !ok && pValue != nil {
				return typeMismatchReason(p, v)
			}
			v.Set(reflect.ValueOf(x))
		default:
			return typeMismatchReason(p, v)
		}
	case reflect.Slice:
		x, ok := pValue.([]byte)
		if !ok && pValue != nil {
			return typeMismatchReason(p, v)
		}
		if v.Type().Elem().Kind() != reflect.Uint8 {
			return typeMismatchReason(p, v)
		}
		v.SetBytes(x)
	default:
		return typeMismatchReason(p, v)
	}
	if slice.IsValid() {
		slice.Set(reflect.Append(slice, v))
	}
	return ""
}

// loadEntity loads an EntityProto into PropertyLoadSaver or struct pointer.
func loadEntity(dst interface{}, src *pb.Entity) (err error) {
	props := protoToProperties(src)
	if e, ok := dst.(PropertyLoadSaver); ok {
		return e.Load(props)
	}
	return LoadStruct(dst, props)
}

func (s structPLS) Load(props []Property) error {
	var fieldName, reason string
	var l propertyLoader

	prev := make(map[string]struct{})
	for _, p := range props {
		if errStr := l.load(s.codec, s.v, p, prev); errStr != "" {
			// We don't return early, as we try to load as many properties as possible.
			// It is valid to load an entity into a struct that cannot fully represent it.
			// That case returns an error, but the caller is free to ignore it.
			fieldName, reason = p.Name, errStr
		}
	}
	if reason != "" {
		return &ErrFieldMismatch{
			StructType: s.v.Type(),
			FieldName:  fieldName,
			Reason:     reason,
		}
	}
	return nil
}

func protoToProperties(src *pb.Entity) []Property {
	props := src.Properties
	out := make([]Property, 0, len(props))
	for name, val := range props {
		out = append(out, Property{
			Name:    name,
			Value:   propToValue(val),
			NoIndex: val.ExcludeFromIndexes,
		})
	}
	return out
}

// propToValue returns a Go value that represents the PropertyValue. For
// example, a TimestampValue becomes a time.Time.
func propToValue(v *pb.Value) interface{} {
	switch v := v.ValueType.(type) {
	case *pb.Value_NullValue:
		return nil
	case *pb.Value_BooleanValue:
		return v.BooleanValue
	case *pb.Value_IntegerValue:
		return v.IntegerValue
	case *pb.Value_DoubleValue:
		return v.DoubleValue
	case *pb.Value_TimestampValue:
		return time.Unix(v.TimestampValue.Seconds, int64(v.TimestampValue.Nanos))
	case *pb.Value_KeyValue:
		// TODO(djd): Don't drop this error.
		key, _ := protoToKey(v.KeyValue)
		return key
	case *pb.Value_StringValue:
		return v.StringValue
	case *pb.Value_BlobValue:
		return []byte(v.BlobValue)
	case *pb.Value_GeoPointValue:
		return GeoPoint{Lat: v.GeoPointValue.Latitude, Lng: v.GeoPointValue.Longitude}
	case *pb.Value_EntityValue:
		// TODO(djd): Support EntityValue.
		return nil
	case *pb.Value_ArrayValue:
		arr := make([]interface{}, 0, len(v.ArrayValue.Values))
		for _, v := range v.ArrayValue.Values {
			arr = append(arr, propToValue(v))
		}
		return arr
	default:
		return nil
	}
}
