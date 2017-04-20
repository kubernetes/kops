package proxy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupObject(t *testing.T) {
	tests := []struct {
		root   jsonObject
		key    string
		result jsonObject
		err    error
	}{
		{
			jsonObject{},
			"a",
			jsonObject{},
			nil,
		},
		{
			jsonObject{"a": map[string]interface{}{"b": int(1)}},
			"a",
			jsonObject{"b": int(1)},
			nil,
		},
		{
			jsonObject{"nonObject": int(1)},
			"nonObject",
			nil,
			&UnmarshalWrongTypeError{Field: "nonObject", Expected: "object", Got: 1},
		},
	}
	for _, test := range tests {
		gotResult, gotErr := test.root.Object(test.key)
		msg := fmt.Sprintf("%q.Object(%q) => %q, %q", test.root, test.key, gotResult, gotErr)
		assert.Equal(t, test.result, gotResult, msg)
		assert.Equal(t, test.err, gotErr, msg)
	}
}

func TestLookupString(t *testing.T) {
	tests := []struct {
		root   jsonObject
		key    string
		result string
		err    error
	}{
		{
			jsonObject{},
			"a",
			"",
			nil,
		},
		{
			jsonObject{"nonString": int(1)},
			"nonString",
			"",
			&UnmarshalWrongTypeError{Field: "nonString", Expected: "string", Got: 1},
		},
	}
	for _, test := range tests {
		gotResult, gotErr := test.root.String(test.key)
		msg := fmt.Sprintf("%q.String(%q) => %q, %q", test.root, test.key, gotResult, gotErr)
		assert.Equal(t, test.result, gotResult, msg)
		assert.Equal(t, test.err, gotErr, msg)
	}
}

func TestLookupStringArray(t *testing.T) {
	tests := []struct {
		root   jsonObject
		key    string
		result []string
		err    error
	}{
		{
			jsonObject{},
			"a",
			nil,
			nil,
		},
		{
			jsonObject{"a": []string{"foo"}},
			"a",
			[]string{"foo"},
			nil,
		},
		{
			jsonObject{"a": []string{}},
			"a",
			[]string{},
			nil,
		},
		{
			jsonObject{"a": "foo"},
			"a",
			[]string{"foo"},
			nil,
		},
		{
			jsonObject{"int": 5},
			"int",
			nil,
			&UnmarshalWrongTypeError{Field: "int", Expected: "string or array of strings", Got: 5},
		},
	}
	for _, test := range tests {
		gotResult, gotErr := test.root.StringArray(test.key)
		msg := fmt.Sprintf("%q.String(%q) => %q, %q", test.root, test.key, gotResult, gotErr)
		assert.Equal(t, test.result, gotResult, msg)
		assert.Equal(t, test.err, gotErr, msg)
	}
}
