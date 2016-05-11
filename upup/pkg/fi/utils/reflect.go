package utils

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"reflect"
)

var SkipReflection = errors.New("skip this value")

// InvokeMethod calls the specified method by reflection
func InvokeMethod(target interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	v := reflect.ValueOf(target)

	method, found := v.Type().MethodByName(name)
	if !found {
		return nil, fmt.Errorf("method %q not found on %T", name, target)
	}

	var argValues []reflect.Value
	for _, a := range args {
		argValues = append(argValues, reflect.ValueOf(a))
	}
	glog.V(8).Infof("Calling method %s on %T", method.Name, target)
	m := v.MethodByName(method.Name)
	rv := m.Call(argValues)
	return rv, nil
}

func BuildTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + BuildTypeName(t.Elem())
	case reflect.Slice:
		return "[]" + BuildTypeName(t.Elem())
	case reflect.Struct, reflect.Interface:
		return t.Name()
	case reflect.String, reflect.Bool, reflect.Int64:
		return t.Name()
	case reflect.Map:
		return "map[" + BuildTypeName(t.Key()) + "]" + BuildTypeName(t.Elem())
	default:
		glog.Errorf("cannot find type name for: %v, assuming %s", t, t.Name())
		return t.Name()
	}
}

type visitorFunc func(path string, field *reflect.StructField, v reflect.Value) error

func WalkRecursive(v reflect.Value, visitor visitorFunc) error {
	return walkRecursive("", v, visitor)
}

func walkRecursive(path string, v reflect.Value, visitor visitorFunc) error {
	vType := v.Type()

	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Struct || e.Kind() == reflect.Map {
				v = e
				vType = v.Type()
			}
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			structField := vType.Field(i)
			if structField.PkgPath != "" {
				// Field not exported
				continue
			}

			f := v.Field(i)

			childPath := path + "." + structField.Name
			err := visitor(childPath, &structField, f)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = walkRecursive(childPath, f, visitor)
				if err != nil {
					return err
				}
			}
		}
		break

	case reflect.Map:
		keys := v.MapKeys()
		for _, key := range keys {
			mv := v.MapIndex(key)

			childPath := path + "[" + fmt.Sprintf("%s", mv.Interface()) + "]"
			err := visitor(childPath, nil, mv)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = walkRecursive(childPath, mv, visitor)
				if err != nil {
					return err
				}
			}
		}
		break

	case reflect.Array, reflect.Slice:
		len := v.Len()
		for i := 0; i < len; i++ {
			av := v.Index(i)

			childPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			err := visitor(childPath, nil, av)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = walkRecursive(childPath, av, visitor)
				if err != nil {
					return err
				}
			}
		}
		break

	case reflect.Ptr, reflect.Interface:
		err := visitor(path, nil, v)
		if err != nil && err != SkipReflection {
			return err
		}
		if err == nil && !v.IsNil() {
			e := v.Elem()
			err = walkRecursive(path, e, visitor)
			if err != nil {
				return err
			}
		}
		break

	default:
		err := visitor(path, nil, v)
		if err != nil && err != SkipReflection {
			return err
		}
	}

	return nil
}

// IsPrimitiveValue returns true if passed a value of primitive type: int, bool, etc
// Note that string (like []byte) is not treated as a primitive type
func IsPrimitiveValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true

	// The less-obvious cases!
	case reflect.String, reflect.Slice, reflect.Array:
		return false

	case reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func, reflect.Map, reflect.Struct, reflect.UnsafePointer:
		return false

	default:
		glog.Fatalf("Unhandled kind: %v", v.Kind())
		return false
	}
}
