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

package protokube

import (
	"fmt"
	"time"
)

const defaultTTL = time.Minute

type DNSProvider interface {
	Set(fqdn string, recordType string, value string, ttl time.Duration) error
}

// CreateInternalDNSNameRecord maps a FQDN to the internal IP address of the current machine
func (k *KubeBoot) CreateInternalDNSNameRecord(fqdn string) error {
	err := k.DNS.Set(fqdn, "A", k.InternalIP.String(), defaultTTL)
	if err != nil {
		return fmt.Errorf("error configuring DNS name %q: %v", fqdn, err)
	}
	return nil
}

// BuildInternalDNSName builds a DNS name for use inside the cluster, adding our internal DNS suffix to the key,
func (k *KubeBoot) BuildInternalDNSName(key string) string {
	fqdn := key + k.InternalDNSSuffix
	return fqdn
}
