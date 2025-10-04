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

package aws

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// IPPermissionSet maps IP strings of strings to EC2 IpPermissions
type IPPermissionSet map[string]ec2types.IpPermission

// IPPermissionPredicate is an predicate to test whether IPPermission matches some condition.
type IPPermissionPredicate interface {
	// Test checks whether specified IPPermission matches condition.
	Test(perm ec2types.IpPermission) bool
}

// NewIPPermissionSet creates a new IPPermissionSet
func NewIPPermissionSet(items ...ec2types.IpPermission) IPPermissionSet {
	s := make(IPPermissionSet)
	s.Insert(items...)
	return s
}

// Ungroup splits permissions out into individual permissions
// EC2 will combine permissions with the same port but different SourceRanges together, for example
// We ungroup them so we can process them
func (s IPPermissionSet) Ungroup() IPPermissionSet {
	l := []ec2types.IpPermission{}
	for _, p := range s.List() {
		if len(p.IpRanges) <= 1 {
			l = append(l, p)
			continue
		}
		for _, ipRange := range p.IpRanges {
			c := ec2types.IpPermission{}
			c = p
			c.IpRanges = []ec2types.IpRange{ipRange}
			l = append(l, c)
		}
	}

	l2 := []ec2types.IpPermission{}
	for _, p := range l {
		if len(p.UserIdGroupPairs) <= 1 {
			l2 = append(l2, p)
			continue
		}
		for _, u := range p.UserIdGroupPairs {
			c := ec2types.IpPermission{}
			c = p
			c.UserIdGroupPairs = []ec2types.UserIdGroupPair{u}
			l2 = append(l2, c)
		}
	}

	l3 := []ec2types.IpPermission{}
	for _, p := range l2 {
		if len(p.PrefixListIds) <= 1 {
			l3 = append(l3, p)
			continue
		}
		for _, v := range p.PrefixListIds {
			c := ec2types.IpPermission{}
			c = p
			c.PrefixListIds = []ec2types.PrefixListId{v}
			l3 = append(l3, c)
		}
	}

	return NewIPPermissionSet(l3...)
}

// Insert adds items to the set.
func (s IPPermissionSet) Insert(items ...ec2types.IpPermission) {
	for _, p := range items {
		k := keyForIPPermission(p)
		s[k] = p
	}
}

// Delete delete permission from the set.
func (s IPPermissionSet) Delete(items ...ec2types.IpPermission) {
	for _, p := range items {
		k := keyForIPPermission(p)
		delete(s, k)
	}
}

// DeleteIf delete permission from the set if permission matches predicate.
func (s IPPermissionSet) DeleteIf(predicate IPPermissionPredicate) {
	for k, p := range s {
		if predicate.Test(p) {
			delete(s, k)
		}
	}
}

// List returns the contents as a slice.  Order is not defined.
func (s IPPermissionSet) List() []ec2types.IpPermission {
	res := make([]ec2types.IpPermission, 0, len(s))
	for _, v := range s {
		res = append(res, v)
	}
	return res
}

// IsSuperset returns true if and only if s is a superset of s2.
func (s IPPermissionSet) IsSuperset(s2 IPPermissionSet) bool {
	for k := range s2 {
		_, found := s[k]
		if !found {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s is equal (as a set) to s2.
// Two sets are equal if their membership is identical.
// (In practice, this means same elements, order doesn't matter)
func (s IPPermissionSet) Equal(s2 IPPermissionSet) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

// Difference returns a set of objects that are not in s2.
// For example:
// s1 = {a1, a2, a3}
// s2 = {a1, a2, a4, a5}
// s1.Difference(s2) = {a3}
// s2.Difference(s1) = {a4, a5}
func (s IPPermissionSet) Difference(s2 IPPermissionSet) IPPermissionSet {
	result := NewIPPermissionSet()
	for _, desired := range s.List() {
		found := false
		for _, existing := range s2.List() {
			if ipPermissionExists(&desired, &existing, false) {
				found = true
				break
			}
		}
		if !found {
			result.Insert(desired)
		}
	}
	return result
}

// Len returns the size of the set.
func (s IPPermissionSet) Len() int {
	return len(s)
}

func keyForIPPermission(p ec2types.IpPermission) string {
	v, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Sprintf("error building JSON representation of ec2.IpPermission: %v", err))
	}
	return string(v)
}

var _ IPPermissionPredicate = IPPermissionMatchDesc{}

// IPPermissionMatchDesc checks whether specific IPPermission contains description.
type IPPermissionMatchDesc struct {
	Description string
}

// Test whether specific IPPermission contains description.
func (p IPPermissionMatchDesc) Test(perm ec2types.IpPermission) bool {
	for _, v4Range := range perm.IpRanges {
		if aws.ToString(v4Range.Description) == p.Description {
			return true
		}
	}
	for _, v6Range := range perm.Ipv6Ranges {
		if aws.ToString(v6Range.Description) == p.Description {
			return true
		}
	}
	for _, prefixListID := range perm.PrefixListIds {
		if aws.ToString(prefixListID.Description) == p.Description {
			return true
		}
	}
	for _, group := range perm.UserIdGroupPairs {
		if aws.ToString(group.Description) == p.Description {
			return true
		}
	}
	return false
}

var _ IPPermissionPredicate = IPPermissionNotMatch{}

// IPPermissionNotMatch is the *not* operator for Predicate
type IPPermissionNotMatch struct {
	Predicate IPPermissionPredicate
}

// Test whether specific IPPermission not match the embed predicate.
func (p IPPermissionNotMatch) Test(perm ec2types.IpPermission) bool {
	return !p.Predicate.Test(perm)
}
