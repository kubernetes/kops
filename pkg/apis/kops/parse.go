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

package kops

import (
	"fmt"
	"strings"

	"k8s.io/kops/upup/pkg/fi/utils"
	"sigs.k8s.io/yaml"
)

// ParseInstanceGroupRole converts a string to an InstanceGroupRole.
//
// If lenient is set to true, the function will match pluralised words too.
// It will return the instance group role and true if a match was found.
func ParseInstanceGroupRole(input string, lenient bool) (InstanceGroupRole, bool) {
	findRole := strings.ToLower(input)
	if lenient {
		// Accept pluralized "bastions" for "bastion"
		findRole = strings.TrimSuffix(findRole, "s")
	}
	findRole = strings.Replace(findRole, "controlplane", "control-plane", 1)

	for _, role := range AllInstanceGroupRoles {
		s := role.ToLowerString()
		if lenient {
			s = strings.TrimSuffix(s, "s")
		}
		if s == findRole {
			return role, true
		}
	}

	if lenient && strings.ToLower(findRole) == "master" {
		return InstanceGroupRoleControlPlane, true
	}

	return "", false
}

// ParseRawYaml parses an object just using yaml, without the full api machinery
// Deprecated: prefer using the API machinery (package kopscodecs)
func ParseRawYaml(data []byte, dest interface{}) error {
	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err := yaml.Unmarshal([]byte(configString), dest, yaml.DisallowUnknownFields)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

// ToRawYaml marshals an object to yaml, without the full api machinery
// Deprecated: prefer using the API machinery (package kopscodecs)
func ToRawYaml(obj interface{}) ([]byte, error) {
	data, err := utils.YamlMarshal(obj)
	if err != nil {
		return nil, fmt.Errorf("error converting to yaml: %v", err)
	}

	return data, nil
}
