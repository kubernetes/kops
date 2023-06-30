package client

import (
	"crypto"
	"fmt"
	"io"
	"math"

	pb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/legacy/tpm2"
)

// NumPCRs is set to the spec minimum of 24, as that's all go-tpm supports.
const NumPCRs = 24

// We hard-code SHA256 as the policy session hash algorithms. Note that this
// differs from the PCR hash algorithm (which selects the bank of PCRs to use)
// and the Public area Name algorithm. We also chose this for compatibility with
// github.com/google/go-tpm/legacy/tpm2, as it hardcodes the nameAlg as SHA256 in
// several places. Two constants are used to avoid repeated conversions.
const (
	SessionHashAlg    = crypto.SHA256
	SessionHashAlgTpm = tpm2.AlgSHA256
)

// CertifyHashAlgTpm is the hard-coded algorithm used in certify PCRs.
const CertifyHashAlgTpm = tpm2.AlgSHA256

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// allocatedPCRs returns a list of selections corresponding to the TPM's implemented PCRs.
func allocatedPCRs(rw io.ReadWriter) ([]tpm2.PCRSelection, error) {
	caps, moreData, err := tpm2.GetCapability(rw, tpm2.CapabilityPCRs, math.MaxUint32, 0)
	if err != nil {
		return nil, fmt.Errorf("listing implemented PCR banks: %w", err)
	}
	if moreData {
		return nil, fmt.Errorf("extra data from GetCapability")
	}
	var sels []tpm2.PCRSelection
	for _, cap := range caps {
		sel, ok := cap.(tpm2.PCRSelection)
		if !ok {
			return nil, fmt.Errorf("unexpected data from GetCapability")
		}
		// skip empty (unallocated) PCR selections
		if len(sel.PCRs) == 0 {
			continue
		}
		sels = append(sels, sel)
	}
	return sels, nil
}

// ReadPCRs fetches all the PCR values specified in sel, making multiple calls
// to the TPM if necessary.
func ReadPCRs(rw io.ReadWriter, sel tpm2.PCRSelection) (*pb.PCRs, error) {
	pl := pb.PCRs{
		Hash: pb.HashAlgo(sel.Hash),
		Pcrs: map[uint32][]byte{},
	}

	for i := 0; i < len(sel.PCRs); i += 8 {
		end := min(i+8, len(sel.PCRs))
		pcrSel := tpm2.PCRSelection{
			Hash: sel.Hash,
			PCRs: sel.PCRs[i:end],
		}

		pcrMap, err := tpm2.ReadPCRs(rw, pcrSel)
		if err != nil {
			return nil, err
		}

		for pcr, val := range pcrMap {
			pl.Pcrs[uint32(pcr)] = val
		}
	}

	return &pl, nil
}

// ReadAllPCRs fetches all the PCR values from all implemented PCR banks.
func ReadAllPCRs(rw io.ReadWriter) ([]*pb.PCRs, error) {
	sels, err := allocatedPCRs(rw)
	if err != nil {
		return nil, err
	}

	allPcrs := make([]*pb.PCRs, len(sels))
	for i, sel := range sels {
		allPcrs[i], err = ReadPCRs(rw, sel)
		if err != nil {
			return nil, fmt.Errorf("reading bank %x PCRs: %w", sel.Hash, err)
		}
	}
	return allPcrs, nil
}

// SealOpts specifies the PCR values that should be used for Seal().
type SealOpts struct {
	// Current seals data to the current specified PCR selection.
	Current tpm2.PCRSelection
	// Target predictively seals data to the given specified PCR values.
	Target *pb.PCRs
}

// UnsealOpts specifies the options that should be used for Unseal().
// Currently, it specifies the PCRs that need to pass certification in order to
// successfully unseal.
// CertifyHashAlgTpm is the hard-coded algorithm that must be used with
// UnsealOpts.
type UnsealOpts struct {
	// CertifyCurrent certifies that a selection of current PCRs have the same
	// value when sealing.
	CertifyCurrent tpm2.PCRSelection
	// CertifyExpected certifies that the TPM had a specific set of PCR values when sealing.
	CertifyExpected *pb.PCRs
}

// FullPcrSel will return a full PCR selection based on the total PCR number
// of the TPM with the given hash algo.
func FullPcrSel(hash tpm2.Algorithm) tpm2.PCRSelection {
	sel := tpm2.PCRSelection{Hash: hash}
	for i := 0; i < NumPCRs; i++ {
		sel.PCRs = append(sel.PCRs, int(i))
	}
	return sel
}

func mergePCRSelAndProto(rw io.ReadWriter, sel tpm2.PCRSelection, proto *pb.PCRs) (*pb.PCRs, error) {
	if proto == nil || len(proto.GetPcrs()) == 0 {
		return ReadPCRs(rw, sel)
	}
	if len(sel.PCRs) == 0 {
		return proto, nil
	}
	if sel.Hash != tpm2.Algorithm(proto.Hash) {
		return nil, fmt.Errorf("current hash (%v) differs from target hash (%v)",
			sel.Hash, tpm2.Algorithm(proto.Hash))
	}

	// At this point, both sel and proto are non-empty.
	// Verify no overlap in sel and proto PCR indexes.
	overlap := make([]int, 0)
	targetMap := proto.GetPcrs()
	for _, pcrVal := range sel.PCRs {
		if _, found := targetMap[uint32(pcrVal)]; found {
			overlap = append(overlap, pcrVal)
		}
	}
	if len(overlap) != 0 {
		return nil, fmt.Errorf("found PCR overlap: %v", overlap)
	}

	currentPcrs, err := ReadPCRs(rw, sel)
	if err != nil {
		return nil, err
	}

	for pcr, val := range proto.GetPcrs() {
		currentPcrs.Pcrs[pcr] = val
	}
	return currentPcrs, nil
}
