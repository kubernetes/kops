package tpm2

// pcrSelectionFormatter is a Platform TPM Profile-specific interface for
// formatting TPM PCR selections.
// This interface isn't (yet) part of the go-tpm public interface. After we
// add a second implementation, we should consider making it public.
type pcrSelectionFormatter interface {
	// PCRs returns the TPM PCR selection bitmask associated with the given PCR indices.
	PCRs(pcrs ...uint) []byte
}

// PCClientCompatible is a pcrSelectionFormatter that formats PCR selections
// suitable for use in PC Client PTP-compatible TPMs (the vast majority):
// https://trustedcomputinggroup.org/resource/pc-client-platform-tpm-profile-ptp-specification/
// PC Client mandates at least 24 PCRs but does not provide an upper limit.
var PCClientCompatible pcrSelectionFormatter = pcClient{}

type pcClient struct{}

// The TPM requires all PCR selections to be at least big enough to select all
// the PCRs in the minimum PCR allocation.
const pcClientMinimumPCRCount = 24

func (pcClient) PCRs(pcrs ...uint) []byte {
	// Find the biggest PCR we selected.
	maxPCR := uint(0)
	for _, pcr := range pcrs {
		if pcr > maxPCR {
			maxPCR = pcr
		}
	}
	selectionSize := maxPCR/8 + 1

	// Enforce the minimum PCR selection size.
	if selectionSize < (pcClientMinimumPCRCount / 8) {
		selectionSize = (pcClientMinimumPCRCount / 8)
	}

	// Allocate a byte array to store the bitfield, that has at least
	// enough bits to store our selections.
	selection := make([]byte, selectionSize)
	for _, pcr := range pcrs {
		// The PCR selection mask is byte-wise little-endian:
		//   select[0] contains bits representing the selection of PCRs 0 through 7
		//   select[1] contains PCRs 8 through 15, and so on.
		byteIdx := pcr / 8
		// Within the byte, the PCR selection is bit-wise big-endian:
		//   bit 0 of select[0] contains the selection of PCR 0
		//   bit 1 of select[0] contains the selection of PCR 1, and so on.
		bitIdx := pcr % 8

		selection[byteIdx] |= (1 << bitIdx)
	}
	return selection
}
