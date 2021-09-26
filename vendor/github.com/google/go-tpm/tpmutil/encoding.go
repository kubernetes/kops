// Copyright (c) 2018, Google LLC All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tpmutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	selfMarshalerType = reflect.TypeOf((*SelfMarshaler)(nil)).Elem()
	handlesAreaType   = reflect.TypeOf((*[]Handle)(nil))
)

// packWithHeader takes a header and a sequence of elements that are either of
// fixed length or slices of fixed-length types and packs them into a single
// byte array using binary.Write. It updates the CommandHeader to have the right
// length.
func packWithHeader(ch commandHeader, cmd ...interface{}) ([]byte, error) {
	hdrSize := binary.Size(ch)
	body, err := Pack(cmd...)
	if err != nil {
		return nil, fmt.Errorf("couldn't pack message body: %v", err)
	}
	bodySize := len(body)
	ch.Size = uint32(hdrSize + bodySize)
	header, err := Pack(ch)
	if err != nil {
		return nil, fmt.Errorf("couldn't pack message header: %v", err)
	}
	return append(header, body...), nil
}

// Pack encodes a set of elements into a single byte array, using
// encoding/binary. This means that all the elements must be encodeable
// according to the rules of encoding/binary.
//
// It has one difference from encoding/binary: it encodes byte slices with a
// prepended length, to match how the TPM encodes variable-length arrays. If
// you wish to add a byte slice without length prefix, use RawBytes.
func Pack(elts ...interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := packType(buf, elts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// tryMarshal attempts to use a TPMMarshal() method defined on the type
// to pack v into buf. True is returned if the method exists and the
// marshal was attempted.
func tryMarshal(buf io.Writer, v reflect.Value) (bool, error) {
	t := v.Type()
	if t.Implements(selfMarshalerType) {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return true, fmt.Errorf("cannot call TPMMarshal on a nil pointer of type %T", v)
		}
		return true, v.Interface().(SelfMarshaler).TPMMarshal(buf)
	}

	// We might have a non-pointer struct field, but we dont have a
	// pointer with which to implement the interface.
	// If the pointer of the type implements the interface, we should be
	// able to construct a value to call TPMMarshal() with.
	// TODO(awly): Try and avoid blowing away private data by using Addr() instead of Set()
	if reflect.PtrTo(t).Implements(selfMarshalerType) {
		tmp := reflect.New(t)
		tmp.Elem().Set(v)
		return true, tmp.Interface().(SelfMarshaler).TPMMarshal(buf)
	}

	return false, nil
}

func packValue(buf io.Writer, v reflect.Value) error {
	if v.Type() == handlesAreaType {
		v = v.Convert(reflect.TypeOf((*handleList)(nil)))
	}
	if canMarshal, err := tryMarshal(buf, v); canMarshal {
		return err
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return fmt.Errorf("cannot pack nil %s", v.Type().String())
		}
		return packValue(buf, v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if err := packValue(buf, f); err != nil {
				return err
			}
		}
	default:
		return binary.Write(buf, binary.BigEndian, v.Interface())
	}
	return nil
}

func packType(buf io.Writer, elts ...interface{}) error {
	for _, e := range elts {
		if err := packValue(buf, reflect.ValueOf(e)); err != nil {
			return err
		}
	}

	return nil
}

// tryUnmarshal attempts to use TPMUnmarshal() to perform the
// unpack, if the given value implements SelfMarshaler.
// True is returned if v implements SelfMarshaler & TPMUnmarshal
// was called, along with an error returned from TPMUnmarshal.
func tryUnmarshal(buf io.Reader, v reflect.Value) (bool, error) {
	t := v.Type()
	if t.Implements(selfMarshalerType) {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return true, fmt.Errorf("cannot call TPMUnmarshal on a nil pointer")
		}
		return true, v.Interface().(SelfMarshaler).TPMUnmarshal(buf)
	}

	// We might have a non-pointer struct field, which is addressable,
	// If the pointer of the type implements the interface, and the
	// value is addressable, we should be able to call TPMUnmarshal().
	if v.CanAddr() && reflect.PtrTo(t).Implements(selfMarshalerType) {
		return true, v.Addr().Interface().(SelfMarshaler).TPMUnmarshal(buf)
	}

	return false, nil
}

// Unpack is a convenience wrapper around UnpackBuf. Unpack returns the number
// of bytes read from b to fill elts and error, if any.
func Unpack(b []byte, elts ...interface{}) (int, error) {
	buf := bytes.NewBuffer(b)
	err := UnpackBuf(buf, elts...)
	read := len(b) - buf.Len()
	return read, err
}

func unpackValue(buf io.Reader, v reflect.Value) error {
	if v.Type() == handlesAreaType {
		v = v.Convert(reflect.TypeOf((*handleList)(nil)))
	}
	if didUnmarshal, err := tryUnmarshal(buf, v); didUnmarshal {
		return err
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return fmt.Errorf("cannot unpack nil %s", v.Type().String())
		}
		return unpackValue(buf, v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if err := unpackValue(buf, f); err != nil {
				return err
			}
		}
		return nil
	default:
		// binary.Read can only set pointer values, so we need to take the address.
		if !v.CanAddr() {
			return fmt.Errorf("cannot unpack unaddressable leaf type %q", v.Type().String())
		}
		return binary.Read(buf, binary.BigEndian, v.Addr().Interface())
	}
}

// UnpackBuf recursively unpacks types from a reader just as encoding/binary
// does under binary.BigEndian, but with one difference: it unpacks a byte
// slice by first reading an integer with lengthPrefixSize bytes, then reading
// that many bytes. It assumes that incoming values are pointers to values so
// that, e.g., underlying slices can be resized as needed.
func UnpackBuf(buf io.Reader, elts ...interface{}) error {
	for _, e := range elts {
		v := reflect.ValueOf(e)
		if v.Kind() != reflect.Ptr {
			return fmt.Errorf("non-pointer value %q passed to UnpackBuf", v.Type().String())
		}
		if v.IsNil() {
			return errors.New("nil pointer passed to UnpackBuf")
		}

		if err := unpackValue(buf, v); err != nil {
			return err
		}
	}
	return nil
}
