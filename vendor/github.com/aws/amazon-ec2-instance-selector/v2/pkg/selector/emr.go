// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package selector

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/blang/semver/v4"
)

const (
	fallbackVersion = "5.20.0"
)

// EMR is a Service type for a custom service filter transform
type EMR struct{}

// Filters implements the Service interface contract for EMR
func (e EMR) Filters(version string) (Filters, error) {
	filters := Filters{}
	if version == "" {
		version = fallbackVersion
	}
	semanticVersion, err := semver.Make(version)
	if err != nil {
		return filters, err
	}
	if err := semanticVersion.Validate(); err != nil {
		return filters, fmt.Errorf("Invalid semantic version passed for EMR")
	}
	instanceTypes, err := e.getEMRInstanceTypes(semanticVersion)
	if err != nil {
		return filters, err
	}
	filters.InstanceTypes = &instanceTypes
	filters.RootDeviceType = aws.String("ebs")
	filters.VirtualizationType = aws.String("hvm")
	return filters, nil
}

// getEMRInstanceTypes returns a list of instance types that emr supports
func (e EMR) getEMRInstanceTypes(version semver.Version) ([]string, error) {
	instanceTypes := []string{}

	for _, instanceType := range e.getAllEMRInstanceTypes() {
		if semver.MustParseRange(">=5.25.0")(version) {
			instanceTypes = append(instanceTypes, instanceType)
		} else if semver.MustParseRange(">=5.20.0 <5.25.0")(version) {
			if e.isOnlyEMR_5_25_0_plus(instanceType) {
				continue
			}
			instanceTypes = append(instanceTypes, instanceType)
		} else if semver.MustParseRange(">=5.15.0 <5.20.0")(version) {
			if instanceType == "c1.medium" {
				continue
			}
			if e.isOnlyEMR_5_20_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_25_0_plus(instanceType) {
				continue
			}
			instanceTypes = append(instanceTypes, instanceType)
		} else if semver.MustParseRange(">=5.13.0 <5.15.0")(version) {
			if e.isOnlyEMR_5_20_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_25_0_plus(instanceType) {
				continue
			}
			instanceTypes = append(instanceTypes, instanceType)
		} else if semver.MustParseRange(">=5.9.0 <5.13.0")(version) {
			if e.isEMR_5_13_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_20_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_25_0_plus(instanceType) {
				continue
			}
			instanceTypes = append(instanceTypes, instanceType)
		} else {
			if e.isEMR_5_13_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_20_0_plus(instanceType) {
				continue
			}
			if e.isOnlyEMR_5_25_0_plus(instanceType) {
				continue
			}
			if strings.HasPrefix(instanceType, "i3") {
				continue
			}
			instanceTypes = append(instanceTypes, instanceType)
		}
	}
	return instanceTypes, nil
}

func (EMR) isEMR_5_13_0_plus(instanceType string) bool {
	prefixes := []string{
		"m5.",
		"m5d.",
		"c5.",
		"c5d.",
		"r5.",
		"r5d.",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(instanceType, prefix) {
			return true
		}
	}
	return false
}

func (EMR) isOnlyEMR_5_20_0_plus(instanceType string) bool {
	prefixes := []string{
		"m5a.",
		"c5n.",
		"r5a.",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(instanceType, prefix) {
			return true
		}
	}
	return false
}

func (EMR) isOnlyEMR_5_25_0_plus(instanceType string) bool {
	prefixes := []string{
		"i3en.",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(instanceType, prefix) {
			return true
		}
	}
	return false
}

func (EMR) getAllEMRInstanceTypes() []string {
	return []string{
		"c1.medium",
		"c1.xlarge",
		"c3.2xlarge",
		"c3.4xlarge",
		"c3.8xlarge",
		"c3.xlarge",
		"c4.2xlarge",
		"c4.4xlarge",
		"c4.8xlarge",
		"c4.large",
		"c4.xlarge",
		"c5.12xlarge",
		"c5.18xlarge",
		"c5.24xlarge",
		"c5.2xlarge",
		"c5.4xlarge",
		"c5.9xlarge",
		"c5.xlarge",
		"c5a.12xlarge",
		"c5a.16xlarge",
		"c5a.24xlarge",
		"c5a.2xlarge",
		"c5a.4xlarge",
		"c5a.8xlarge",
		"c5a.xlarge",
		"c5d.12xlarge",
		"c5d.18xlarge",
		"c5d.24xlarge",
		"c5d.2xlarge",
		"c5d.4xlarge",
		"c5d.9xlarge",
		"c5d.xlarge",
		"c5n.18xlarge",
		"c5n.2xlarge",
		"c5n.4xlarge",
		"c5n.9xlarge",
		"c5n.xlarge",
		"c6g.12xlarge",
		"c6g.16xlarge",
		"c6g.2xlarge",
		"c6g.4xlarge",
		"c6g.8xlarge",
		"c6g.xlarge",
		"cc2.8xlarge",
		"d2.2xlarge",
		"d2.4xlarge",
		"d2.8xlarge",
		"d2.xlarge",
		"g2.2xlarge",
		"g3.16xlarge",
		"g3.4xlarge",
		"g3.8xlarge",
		"g3s.xlarge",
		"g4dn.12xlarge",
		"g4dn.16xlarge",
		"g4dn.2xlarge",
		"g4dn.4xlarge",
		"g4dn.8xlarge",
		"g4dn.xlarge",
		"h1.16xlarge",
		"h1.2xlarge",
		"h1.4xlarge",
		"h1.8xlarge",
		"i2.2xlarge",
		"i2.4xlarge",
		"i2.8xlarge",
		"i2.xlarge",
		"i3.16xlarge",
		"i3.2xlarge",
		"i3.4xlarge",
		"i3.8xlarge",
		"i3.xlarge",
		"i3en.12xlarge",
		"i3en.24xlarge",
		"i3en.2xlarge",
		"i3en.3xlarge",
		"i3en.6xlarge",
		"i3en.xlarge",
		"m1.large",
		"m1.medium",
		"m1.small",
		"m1.xlarge",
		"m2.2xlarge",
		"m2.4xlarge",
		"m2.xlarge",
		"m3.2xlarge",
		"m3.xlarge",
		"m4.10xlarge",
		"m4.16xlarge",
		"m4.2xlarge",
		"m4.4xlarge",
		"m4.large",
		"m4.xlarge",
		"m5.12xlarge",
		"m5.16xlarge",
		"m5.24xlarge",
		"m5.2xlarge",
		"m5.4xlarge",
		"m5.8xlarge",
		"m5.metal",
		"m5.xlarge",
		"m5a.12xlarge",
		"m5a.16xlarge",
		"m5a.24xlarge",
		"m5a.2xlarge",
		"m5a.4xlarge",
		"m5a.8xlarge",
		"m5a.xlarge",
		"m5ad.12xlarge",
		"m5ad.16xlarge",
		"m5ad.24xlarge",
		"m5ad.2xlarge",
		"m5ad.4xlarge",
		"m5ad.8xlarge",
		"m5ad.xlarge",
		"m5d.12xlarge",
		"m5d.16xlarge",
		"m5d.24xlarge",
		"m5d.2xlarge",
		"m5d.4xlarge",
		"m5d.8xlarge",
		"m5d.metal",
		"m5d.xlarge",
		"m5dn.12xlarge",
		"m5dn.16xlarge",
		"m5dn.24xlarge",
		"m5dn.2xlarge",
		"m5dn.4xlarge",
		"m5dn.8xlarge",
		"m5dn.xlarge",
		"m5n.12xlarge",
		"m5n.16xlarge",
		"m5n.24xlarge",
		"m5n.2xlarge",
		"m5n.4xlarge",
		"m5n.8xlarge",
		"m5n.xlarge",
		"m6g.12xlarge",
		"m6g.16xlarge",
		"m6g.2xlarge",
		"m6g.4xlarge",
		"m6g.8xlarge",
		"m6g.xlarge",
		"mac1.metal",
		"p2.16xlarge",
		"p2.8xlarge",
		"p2.xlarge",
		"p3.16xlarge",
		"p3.2xlarge",
		"p3.8xlarge",
		"p3dn.24xlarge",
		"r3.2xlarge",
		"r3.4xlarge",
		"r3.8xlarge",
		"r3.xlarge",
		"r4.16xlarge",
		"r4.2xlarge",
		"r4.4xlarge",
		"r4.8xlarge",
		"r4.xlarge",
		"r5.12xlarge",
		"r5.16xlarge",
		"r5.24xlarge",
		"r5.2xlarge",
		"r5.4xlarge",
		"r5.8xlarge",
		"r5.metal",
		"r5.xlarge",
		"r5a.12xlarge",
		"r5a.16xlarge",
		"r5a.24xlarge",
		"r5a.2xlarge",
		"r5a.4xlarge",
		"r5a.8xlarge",
		"r5a.xlarge",
		"r5d.12xlarge",
		"r5d.16xlarge",
		"r5d.24xlarge",
		"r5d.2xlarge",
		"r5d.4xlarge",
		"r5d.8xlarge",
		"r5d.metal",
		"r5d.xlarge",
		"r5n.12xlarge",
		"r5n.16xlarge",
		"r5n.24xlarge",
		"r5n.2xlarge",
		"r5n.4xlarge",
		"r5n.8xlarge",
		"r5n.xlarge",
		"r6g.12xlarge",
		"r6g.16xlarge",
		"r6g.2xlarge",
		"r6g.4xlarge",
		"r6g.8xlarge",
		"r6g.xlarge",
		"x1.32xlarge",
		"z1d.12xlarge",
		"z1d.2xlarge",
		"z1d.3xlarge",
		"z1d.6xlarge",
		"z1d.xlarge",
	}
}
