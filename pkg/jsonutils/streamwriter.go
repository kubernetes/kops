/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jsonutils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// JSONStreamWriter writes tokens as parsed by a json.Decoder back to a string
type JSONStreamWriter struct {
	// out is the output destination
	out io.Writer

	// indent is the current indent level
	indent string

	// state stores a stack of the json state, comprised of [ { and F characters.  F=field
	state string

	// deferred is used to buffer the output temporarily, used to prevent a trailing comma in an object
	deferred string

	// path is the current stack of fields, used to support the Path() function
	path []string
}

// NewJSONStreamWriter is the constructor for a JSONStreamWriter
func NewJSONStreamWriter(out io.Writer) *JSONStreamWriter {
	return &JSONStreamWriter{
		out: out,
	}
}

// Path returns the path to the current position in the JSON tree
func (j *JSONStreamWriter) Path() string {
	return strings.Join(j.path, ".")
}

// WriteToken writes the next token to the output
func (j *JSONStreamWriter) WriteToken(token json.Token) error {
	state := byte(0)
	if j.state != "" {
		state = j.state[len(j.state)-1]
	}

	var v string
	switch tt := token.(type) {
	//	Delim, for the four JSON delimiters [ ] { }
	case json.Delim:
		v = tt.String()
		indent := j.indent
		switch tt {
		case json.Delim('{'):
			j.indent += "  "
			j.state += "{"
		case json.Delim('['):
			j.indent += "  "
			j.state += "["
		case json.Delim(']'), json.Delim('}'):
			j.indent = j.indent[:len(j.indent)-2]
			indent = j.indent
			j.state = j.state[:len(j.state)-1]
			if j.state != "" && j.state[len(j.state)-1] == 'F' {
				j.state = j.state[:len(j.state)-1]
				j.path = j.path[:len(j.path)-1]
			}
			// Don't put a comma on the last field in a block
			if j.deferred == ",\n" {
				j.deferred = "\n"
			}
		default:
			return fmt.Errorf("unknown delim: %v", tt)
		}

		switch state {
		case 0:
			if err := j.writeRaw(indent + v); err != nil {
				return err
			}
		case '{':
			if err := j.writeRaw(indent + v); err != nil {
				return err
			}
		case '[':
			if err := j.writeRaw(indent + v); err != nil {
				return err
			}
		case 'F':
			if err := j.writeRaw(v); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unhandled state for json delim serialization: %v %q", state, j.state)
		}

		switch tt {
		case json.Delim('{'):
			j.deferred = "\n"
		case json.Delim('['):
			j.deferred = "\n"
		case json.Delim(']'), json.Delim('}'):
			j.deferred = ",\n"
		default:
			return fmt.Errorf("unknown delim: %v", tt)
		}

		return nil

		//		bool, for JSON booleans
	case bool:
		v = fmt.Sprintf("%v", tt)

		//	string, for JSON string literals
	case string:
		v = "\"" + tt + "\""

		//	float64, for JSON numbers
	case float64:
		v = fmt.Sprintf("%g", tt)

		//	Number, for JSON numbers
	case json.Number:
		v = tt.String()

		//	nil, for JSON null
	case nil:
		v = "null"

	default:
		return fmt.Errorf("unhandled token type %T", tt)
	}

	switch state {
	case '{':
		j.state += "F"
		j.path = append(j.path, fmt.Sprintf("%s", token))
		return j.writeRaw(j.indent + v + ": ")
	case '[':
		if err := j.writeRaw(j.indent + v); err != nil {
			return err
		}
		j.deferred = ",\n"
		return nil
	case 'F':
		j.state = j.state[:len(j.state)-1]
		j.path = j.path[:len(j.path)-1]
		if err := j.writeRaw(v); err != nil {
			return err
		}
		j.deferred = ",\n"
		return nil
	}

	return fmt.Errorf("unhandled state for json value (%T %q) serialization: %v %q", token, v, state, j.state)
}

func (j *JSONStreamWriter) writeRaw(s string) error {
	if j.deferred != "" {
		if _, err := j.out.Write([]byte(j.deferred)); err != nil {
			return err
		}
		j.deferred = ""
	}
	_, err := j.out.Write([]byte(s))
	return err
}
