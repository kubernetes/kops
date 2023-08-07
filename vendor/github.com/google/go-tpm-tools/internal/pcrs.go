// Package internal contains private helper functions needed in client and server
package internal

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"io"

	pb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

const minPCRIndex = uint32(0)

func maxPCRIndex(p *pb.PCRs) uint32 {
	max := minPCRIndex
	for idx := range p.GetPcrs() {
		if idx > max {
			max = idx
		}
	}
	return max
}

// FormatPCRs writes a multiline representation of the PCR values to w.
func FormatPCRs(w io.Writer, p *pb.PCRs) error {
	if _, err := fmt.Fprintf(w, "%v:\n", p.Hash); err != nil {
		return err
	}
	for idx := minPCRIndex; idx <= maxPCRIndex(p); idx++ {
		if val, ok := p.GetPcrs()[idx]; ok {
			if _, err := fmt.Fprintf(w, "  %2d: 0x%X\n", idx, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckSubset verifies if the pcrs PCRs are a valid "subset" of the provided
// "superset" of PCRs. The PCR values must match (if present), and all PCRs must
// be present in the superset. This function will return an error containing the
// first missing or mismatched PCR number.
func CheckSubset(subset, superset *pb.PCRs) error {
	if subset.GetHash() != superset.GetHash() {
		return fmt.Errorf("PCR hash algo not matching: %v, %v", subset.GetHash(), superset.GetHash())
	}
	for pcrNum, pcrVal := range subset.GetPcrs() {
		if expectedVal, ok := superset.GetPcrs()[pcrNum]; ok {
			if !bytes.Equal(expectedVal, pcrVal) {
				return fmt.Errorf("PCR %d mismatch: expected %v, got %v",
					pcrNum, hex.EncodeToString(expectedVal), hex.EncodeToString(pcrVal))
			}
		} else {
			return fmt.Errorf("PCR %d mismatch: value missing from the superset PCRs", pcrNum)
		}
	}
	return nil
}

// PCRSelection returns the corresponding tpm2.PCRSelection for the PCR data.
func PCRSelection(p *pb.PCRs) tpm2.PCRSelection {
	sel := tpm2.PCRSelection{Hash: tpm2.Algorithm(p.GetHash())}

	for pcrNum := range p.GetPcrs() {
		sel.PCRs = append(sel.PCRs, int(pcrNum))
	}
	return sel
}

// SamePCRSelection checks if the Pcrs has the same PCRSelection as the
// provided given tpm2.PCRSelection (including the hash algorithm).
func SamePCRSelection(p *pb.PCRs, sel tpm2.PCRSelection) bool {
	if tpm2.Algorithm(p.GetHash()) != sel.Hash {
		return false
	}
	if len(p.GetPcrs()) != len(sel.PCRs) {
		return false
	}
	for _, pcr := range sel.PCRs {
		if _, ok := p.Pcrs[uint32(pcr)]; !ok {
			return false
		}
	}
	return true
}

// PCRSessionAuth calculates the authorization value for the given PCRs.
func PCRSessionAuth(p *pb.PCRs, hashAlg crypto.Hash) []byte {
	// Start with all zeros, we only use a single policy command on our session.
	oldDigest := make([]byte, hashAlg.Size())
	ccPolicyPCR, _ := tpmutil.Pack(tpm2.CmdPolicyPCR)

	// Extend the policy digest, see TPM2_PolicyPCR in Part 3 of the spec.
	hash := hashAlg.New()
	hash.Write(oldDigest)
	hash.Write(ccPolicyPCR)
	hash.Write(encodePCRSelection(PCRSelection(p)))
	hash.Write(PCRDigest(p, hashAlg))
	newDigest := hash.Sum(nil)
	return newDigest[:]
}

// PCRDigest computes the digest of the Pcrs. Note that the digest hash
// algorithm may differ from the PCRs' hash (which denotes the PCR bank).
func PCRDigest(p *pb.PCRs, hashAlg crypto.Hash) []byte {
	hash := hashAlg.New()
	for i := uint32(0); i < 24; i++ {
		if pcrValue, exists := p.GetPcrs()[i]; exists {
			hash.Write(pcrValue)
		}
	}
	return hash.Sum(nil)
}

// Encode a tpm2.PCRSelection as if it were a TPML_PCR_SELECTION
func encodePCRSelection(sel tpm2.PCRSelection) []byte {
	// Encode count, pcrSelections.hash and pcrSelections.sizeofSelect fields
	buf, _ := tpmutil.Pack(uint32(1), sel.Hash, byte(3))
	// Encode pcrSelect bitmask
	pcrBits := make([]byte, 3)
	for _, pcr := range sel.PCRs {
		byteNum := pcr / 8
		bytePos := 1 << uint(pcr%8)
		pcrBits[byteNum] |= byte(bytePos)
	}

	return append(buf, pcrBits...)
}
