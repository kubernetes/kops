/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package subnet

import (
	"encoding/binary"
	"fmt"
	"net"

	"k8s.io/klog/v2"
)

// CIDRMap is a helper structure to allocate unused CIDRs
type CIDRMap struct {
	used []net.IPNet
}

func (c *CIDRMap) MarkInUse(s string) error {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("error parsing network cidr %q: %v", s, err)
	}
	c.used = append(c.used, *cidr)
	return nil
}

func incrementIP(ip net.IP, mask net.IPMask) error {
	maskOnes, maskBits := mask.Size()

	if maskBits == 32 {
		ip4 := ip.To4()

		n := binary.BigEndian.Uint32(ip4)
		n += 1 << uint(32-maskOnes)
		binary.BigEndian.PutUint32(ip, n)
	} else {
		ipv6 := ip.To16()

		high := binary.BigEndian.Uint64(ipv6[0:8])
		low := binary.BigEndian.Uint64(ipv6[8:16])

		newHigh := high
		newLow := low

		shift := 128 - maskOnes
		if shift >= 64 {
			newHigh += 1 << uint64(shift-64)
		} else {
			newLow += 1 << uint64(shift)
			// Did we overflow?
			if newLow < low {
				newHigh++
			}
		}
		binary.BigEndian.PutUint64(ip[0:8], newHigh)
		binary.BigEndian.PutUint64(ip[8:16], newLow)
	}
	return nil
}

func duplicateIP(src net.IP) net.IP {
	ret := make(net.IP, len(src))
	copy(ret, src)
	return ret
}

func (c *CIDRMap) Allocate(from string, mask net.IPMask) (*net.IPNet, error) {
	_, cidr, err := net.ParseCIDR(from)
	if err != nil {
		return nil, fmt.Errorf("error parsing CIDR %q: %v", from, err)
	}

	var candidate net.IPNet
	candidate.Mask = mask
	candidate.IP = duplicateIP(cidr.IP)

	for {
		// Note we increment first, so we won't ever use the first range (e.g. 10.0.0.0/n)
		if err := incrementIP(candidate.IP, mask); err != nil {
			return nil, err
		}

		// Check we're still in the range we're drawing from
		if !cidrsOverlap(cidr, &candidate) {
			klog.Infof("candidate CIDR %v is not in CIDR %v", candidate, cidr)
			break
		}

		if !c.isInUse(&candidate) {
			if err := c.MarkInUse(candidate.String()); err != nil {
				return nil, err
			}
			return &candidate, nil
		}
	}

	return nil, fmt.Errorf("cannot allocate CIDR of size %v", mask)
}

func (c *CIDRMap) isInUse(n *net.IPNet) bool {
	for i := range c.used {
		if cidrsOverlap(&c.used[i], n) {
			return true
		}
	}
	return false
}

// cidrsOverlap returns true if and only if the two CIDRs are non-disjoint
func cidrsOverlap(l, r *net.IPNet) bool {
	return l.Contains(r.IP) || r.Contains(l.IP)
}
