package proxy

import (
	"encoding/json"
	"fmt"
)

type UnmarshalWrongTypeError struct {
	Field, Expected string
	Got             interface{}
}

func (e *UnmarshalWrongTypeError) Error() string {
	return fmt.Sprintf("Wrong type for %s field, expected %s, but got %T", e.Field, e.Expected, e.Got)
}

type jsonObject map[string]interface{}

func (j jsonObject) Object(key string) (jsonObject, error) {
	iface, ok := j[key]
	if !ok || iface == nil {
		result := jsonObject{}
		j[key] = result
		return result, nil
	}

	result, ok := iface.(map[string]interface{})
	if !ok {
		return nil, &UnmarshalWrongTypeError{key, "object", iface}
	}

	return jsonObject(result), nil
}

func (j jsonObject) String(key string) (string, error) {
	iface, ok := j[key]
	if !ok || iface == nil {
		return "", nil
	}

	result, ok := iface.(string)
	if !ok {
		return "", &UnmarshalWrongTypeError{key, "string", iface}
	}

	return result, nil
}

func (j jsonObject) Int(key string) (int, error) {
	iface, ok := j[key]
	if !ok || iface == nil {
		return 0, nil
	}

	result, ok := iface.(json.Number)
	if !ok {
		return 0, &UnmarshalWrongTypeError{key, "json.Number", iface}
	}

	i64, err := result.Int64()
	if err != nil {
		return 0, err
	}

	return int(i64), nil
}

func (j jsonObject) StringArray(key string) ([]string, error) {
	iface, ok := j[key]
	if !ok || iface == nil {
		return nil, nil
	}

	switch o := iface.(type) {
	case string:
		return []string{o}, nil
	case []string:
		return o, nil
	case []interface{}:
		result := []string{}
		for _, s := range o {
			if s, ok := s.(string); ok {
				result = append(result, s)
			} else {
				return nil, &UnmarshalWrongTypeError{key, "string or array of strings", iface}
			}
		}
		return result, nil
	}

	return nil, &UnmarshalWrongTypeError{key, "string or array of strings", iface}
}
