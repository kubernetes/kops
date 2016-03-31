package utils

import (
	"fmt"
	"github.com/golang/glog"
	"reflect"
	"strings"
)

// This file is (probably) the most complex code we have
// It populates an object's fields, from the values stored in a map
// We typically get this map by unmarshalling yaml or json
// So: why not just use go's built in unmarshalling?
// The reason is that we want richer functionality for when we link objects to each other
// By doing our own marshalling, we can then link objects just by specifying a string identifier.
// Then, while we're here, we add nicer functionality like case-insensitivity and nicer handling for resources.

// Unmarshaller implements our specialized marshalling from a map to an object
type Unmarshaller struct {
	SpecialCases UnmarshallerSpecialCaseHandler
}

// UnmarshallerSpecialCaseHandler is the function type that a handler for non-standard types must implement
type UnmarshallerSpecialCaseHandler func(name string, dest Settable, src interface{}, destTypeName string) (bool, error)

// Settable is a workaround for the fact that map entries are not settable
type Settable struct {
	Value reflect.Value

	MapValue reflect.Value
	MapKey   reflect.Value
}

// Set sets the target value to the specified value
func (s *Settable) Set(v reflect.Value) {
	if s.MapValue.IsValid() {
		s.MapValue.SetMapIndex(s.MapKey, v)
	} else {
		s.Value.Set(v)
	}
}

func (s *Settable) Type() reflect.Type {
	return s.Value.Type()
}

func (r *Unmarshaller) UnmarshalSettable(name string, dest Settable, src interface{}) error {
	if src == nil {
		return nil
	}

	// Divert special cases
	switch dest.Type().Kind() {
	case reflect.Map:
		return r.unmarshalMap(name, dest, src)
	}

	destTypeName := BuildTypeName(dest.Type())

	switch destTypeName {
	case "*string":
		{
			switch src := src.(type) {
			case string:
				v := src
				dest.Set(reflect.ValueOf(&v))
				return nil
			default:
				return fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, destTypeName)
			}
		}

	case "[]string":
		{
			switch src := src.(type) {
			case string:
				// We allow a single string to populate an array
				v := []string{src}
				dest.Set(reflect.ValueOf(v))
				return nil
			case []interface{}:
				v := []string{}
				for _, i := range src {
					v = append(v, fmt.Sprintf("%v", i))
				}
				dest.Set(reflect.ValueOf(v))
				return nil
			default:
				return fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, destTypeName)
			}
		}

	case "*int64":
		{
			switch src := src.(type) {
			case int:
				v := int64(src)
				dest.Set(reflect.ValueOf(&v))
				return nil
			case float64:
				v := int64(src)
				dest.Set(reflect.ValueOf(&v))
				return nil
			default:
				return fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, destTypeName)
			}
		}

	case "*bool":
		{
			switch src := src.(type) {
			case bool:
				v := bool(src)
				dest.Set(reflect.ValueOf(&v))
				return nil
			default:
				return fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, destTypeName)
			}
		}

	default:
		if r.SpecialCases != nil {
			handled, err := r.SpecialCases(name, dest, src, destTypeName)
			if err != nil {
				return err
			}
			if handled {
				return nil
			}
		}
		return fmt.Errorf("unhandled destination type for %q: %s", name, destTypeName)
	}
}

func (r *Unmarshaller) unmarshalMap(name string, dest Settable, src interface{}) error {
	if src == nil {
		return nil
	}

	glog.Infof("populateMap on type %s", BuildTypeName(dest.Type()))

	destType := dest.Type()

	if destType.Kind() != reflect.Map {
		glog.Errorf("expected map type, got %v", destType)
	}

	if dest.Value.IsNil() {
		m := reflect.MakeMap(dest.Type())
		dest.Set(m)
		dest = Settable{Value: m}
	}

	srcMap, ok := src.(map[string]interface{})
	if ok {
		for k, v := range srcMap {
			newValue := reflect.New(destType.Elem()).Elem()

			settable := Settable{
				Value:    newValue,
				MapValue: dest.Value,
				MapKey:   reflect.ValueOf(k),
			}
			settable.Set(newValue)
			err := r.UnmarshalSettable(name+"."+k, settable, v)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("unexpected source type for map %q: %T", name, src)
}

func (r *Unmarshaller) UnmarshalStruct(name string, dest reflect.Value, src interface{}) error {
	m, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected type of source data for %q: %T", name, src)
	}

	if dest.Kind() == reflect.Ptr && !dest.IsNil() {
		dest = dest.Elem()
	}

	if dest.Kind() != reflect.Struct {
		return fmt.Errorf("UnmarshalStruct called on non-struct: %v", dest.Kind())
	}

	// TODO: Pre-calculate / cache?
	destType := dest.Type()
	fieldMap := map[string]reflect.StructField{}
	for i := 0; i < destType.NumField(); i++ {
		f := destType.Field(i)
		fieldName := f.Name
		fieldName = strings.ToLower(fieldName)
		_, exists := fieldMap[fieldName]
		if exists {
			glog.Fatalf("ambiguous field name in %q: %q", destType.Name(), fieldName)
		}
		fieldMap[fieldName] = f
	}

	//t := dest.Type()
	for k, v := range m {
		k = strings.ToLower(k)
		fieldInfo, found := fieldMap[k]
		if !found {
			return fmt.Errorf("unknown field %q in %q", k, name)
		}
		field := dest.FieldByIndex(fieldInfo.Index)

		err := r.UnmarshalSettable(name+"."+k, Settable{Value: field}, v)
		if err != nil {
			return err
		}
	}

	return nil
}
