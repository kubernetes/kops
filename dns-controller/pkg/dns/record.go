/*
Copyright 2019 The Kubernetes Authors.

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

package dns

type RecordType string

const (
	// RecordTypeAlias is unusual: the controller will try to resolve the target locally
	RecordTypeAlias = "_alias"

	RecordTypeA     = "A"
	RecordTypeCNAME = "CNAME"

	RoleTypeExternal = "external"
	RoleTypeInternal = "internal"
)

type Record struct {
	RecordType RecordType
	FQDN       string
	Value      string

	// If AliasTarget is set, this entry will not actually be set in DNS,
	// but will be used as an expansion for Records with type=RecordTypeAlias,
	// where the referring record has Value = our FQDN
	AliasTarget bool
}

// AliasForNodesInRole returns the alias for nodes in the given role
func AliasForNodesInRole(role, roleType string) string {
	return "node/role=" + role + "/" + roleType
}

func (r *Record) String() string {
	s := "Record:[Type=" + string(r.RecordType) + ",FQDN=" + r.FQDN + ",Value=" + r.Value

	if r.AliasTarget {
		s += ",AliasTarget"
	}

	s += "]"

	return s
}
