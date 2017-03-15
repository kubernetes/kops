/*
Copyright 2016 The Kubernetes Authors.

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

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net"
	"net/url"
)

// isSubnet checks if child is a subnet of parent
func isSubnet(parent *net.IPNet, child *net.IPNet) bool {
	parentOnes, parentBits := parent.Mask.Size()
	childOnes, childBits := child.Mask.Size()
	if childBits != parentBits {
		return false
	}
	if parentOnes > childOnes {
		return false
	}
	childMasked := child.IP.Mask(parent.Mask)
	parentMasked := parent.IP.Mask(parent.Mask)
	return childMasked.Equal(parentMasked)
}

// subnetsOverlap checks if two subnets overlap
func subnetsOverlap(l *net.IPNet, r *net.IPNet) bool {
	return l.Contains(r.IP) || r.Contains(l.IP)
}

func isValidAPIServersURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Host == "" || u.Scheme == "" {
		return false
	}
	return true
}

func IsValidValue(fldPath *field.Path, v *string, validValues []string) field.ErrorList {
	allErrs := field.ErrorList{}
	if v != nil {
		found := false
		for _, validValue := range validValues {
			if *v == validValue {
				found = true
				break
			}
		}
		if !found {
			allErrs = append(allErrs, field.NotSupported(fldPath, *v, validValues))
		}
	}
	return allErrs
}
