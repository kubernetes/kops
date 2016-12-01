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

package awstasks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"testing"
)

func TestParseRemovalRule(t *testing.T) {
	testNotParse(t, "port 22")
	testNotParse(t, "port22")
	testNotParse(t, "port=a")
	testNotParse(t, "port=22-23")

	testParsesAsPort(t, "port=22", 22)
	testParsesAsPort(t, "port=443", 443)
}

func testNotParse(t *testing.T, rule string) {
	r, err := ParseRemovalRule(rule)
	if err == nil {
		t.Fatalf("expected failure to parse removal rule %q, got %v", rule, r)
	}
}

func testParsesAsPort(t *testing.T, rule string, port int) {
	r, err := ParseRemovalRule(rule)
	if err != nil {
		t.Fatalf("unexpected failure to parse rule %q: %v", rule, err)
	}
	portRemovalRule, ok := r.(*PortRemovalRule)
	if !ok {
		t.Fatalf("unexpected rule type for rule %q: %T", r, err)
	}
	if portRemovalRule.Port != port {
		t.Fatalf("unexpected port for %q, expecting %d, got %q", rule, port, r)
	}
}

func TestPortRemovalRule(t *testing.T) {
	r := &PortRemovalRule{Port: 22}
	testMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(22), ToPort: aws.Int64(22)})

	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(0), ToPort: aws.Int64(0)})
	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(23), ToPort: aws.Int64(23)})
	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(20), ToPort: aws.Int64(22)})
	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(22), ToPort: aws.Int64(23)})
	testNotMatches(t, r, &ec2.IpPermission{ToPort: aws.Int64(22)})
	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(22)})
	testNotMatches(t, r, &ec2.IpPermission{})
}

func TestPortRemovalRule_Zero(t *testing.T) {
	r := &PortRemovalRule{Port: 0}
	testMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(0), ToPort: aws.Int64(0)})

	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(0), ToPort: aws.Int64(20)})
	testNotMatches(t, r, &ec2.IpPermission{ToPort: aws.Int64(0)})
	testNotMatches(t, r, &ec2.IpPermission{FromPort: aws.Int64(0)})
	testNotMatches(t, r, &ec2.IpPermission{})
}

func testMatches(t *testing.T, rule *PortRemovalRule, permission *ec2.IpPermission) {
	if !rule.Matches(permission) {
		t.Fatalf("rule %q failed to match permission %q", rule, permission)
	}
}

func testNotMatches(t *testing.T, rule *PortRemovalRule, permission *ec2.IpPermission) {
	if rule.Matches(permission) {
		t.Fatalf("rule %q unexpectedly matched permission %q", rule, permission)
	}
}
