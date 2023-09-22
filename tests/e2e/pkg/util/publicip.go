/*
Copyright 2020 The Kubernetes Authors.

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

package util

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var externalIPServiceURLs = []string{
	"https://ip.jsb.workers.dev",
	"https://ifconfig.co",
}

// ExternalIPRange returns the CIDR block for the public IP
// in front of the kubetest2 client
func ExternalIPRange() (string, error) {
	var b bytes.Buffer

	for attempt := 0; attempt < 5; attempt++ {
		for _, u := range externalIPServiceURLs {
			b.Reset()
			err := HTTPGETWithHeaders(u, nil, &b)
			if err != nil {
				// The external service may well be down
				log.Printf("failed to get external ip from %s: %v", u, err)
			} else if ip := net.ParseIP(strings.TrimSpace(b.String())); ip != nil {
				if ip.To4() != nil {
					return ip.String() + "/32", nil
				}
				return ip.String() + "/128", nil
			} else {
				log.Printf("service %s returned invalid ip %q", u, b.String())
			}
		}

		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("external IP cannot be retrieved")
}
