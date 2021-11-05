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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestNTPAmazonLinux2(t *testing.T) {
	RunGoldenTest(t, "tests/ntpbuilder/amazonlinux2", "ntp", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Distribution = distributions.DistributionAmazonLinux2
		builder := NTPBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestNTPDebian11(t *testing.T) {
	RunGoldenTest(t, "tests/ntpbuilder/debian11", "ntp", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Distribution = distributions.DistributionDebian11
		builder := NTPBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}
func TestNTPUbuntu2004(t *testing.T) {
	RunGoldenTest(t, "tests/ntpbuilder/ubuntu2004", "ntp", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Distribution = distributions.DistributionUbuntu2004
		builder := NTPBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestNTPUbuntu2004Chrony(t *testing.T) {
	RunGoldenTest(t, "tests/ntpbuilder/ubuntu2004chrony", "ntp", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Distribution = distributions.DistributionUbuntu2004
		nodeupModelContext.Cluster.Spec.NTP = &kops.NTPConfig{
			Chrony: true,
		}
		builder := NTPBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestNTPUbuntu2110(t *testing.T) {
	RunGoldenTest(t, "tests/ntpbuilder/ubuntu2110", "ntp", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Distribution = distributions.DistributionUbuntu2110
		builder := NTPBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}
