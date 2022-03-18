/*
Copyright 2022 The Kubernetes Authors.

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

package gcetasks

import (
	"fmt"

	"google.golang.org/api/dns/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type ManagedZone struct {
	Name        *string
	Description *string
	DNSName     *string
	Visibility  *string
	Labels      map[string]string
	Lifecycle   fi.Lifecycle
}

func (e *ManagedZone) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *ManagedZone) Find(c *fi.Context) (*ManagedZone, error) {
	cloud := c.Cloud.(gce.GCECloud)

	if e.Name == nil {
		return nil, nil
	}
	if e.DNSName == nil {
		return nil, nil
	}

	mzs, err := cloud.CloudDNS().ManagedZones().List(cloud.Project())
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Managed Zones: %v", err)
	}
	for _, mz := range mzs {
		if mz.DnsName == *e.DNSName && mz.Name == *e.Name {
			actual := &ManagedZone{}
			actual.Lifecycle = e.Lifecycle
			actual.Name = &mz.Name
			actual.DNSName = &mz.DnsName
			actual.Visibility = &mz.Visibility
			return actual, nil
		}
	}
	return nil, nil
}

func (_ *ManagedZone) CheckChanges(a, e, changes *ManagedZone) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	if e.DNSName == nil {
		return fi.RequiredField("DNSName")
	}
	return nil
}

func (_ *ManagedZone) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ManagedZone) error {
	if a == nil {
		klog.Infof("Creating zone: %v", e)
		visibility := ""
		if e.Visibility != nil {
			visibility = *e.Visibility
		}
		if visibility != "private" {
			return fmt.Errorf("cannot create zone with visibility %q - only \"private\" zones can be created automatically", visibility)
		}
		description := fmt.Sprintf("DNS records for kops cluster %v.", *e.DNSName)
		if e.Description != nil {
			description = *e.Description
		}
		err := t.Cloud.CloudDNS().ManagedZones().Insert(t.Cloud.Project(),
			&dns.ManagedZone{
				Name:        *e.Name,
				DnsName:     *e.DNSName,
				Visibility:  visibility,
				Description: description,
				Labels:      e.Labels,
			})
		if err != nil {
			return err
		}
	} else {
		klog.Infof("Found zone already existing: %v", a)
	}
	return nil
}

type terraformManagedZone struct {
	Name        string            `cty:"name"`
	DNSName     string            `cty:"dns_name"`
	Description string            `cty:"description"`
	Labels      map[string]string `cty:"labels"`
}

func (_ *ManagedZone) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ManagedZone) error {
	name := fi.StringValue(e.Name)

	tf := &terraformManagedZone{
		Name:        fi.StringValue(e.Name),
		DNSName:     fi.StringValue(e.DNSName),
		Description: fi.StringValue(e.Description),
		Labels:      e.Labels,
	}

	return t.RenderResource("google_dns_managed_zone", name, tf)
}
