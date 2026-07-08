/*
Copyright 2023 The Kubernetes Authors.

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

// Package openstackcloudconfig renders the OpenStack cloud.conf contents; it
// is used both by the cloudup model and by nodeup on the nodes, and is kept
// separate so that nodeup does not link the full cloud provider
// implementation.
package openstackcloudconfig

import (
	"fmt"
	"os"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func MakeCloudConfig(osc *kops.OpenstackSpec) []string {
	var lines []string

	// Support mapping of older keystone API
	tenantName := os.Getenv("OS_TENANT_NAME")
	if tenantName == "" {
		tenantName = os.Getenv("OS_PROJECT_NAME")
	}
	tenantID := os.Getenv("OS_TENANT_ID")
	if tenantID == "" {
		tenantID = os.Getenv("OS_PROJECT_ID")
	}
	lines = append(lines,
		fmt.Sprintf("auth-url=\"%s\"", os.Getenv("OS_AUTH_URL")),
		fmt.Sprintf("username=\"%s\"", os.Getenv("OS_USERNAME")),
		fmt.Sprintf("password=\"%s\"", os.Getenv("OS_PASSWORD")),
		fmt.Sprintf("region=\"%s\"", os.Getenv("OS_REGION_NAME")),
		fmt.Sprintf("tenant-id=\"%s\"", tenantID),
		fmt.Sprintf("tenant-name=\"%s\"", tenantName),
		fmt.Sprintf("domain-name=\"%s\"", os.Getenv("OS_DOMAIN_NAME")),
		fmt.Sprintf("domain-id=\"%s\"", os.Getenv("OS_DOMAIN_ID")),
		fmt.Sprintf("application-credential-id=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_ID")),
		fmt.Sprintf("application-credential-secret=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET")),
		"",
	)

	if lb := osc.Loadbalancer; lb != nil {
		ingressHostnameSuffix := "nip.io"
		if fi.ValueOf(lb.IngressHostnameSuffix) != "" {
			ingressHostnameSuffix = fi.ValueOf(lb.IngressHostnameSuffix)
		}

		lines = append(lines,
			"[LoadBalancer]",
			fmt.Sprintf("floating-network-id=%s", fi.ValueOf(lb.FloatingNetworkID)),
			fmt.Sprintf("lb-method=%s", fi.ValueOf(lb.Method)),
			fmt.Sprintf("lb-provider=%s", fi.ValueOf(lb.Provider)),
			fmt.Sprintf("use-octavia=%t", fi.ValueOf(lb.UseOctavia)),
			fmt.Sprintf("manage-security-groups=%t", fi.ValueOf(lb.ManageSecGroups)),
			fmt.Sprintf("enable-ingress-hostname=%t", fi.ValueOf(lb.EnableIngressHostname)),
			fmt.Sprintf("ingress-hostname-suffix=%s", ingressHostnameSuffix),
			"",
		)

		if monitor := osc.Monitor; monitor != nil {
			lines = append(lines,
				"create-monitor=yes",
				fmt.Sprintf("monitor-delay=%s", fi.ValueOf(monitor.Delay)),
				fmt.Sprintf("monitor-timeout=%s", fi.ValueOf(monitor.Timeout)),
				fmt.Sprintf("monitor-max-retries=%d", fi.ValueOf(monitor.MaxRetries)),
				"",
			)
		}
	}

	if bs := osc.BlockStorage; bs != nil {
		// Block Storage Config
		lines = append(lines,
			"[BlockStorage]",
			fmt.Sprintf("bs-version=%s", fi.ValueOf(bs.Version)),
			fmt.Sprintf("ignore-volume-az=%t", fi.ValueOf(bs.IgnoreAZ)),
			fmt.Sprintf("ignore-volume-microversion=%t", fi.ValueOf(bs.IgnoreVolumeMicroVersion)),
			"")
	}

	if networking := osc.Network; networking != nil {
		// Networking Config
		// https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/openstack-cloud-controller-manager/using-openstack-cloud-controller-manager.md#networking
		var networkingLines []string

		if networking.IPv6SupportDisabled != nil {
			networkingLines = append(networkingLines, fmt.Sprintf("ipv6-support-disabled=%t", fi.ValueOf(networking.IPv6SupportDisabled)))
		}
		for _, name := range networking.PublicNetworkNames {
			networkingLines = append(networkingLines, fmt.Sprintf("public-network-name=%s", fi.ValueOf(name)))
		}
		for _, name := range networking.InternalNetworkNames {
			networkingLines = append(networkingLines, fmt.Sprintf("internal-network-name=%s", fi.ValueOf(name)))
		}
		if networking.AddressSortOrder != nil {
			networkingLines = append(networkingLines, fmt.Sprintf("address-sort-order=%s", fi.ValueOf(networking.AddressSortOrder)))
		}

		if len(networkingLines) > 0 {
			lines = append(lines, "[Networking]")
			lines = append(lines, networkingLines...)
			lines = append(lines, "")
		}
	}

	return lines
}
