package tpm2

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

// HandleName returns the TPM Name of a PCR, session, or permanent value
// (e.g., hierarchy) handle.
func HandleName(h TPMHandle) TPM2BName {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(h))
	return TPM2BName{
		Buffer: result,
	}
}

// objectOrNVName calculates the Name of an NV index or object.
// pub is a pointer to either a TPMTPublic or TPMSNVPublic.
func objectOrNVName(alg TPMAlgID, pub interface{}) (*TPM2BName, error) {
	h, err := alg.Hash()
	if err != nil {
		return nil, err
	}

	// Create a byte slice with the correct reserved size and marshal the
	// NameAlg to it.
	result := make([]byte, 2, 2+h.Size())
	binary.BigEndian.PutUint16(result, uint16(alg))

	// Calculate the hash of the entire Public contents and append it to the
	// result.
	ha := h.New()
	var buf bytes.Buffer
	if err := marshal(&buf, reflect.ValueOf(pub)); err != nil {
		return nil, err
	}
	ha.Write(buf.Bytes())
	result = ha.Sum(result)

	return &TPM2BName{
		Buffer: result,
	}, nil
}

// ObjectName returns the TPM Name of an object.
func ObjectName(p *TPMTPublic) (*TPM2BName, error) {
	return objectOrNVName(p.NameAlg, p)
}

// NVName returns the TPM Name of an NV index.
func NVName(p *TPMSNVPublic) (*TPM2BName, error) {
	return objectOrNVName(p.NameAlg, p)
}

// PrimaryHandleName returns the TPM Name of a primary handle.
func PrimaryHandleName(h TPMHandle) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(h))
	return result
}
