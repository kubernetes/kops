package tpm2

import (
	"fmt"
)

// Bitfield represents a TPM bitfield (i.e., TPMA_*) type.
type Bitfield interface {
	// Length returns the length of the bitfield.
	Length() int
}

// BitGetter represents a TPM bitfield (i.e., TPMA_*) type that can be read.
type BitGetter interface {
	Bitfield
	// GetReservedBit returns the value of the given reserved bit.
	// If the bit is not reserved, returns false.
	GetReservedBit(pos int) bool
}

// BitSetter represents a TPM bitfield (i.e., TPMA_*) type that can be written.
type BitSetter interface {
	Bitfield
	// GetReservedBit sets the value of the given reserved bit.
	SetReservedBit(pos int, val bool)
}

func checkPos(pos int, len int) {
	if pos >= len || pos < 0 {
		panic(fmt.Errorf("bit %d out of range for %d-bit field", pos, len))
	}
}

// bitfield8 represents an 8-bit bitfield which may have reserved bits.
// 8-bit TPMA_* types embed this one, and the reserved bits are stored in it.
type bitfield8 uint8

// Length implements the Bitfield interface.
func (bitfield8) Length() int {
	return 8
}

// GetReservedBit implements the BitGetter interface.
func (r bitfield8) GetReservedBit(pos int) bool {
	checkPos(pos, 8)
	return r&(1<<pos) != 0
}

// SetReservedBit implements the BitSetter interface.
func (r *bitfield8) SetReservedBit(pos int, val bool) {
	checkPos(pos, 8)
	if val {
		*r |= 1 << pos
	} else {
		*r &= ^(1 << pos)
	}
}

// bitfield32 represents a 32-bit bitfield which may have reserved bits.
// 32-bit TPMA_* types embed this one, and the reserved bits are stored in it.
type bitfield32 uint32

// Length implements the Bitfield interface.
func (bitfield32) Length() int {
	return 32
}

// GetReservedBit implements the BitGetter interface.
func (r bitfield32) GetReservedBit(pos int) bool {
	checkPos(pos, 32)
	return r&(1<<pos) != 0
}

// SetReservedBit implements the BitSetter interface.
func (r *bitfield32) SetReservedBit(pos int, val bool) {
	checkPos(pos, 32)
	if val {
		*r |= 1 << pos
	} else {
		*r &= ^(1 << pos)
	}
}
