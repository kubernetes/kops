package client

import (
	"fmt"
	"io"
	"math"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Reserved Handles from "TCG TPM v2.0 Provisioning Guidance" - v1r1 - Table 2
const (
	EKReservedHandle     = tpmutil.Handle(0x81010001)
	EKECCReservedHandle  = tpmutil.Handle(0x81010002)
	SRKReservedHandle    = tpmutil.Handle(0x81000001)
	SRKECCReservedHandle = tpmutil.Handle(0x81000002)
)

// From "TCG EK Credential Profile", v2.3r2 Section 2.2.1.4
const (
	// RSA 2048 EK Cert.
	EKCertNVIndexRSA uint32 = 0x01c00002
	// ECC P256 EK Cert.
	EKCertNVIndexECC uint32 = 0x01c0000a
)

// Picked available handles from TPM 2.0 Handles and Localities 2.3.1 - Table 11
// go-tpm-tools will use handles in the range from 0x81008F00 to 0x81008FFF
const (
	DefaultAKECCHandle = tpmutil.Handle(0x81008F00)
	DefaultAKRSAHandle = tpmutil.Handle(0x81008F01)
)

// GCE Attestation Key NV Indices
const (
	// RSA 2048 AK.
	GceAKCertNVIndexRSA     uint32 = 0x01c10000
	GceAKTemplateNVIndexRSA uint32 = 0x01c10001
	// ECC P256 AK.
	GceAKCertNVIndexECC     uint32 = 0x01c10002
	GceAKTemplateNVIndexECC uint32 = 0x01c10003
)

func isHierarchy(h tpmutil.Handle) bool {
	return h == tpm2.HandleOwner || h == tpm2.HandleEndorsement ||
		h == tpm2.HandlePlatform || h == tpm2.HandleNull
}

// Handles returns a slice of tpmutil.Handle objects of all handles within
// the TPM rw of type handleType.
func Handles(rw io.ReadWriter, handleType tpm2.HandleType) ([]tpmutil.Handle, error) {
	// Handle type is determined by the most-significant octet (MSO) of the property.
	property := uint32(handleType) << 24

	vals, moreData, err := tpm2.GetCapability(rw, tpm2.CapabilityHandles, math.MaxUint32, property)
	if err != nil {
		return nil, err
	}
	if moreData {
		return nil, fmt.Errorf("tpm2.GetCapability() should never return moreData==true for tpm2.CapabilityHandles")
	}
	handles := make([]tpmutil.Handle, len(vals))
	for i, v := range vals {
		handle, ok := v.(tpmutil.Handle)
		if !ok {
			return nil, fmt.Errorf("unable to assert type tpmutil.Handle of value %#v", v)
		}
		handles[i] = handle
	}
	return handles, nil
}
