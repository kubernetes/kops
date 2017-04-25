package address

import (
	"fmt"
	"net"

	"github.com/weaveworks/weave/common"
)

// Using 32-bit integer to represent IPv4 address
type Address uint32
type Offset uint32
type Count uint32

type Range struct {
	Start, End Address // [Start, End); Start <= End
}

func NewRange(start Address, size Offset) Range {
	return Range{Start: start, End: Add(start, size)}
}
func (r Range) Size() Count                { return Length(r.End, r.Start) }
func (r Range) String() string             { return fmt.Sprintf("%s-%s", r.Start, r.End-1) }
func (r Range) Overlaps(or Range) bool     { return !(r.Start >= or.End || r.End <= or.Start) }
func (r Range) Contains(addr Address) bool { return addr >= r.Start && addr < r.End }

func (r Range) AsCIDRString() string {
	prefixLen := 32
	for size := r.Size(); size > 1; size = size / 2 {
		if size%2 != 0 { // Size not a power of two; cannot be expressed as a CIDR.
			return r.String()
		}
		prefixLen--
	}
	return CIDR{Addr: r.Start, PrefixLen: prefixLen}.String()
}

// return the highest bit set in v
// algorithm from http://graphics.stanford.edu/~seander/bithacks.html#RoundUpPowerOf2
func msb(v Count) Count {
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	// Cast to 64-bit to avoid overflow when original v >= 0x80000000
	return Count((uint64(v) + 1) / 2)
}

func (r Range) BiggestCIDRRange() Range {
	sizeMsb := Offset(msb(r.Size()))
	maskedSize := Offset(r.Size()) & (sizeMsb - 1)
	maskedStart := Offset(r.Start) & (sizeMsb - 1)
	if maskedStart == 0 {
		return NewRange(r.Start, sizeMsb)
	}
	if sizeMsb-maskedStart <= maskedSize {
		return NewRange(Add(r.Start, sizeMsb-maskedStart), sizeMsb)
	}
	size := sizeMsb / 2
	if size >= maskedStart {
		return NewRange(Add(r.Start, size-maskedStart), size)
	}
	return NewRange(Add(r.Start, sizeMsb-maskedStart), size)
}

func MakeCIDR(subnet CIDR, addr Address) CIDR {
	return CIDR{Addr: addr, PrefixLen: subnet.PrefixLen}
}

type CIDR struct {
	Addr      Address
	PrefixLen int
}

// CIDRs returns a list of CIDR-aligned ranges which cover this range.
func (r Range) CIDRs() []CIDR {
	const (
		fullMask         = ^Address(0)
		cidrMaxPrefixLen = 32
	)
	var cidrs []CIDR

	for start, end := r.Start, r.End-1; end >= start; {
		mask, prefixLen := fullMask, cidrMaxPrefixLen
		// Find the smallest mask which would cover some part of [start;end].
		// Once we found such, apply it by OR'ing
		for mask > 0 {
			tmpMask := mask << 1
			// Check whether mask neither too short nor too long
			if (start&tmpMask) != start || (start|^tmpMask) > end {
				break
			}
			mask = tmpMask
			prefixLen--
		}
		cidrs = append(cidrs, CIDR{start, prefixLen})
		// Apply mask
		start |= ^mask
		// Check for overflow
		if start+1 < start {
			break
		}
		start++
	}

	return cidrs
}

func ParseIP(s string) (Address, error) {
	if ip := net.ParseIP(s); ip != nil {
		return FromIP4(ip), nil
	}
	return 0, &net.ParseError{Type: "IP Address", Text: s}
}

func ParseCIDR(s string) (CIDR, error) {
	if ip, ipnet, err := net.ParseCIDR(s); err != nil {
		return CIDR{}, err
	} else if ipnet.IP.To4() == nil {
		return CIDR{}, &net.ParseError{Type: "Non-IPv4 address not supported", Text: s}
	} else {
		prefixLen, _ := ipnet.Mask.Size()
		return CIDR{Addr: FromIP4(ip), PrefixLen: prefixLen}, nil
	}
}

func NewCIDRs(ranges []Range) (cidrs []CIDR) {
	for _, r := range ranges {
		cidrs = append(cidrs, r.CIDRs()...)
	}
	return cidrs
}

func (cidr CIDR) Start() Address {
	return cidr.Addr
}

// cidr = [Start; End)
func (cidr CIDR) End() Address {
	return cidr.Range().End
}

func (cidr CIDR) IsSubnet() bool {
	mask := cidr.Size() - 1
	return Offset(cidr.Addr)&mask == 0
}

func (cidr CIDR) Size() Offset { return 1 << uint(32-cidr.PrefixLen) }

func (cidr CIDR) Range() Range {
	return NewRange(cidr.Addr, cidr.Size())
}
func (cidr CIDR) HostRange() Range {
	// Respect RFC1122 exclusions of first and last addresses
	return NewRange(cidr.Addr+1, cidr.Size()-2)
}

func (cidr CIDR) String() string {
	return fmt.Sprintf("%s/%d", cidr.Addr.String(), cidr.PrefixLen)
}

// FromIP4 converts an ipv4 address to our integer address type
func FromIP4(ip4 net.IP) (r Address) {
	for _, b := range ip4.To4() {
		r <<= 8
		r |= Address(b)
	}
	return
}

// IP4 converts our integer address type to an ipv4 address
func (addr Address) IP4() (r net.IP) {
	r = make([]byte, net.IPv4len)
	for i := 3; i >= 0; i-- {
		r[i] = byte(addr)
		addr >>= 8
	}
	return
}

func (addr Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", addr.String())), nil
}

func (addr Address) String() string {
	return addr.IP4().String()
}

func Add(addr Address, i Offset) Address {
	return addr + Address(i)
}

func Subtract(a, b Address) Offset {
	common.Assert(a >= b)
	return Offset(a - b)
}

func Length(a, b Address) Count {
	common.Assert(a >= b)
	return Count(a - b)
}

func Min(a, b Count) Count {
	if a > b {
		return b
	}
	return a
}

func (addr Address) Reverse() Address {
	return ((addr >> 24) & 0xff) | // move byte 3 to byte 0
		((addr << 8) & 0xff0000) | // move byte 1 to byte 2
		((addr >> 8) & 0xff00) | // move byte 2 to byte 1
		((addr << 24) & 0xff000000) // byte 0 to byte 3
}
