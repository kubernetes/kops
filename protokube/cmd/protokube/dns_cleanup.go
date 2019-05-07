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

package main

import (
	"strings"
	"time"

	"k8s.io/klog"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/protokube/pkg/protokube"
)

// removeDNSRecords removes the specified DNS records
func removeDNSRecords(nameList string, dnsProvider protokube.DNSProvider) {
	for {
		var removeRecords []dns.Record
		for _, s := range strings.Split(nameList, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}

			removeRecords = append(removeRecords, dns.Record{
				FQDN:       s,
				RecordType: dns.RecordTypeA,
			})
		}

		if len(removeRecords) == 0 {
			return
		}

		if err := dnsProvider.RemoveRecordsImmediate(removeRecords); err != nil {
			klog.Warningf("error removing records %q, will retry: %v", nameList, err)
		} else {
			klog.Infof("removed DNS records %q", nameList)
			return
		}
		time.Sleep(5 * time.Second)
	}
}
