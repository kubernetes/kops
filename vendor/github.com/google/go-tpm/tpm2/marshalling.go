package tpm2

import (
	"bytes"
	"encoding/binary"
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

// MarshalCommand marshals a TPM command into its raw cpHash preimage format.
// The returned bytes can be directly hashed to compute cpHash.
//
// Example:
//
//	cmdData, _ := MarshalCommand(myCmd)
//	cpHash := sha256.Sum256(cmdData)
//
// Note: Encrypted command parameters (via sessions) are not currently supported.
// The marshaled parameters are in their unencrypted form.
func MarshalCommand[C Command[R, *R], R any](cmd C) ([]byte, error) {
	cc := cmd.Command()

	names, err := cmdNames(cmd)
	if err != nil {
		return nil, err
	}

	params, err := cmdParameters(cmd, nil)
	if err != nil {
		return nil, err
	}

	// Build raw cpHash preimage: CommandCode {∥ Name1 {∥ Name2 {∥ Name3 }}} {∥ Parameters }
	// See section 16.7 of TPM 2.0 specification, part 1.
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, cc); err != nil {
		return nil, fmt.Errorf("marshalling command code: %w", err)
	}

	for i, name := range names {
		if _, err := buf.Write(name.Buffer); err != nil {
			return nil, fmt.Errorf("marshalling name %d: %w", i, err)
		}
	}

	if _, err := buf.Write(params); err != nil {
		return nil, fmt.Errorf("marshalling parameters: %w", err)
	}

	return buf.Bytes(), nil
}

// UnmarshalCommand unmarshals a raw cpHash preimage back into a TPM command.
// The data should be the output from [MarshalCommand].
//
// Example:
//
//	cmdData, _ := MarshalCommand(myCmd)
//	cmd, _ := UnmarshalCommand[MyCommandType](cmdData)
//
// Notes:
//   - command produced from this function is not meant to be executed directly on a TPM,
//     instead it is expected to be used for purposes such as auditing or inspection.
//   - encrypted command parameters (via sessions) are not currently supported.
func UnmarshalCommand[C Command[R, *R], R any](data []byte) (C, error) {
	var cmd C

	if data == nil {
		return cmd, fmt.Errorf("data cannot be nil")
	}

	buf := bytes.NewBuffer(data)

	var cc TPMCC
	if err := binary.Read(buf, binary.BigEndian, &cc); err != nil {
		return cmd, fmt.Errorf("unmarshalling command code: %w", err)
	}

	if cc != cmd.Command() {
		return cmd, fmt.Errorf("command code mismatch: expected %v, got %v", cmd.Command(), cc)
	}

	expectedNames, err := cmdNames(cmd)
	if err != nil {
		return cmd, fmt.Errorf("getting expected names count: %w", err)
	}
	numNames := len(expectedNames)

	names := make([]TPM2BName, numNames)
	for i := range numNames {
		remaining := buf.Bytes()
		if len(remaining) == 0 {
			return cmd, fmt.Errorf("unexpected end of data while parsing name %d", i)
		}

		nameSize, err := parseNameSize(remaining)
		if err != nil {
			return cmd, fmt.Errorf("parsing name %d size: %w", i, err)
		}

		if len(remaining) < nameSize {
			return cmd, fmt.Errorf("insufficient data for name %d: need %d bytes, have %d", i, nameSize, len(remaining))
		}

		nameBytes := make([]byte, nameSize)
		if _, err := buf.Read(nameBytes); err != nil {
			return cmd, fmt.Errorf("reading name %d: %w", i, err)
		}

		names[i] = TPM2BName{Buffer: nameBytes}
	}

	// Populate the command's handle fields from the names
	if err := populateHandlesFromNames(&cmd, names); err != nil {
		return cmd, err
	}

	params := buf.Bytes()

	paramsBuf := bytes.NewBuffer(params)
	if err := unmarshalCmdParameters(paramsBuf, &cmd, nil); err != nil {
		return cmd, err
	}
	return cmd, nil
}

// parseNameSize determines the size of a TPM2BName by inspecting its first bytes.
// Returns the total size in bytes for the name.
//
// Case 1: Handle-based names (4 bytes)
//   - 0x0000... → PCR
//   - 0x02...   → HMAC Session
//   - 0x03...   → Policy Session
//   - 0x40...   → Permanent Values
//
// Case 2: Hash-based names (2 + hash_size bytes) - for all other entities
//   - Format: nameAlg (2 bytes) || H_nameAlg (hash digest)
//
// See section 14 of TPM 2.0 specification, part 1.
func parseNameSize(buf []byte) (int, error) {
	if len(buf) < 2 {
		return 0, fmt.Errorf("buffer too short to parse name")
	}

	firstByte := TPMHT(buf[0])
	firstTwoBytes := binary.BigEndian.Uint16(buf[0:2])

	// Case 1: Handle-based names (4 bytes)
	switch {
	case firstTwoBytes == 0x0000:
		// PCR handles (pattern: 0x0000XXXX)
		// Must check both bytes to distinguish from hash algorithms
		// that also start with 0x00 (e.g., TPMAlgSHA256 = 0x000B)
		return 4, nil
	case firstByte == TPMHTHMACSession: // 0x02
		return 4, nil
	case firstByte == TPMHTPolicySession: // 0x03
		return 4, nil
	case firstByte == TPMHTPermanent: // 0x40
		return 4, nil
	}

	// Case 2: Hash-based names (nameAlg || hash)
	// firstTwoBytes is the algorithm ID (0x0001 to 0x00B3)
	algID := TPMIAlgHash(firstTwoBytes)
	hashAlg, err := algID.Hash()
	if err != nil {
		return 0, fmt.Errorf("unsupported hash algorithm 0x%x in name: %w", firstTwoBytes, err)
	}

	// 2 bytes for algID + hash size
	return 2 + hashAlg.Size(), nil
}

// MarshalResponse marshals a TPM response into its raw rpHash preimage format.
// The returned bytes can be directly hashed to compute rpHash.
//
// Example:
//
//	rspData, _ := MarshalResponse(myCmd, myRsp)
//	rpHash := sha256.Sum256(rspData)
//
// Note: Encrypted response parameters (via sessions) are not currently supported.
func MarshalResponse[C Command[R, *R], R any](cmd C, rsp *R) ([]byte, error) {
	cc := cmd.Command()

	params, err := marshalRspParameters(rsp, nil)
	if err != nil {
		return nil, err
	}

	// Build raw rpHash preimage: responseCode || commandCode || parameters
	buf := new(bytes.Buffer)

	// Write responseCode (4 bytes, always 0 for successful responses)
	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, fmt.Errorf("marshalling response code: %w", err)
	}

	if err := binary.Write(buf, binary.BigEndian, cc); err != nil {
		return nil, fmt.Errorf("marshalling command code: %w", err)
	}

	if _, err := buf.Write(params); err != nil {
		return nil, fmt.Errorf("marshalling parameters: %w", err)
	}

	return buf.Bytes(), nil
}

// UnmarshalResponse unmarshals a raw rpHash preimage back into a TPM response.
// The data should be the output from [MarshalResponse].
//
// Example:
//
//	rspData, _ := MarshalResponse(commandCode, myRsp)
//	rsp, _ := UnmarshalResponse[MyResponseType](rspData)
//
// Notes:
//   - the result from this function is expected to be used for purposes such as auditing or inspection.
//   - encrypted response parameters (via sessions) are not currently supported.
func UnmarshalResponse[R any](data []byte) (*R, error) {
	var rsp R

	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	if len(data) < 8 {
		return nil, fmt.Errorf("data too short: need at least 8 bytes (responseCode + commandCode), got %d", len(data))
	}

	buf := bytes.NewBuffer(data)

	var responseCode uint32
	if err := binary.Read(buf, binary.BigEndian, &responseCode); err != nil {
		return nil, fmt.Errorf("unmarshalling response code: %w", err)
	}

	if responseCode != 0 {
		return nil, fmt.Errorf("invalid response code: expected 0, got 0x%x", responseCode)
	}

	var cc TPMCC
	if err := binary.Read(buf, binary.BigEndian, &cc); err != nil {
		return nil, fmt.Errorf("unmarshalling command code: %w", err)
	}

	params := buf.Bytes()

	if err := rspParameters(params, nil, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}
