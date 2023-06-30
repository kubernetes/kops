package tpm2

import (
	"bytes"
	"fmt"
	"reflect"
)

// Marshallable represents any TPM type that can be marshalled.
type Marshallable interface {
	// marshal will serialize the given value, appending onto the given buffer.
	// Returns an error if the value is not marshallable.
	marshal(buf *bytes.Buffer)
}

// marshallableWithHint represents any TPM type that can be marshalled,
// but that requires a selector ("hint") value when marshalling. Most TPMU_ are
// an example of this.
type marshallableWithHint interface {
	// get will return the corresponding union member by copy. If the union is
	// uninitialized, it will initialize a new zero-valued one.
	get(hint int64) (reflect.Value, error)
}

// Unmarshallable represents any TPM type that can be marshalled or unmarshalled.
type Unmarshallable interface {
	Marshallable
	// marshal will deserialize the given value from the given buffer.
	// Returns an error if there was an unmarshalling error or if there was not
	// enough data in the buffer.
	unmarshal(buf *bytes.Buffer) error
}

// unmarshallableWithHint represents any TPM type that can be marshalled or unmarshalled,
// but that requires a selector ("hint") value when unmarshalling. Most TPMU_ are
// an example of this.
type unmarshallableWithHint interface {
	marshallableWithHint
	// create will instantiate and return the corresponding union member.
	create(hint int64) (reflect.Value, error)
}

// Marshal will serialize the given values, returning them as a byte slice.
func Marshal(v Marshallable) []byte {
	var buf bytes.Buffer
	if err := marshal(&buf, reflect.ValueOf(v)); err != nil {
		panic(fmt.Sprintf("unexpected error marshalling %v: %v", reflect.TypeOf(v).Name(), err))
	}
	return buf.Bytes()
}

// Unmarshal unmarshals the given type from the byte array.
// Returns an error if the buffer does not contain enough data to satisfy the
// types, or if the types are not unmarshallable.
func Unmarshal[T Marshallable, P interface {
	*T
	Unmarshallable
}](data []byte) (*T, error) {
	buf := bytes.NewBuffer(data)
	var t T
	value := reflect.New(reflect.TypeOf(t))
	if err := unmarshal(buf, value.Elem()); err != nil {
		return nil, err
	}
	return value.Interface().(*T), nil
}

// marshallableByReflection is a placeholder interface, to hint to the unmarshalling
// library that it is supposed to use reflection.
type marshallableByReflection interface {
	reflectionSafe()
}

// marshalByReflection is embedded into any type that can be marshalled by reflection,
// needing no custom logic.
type marshalByReflection struct{}

func (marshalByReflection) reflectionSafe() {}

// These placeholders are required because a type constraint cannot union another interface
// that contains methods.
// Otherwise, marshalByReflection would not implement Unmarshallable, and the Marshal/Unmarshal
// functions would accept interface{ Marshallable | marshallableByReflection } instead.

// Placeholder: because this type implements the defaultMarshallable interface,
// the reflection library knows not to call this.
func (marshalByReflection) marshal(_ *bytes.Buffer) {
	panic("not implemented")
}

// Placeholder: because this type implements the defaultMarshallable interface,
// the reflection library knows not to call this.
func (*marshalByReflection) unmarshal(_ *bytes.Buffer) error {
	panic("not implemented")
}

// boxed is a helper type for corner cases such as unions, where all members must be structs.
type boxed[T any] struct {
	Contents *T
}

// box will put a value into a box.
func box[T any](contents *T) boxed[T] {
	return boxed[T]{
		Contents: contents,
	}
}

// unbox will take a value out of a box.
func (b *boxed[T]) unbox() *T {
	return b.Contents
}

// marshal implements the Marshallable interface.
func (b *boxed[T]) marshal(buf *bytes.Buffer) {
	if b.Contents == nil {
		var contents T
		marshal(buf, reflect.ValueOf(&contents))
	} else {
		marshal(buf, reflect.ValueOf(b.Contents))
	}
}

// unmarshal implements the Unmarshallable interface.
func (b *boxed[T]) unmarshal(buf *bytes.Buffer) error {
	b.Contents = new(T)
	return unmarshal(buf, reflect.ValueOf(b.Contents))
}
