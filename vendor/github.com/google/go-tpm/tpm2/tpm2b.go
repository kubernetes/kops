package tpm2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// TPM2B is a helper type for all sized TPM structures. It can be instantiated with either a raw byte buffer or the actual struct.
type TPM2B[T Marshallable, P interface {
	*T
	Unmarshallable
}] struct {
	contents *T
	buffer   []byte
}

// New2B creates a new TPM2B containing the given contents.
func New2B[T Marshallable, P interface {
	*T
	Unmarshallable
}](t T) TPM2B[T, P] {
	return TPM2B[T, P]{contents: &t}
}

// BytesAs2B creates a new TPM2B containing the given byte array.
func BytesAs2B[T Marshallable, P interface {
	*T
	Unmarshallable
}](b []byte) TPM2B[T, P] {
	return TPM2B[T, P]{buffer: b}
}

// Contents returns the structured contents of the TPM2B.
// It can fail if the TPM2B was instantiated with an invalid byte buffer.
func (value *TPM2B[T, P]) Contents() (*T, error) {
	if value.contents != nil {
		return value.contents, nil
	}
	if value.buffer == nil {
		return nil, fmt.Errorf("TPMB had no contents or buffer")
	}
	contents, err := Unmarshal[T, P](value.buffer)
	if err != nil {
		return nil, err
	}
	// Cache the result
	value.contents = (*T)(contents)
	return value.contents, nil
}

// Bytes returns the inner contents of the TPM2B as a byte array, not including the length field.
func (value *TPM2B[T, P]) Bytes() []byte {
	if value.buffer != nil {
		return value.buffer
	}
	if value.contents == nil {
		return []byte{}
	}

	// Cache the result
	value.buffer = Marshal(*value.contents)
	return value.buffer
}

// marshal implements the tpm2.Marshallable interface.
func (value TPM2B[T, P]) marshal(buf *bytes.Buffer) {
	b := value.Bytes()
	binary.Write(buf, binary.BigEndian, uint16(len(b)))
	buf.Write(b)
}

// unmarshal implements the tpm2.Unmarshallable interface.
// Note: the structure contents are not validated during unmarshalling.
func (value *TPM2B[T, P]) unmarshal(buf *bytes.Buffer) error {
	var size uint16
	binary.Read(buf, binary.BigEndian, &size)
	value.contents = nil
	value.buffer = make([]byte, size)
	_, err := io.ReadAtLeast(buf, value.buffer, int(size))
	return err
}
