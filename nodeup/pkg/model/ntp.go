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

package model

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// NTPBuilder installs and starts NTP, to ensure accurate clock times.
// As well as general log confusion, clock-skew of more than 5 minutes
// causes AWS API calls to fail
type NTPBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NTPBuilder{}

type ntpDaemon string

var (
	chronyd ntpDaemon = "chronyd"
	ntpd    ntpDaemon = "ntpd"
)

// Build is responsible for configuring NTP
func (b *NTPBuilder) Build(c *fi.ModelBuilderContext) error {
	switch b.Distribution {
	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install ntp")
		return nil
	case distros.DistributionCoreOS:
		klog.Infof("Detected CoreOS; won't install ntp")
		return nil
	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install ntp")
		return nil
	}

	var ntpIP string
	switch b.Cluster.Spec.CloudProvider {
	case "aws":
		ntpIP = "169.254.169.123"
	case "gce":
		ntpIP = "time.google.com"
	default:
		ntpIP = ""
	}

	if b.Distribution.IsDebianFamily() {
		c.AddTask(&nodetasks.Package{Name: "ntp"})

		if ntpIP != "" {
			bytes, err := updateNtpIP(ntpIP, ntpd)
			if err != nil {
				return err
			}
			c.AddTask(&nodetasks.File{
				Path:     "/etc/ntp.conf",
				Contents: fi.NewBytesResource(bytes),
				Type:     nodetasks.FileType_File,
				Mode:     s("0644"),
			})
		}

		c.AddTask((&nodetasks.Service{Name: "ntp"}).InitDefaults())
	} else if b.Distribution.IsRHELFamily() {
		switch b.Distribution {
		case distros.DistributionCentos8, distros.DistributionRhel8:
			c.AddTask(&nodetasks.Package{Name: "chrony"})

			if ntpIP != "" {
				bytes, err := updateNtpIP(ntpIP, chronyd)
				if err != nil {
					return err
				}
				c.AddTask(&nodetasks.File{
					Path:     "/etc/chrony.conf",
					Contents: fi.NewBytesResource(bytes),
					Type:     nodetasks.FileType_File,
					Mode:     s("0644"),
				})
			}
			c.AddTask((&nodetasks.Service{Name: "chronyd"}).InitDefaults())

		default:
			c.AddTask(&nodetasks.Package{Name: "ntp"})

			if ntpIP != "" {
				bytes, err := updateNtpIP(ntpIP, ntpd)
				if err != nil {
					return err
				}
				c.AddTask(&nodetasks.File{
					Path:     "/etc/ntp.conf",
					Contents: fi.NewBytesResource(bytes),
					Type:     nodetasks.FileType_File,
					Mode:     s("0644"),
				})
			}

			c.AddTask((&nodetasks.Service{Name: "ntpd"}).InitDefaults())
		}
	} else {
		klog.Warningf("unknown distribution, skipping ntp install: %v", b.Distribution)
		return nil
	}
	return nil
}

// updateNtpIP takes a ip and a ntpDaemon and will comment out
// the default server or pool values and append the correct cloud
// ip to the ntp config file.
func updateNtpIP(ip string, daemon ntpDaemon) ([]byte, error) {
	var address string
	var r *regexp.Regexp
	var path string
	switch ntpd {
	case ntpd:
		address = fmt.Sprintf("server %s prefer iburst", ip)
		// the regex strings might need a bit more work
		r = regexp.MustCompile(`pool\s\d.*[a-z].[a-z].[a-z]\siburst`)
		path = "/etc/ntp.conf"
	case chronyd:
		address = fmt.Sprintf("server %s prefer iburst minpoll 4 maxpoll 4", ip)
		// the regex strings might need a bit more work
		r = regexp.MustCompile(`server\s.*iburst.*`)
		path = "/etc/chrony.conf"
	default:
		return nil, fmt.Errorf("%s is not a supported ntp application", ntpd)
	}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	new := r.ReplaceAllFunc(f, func(b []byte) []byte {
		return []byte(fmt.Sprintf("#commented out by kops %s", string(b)))
	})
	new = append(new, []byte(address)...)
	return new, nil
}
