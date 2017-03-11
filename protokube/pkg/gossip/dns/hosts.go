/*
Copyright 2017 The Kubernetes Authors.

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

package dns

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const GUARD_BEGIN = "# Begin host entries managed by kops - do not edit"
const GUARD_END = "# End host entries managed by kops"

// HostsFile stores DNS records into /etc/hosts
type HostsFile struct {
	Path string
}

var _ DNSTarget = &HostsFile{}

func (h *HostsFile) Update(snapshot *DNSViewSnapshot) error {
	glog.V(2).Infof("Updating hosts file with snapshot version %v", snapshot.version)

	addrToHosts := make(map[string][]string)

	zones := snapshot.ListZones()
	for _, zone := range zones {
		records := snapshot.RecordsForZone(zone)

		for _, record := range records {
			if record.RrsType != "A" {
				glog.Warningf("skipping record of unhandled type: %v", record)
				continue
			}

			for _, addr := range record.Rrdatas {
				addrToHosts[addr] = append(addrToHosts[addr], record.Name)
			}
		}
	}

	stat, err := os.Stat(h.Path)
	if err != nil {
		return fmt.Errorf("error getting file status of %q: %v", h.Path, err)
	}

	data, err := ioutil.ReadFile(h.Path)
	if err != nil {
		return fmt.Errorf("error reading file %q: %v", h.Path, err)
	}

	var out []string
	depth := 0
	for _, line := range strings.Split(string(data), "\n") {
		k := strings.TrimSpace(line)
		if k == GUARD_BEGIN {
			depth++
		}

		if depth <= 0 {
			out = append(out, line)
		}

		if k == GUARD_END {
			depth--
		}
	}

	// Ensure a single blank line
	for {
		if len(out) == 0 {
			break
		}

		if out[len(out)-1] != "" {
			break
		}

		out = out[:len(out)-1]
	}
	out = append(out, "")

	out = append(out, GUARD_BEGIN)
	for addr, hosts := range addrToHosts {
		sort.Strings(hosts)
		out = append(out, addr+"\t"+strings.Join(hosts, " "))
	}
	out = append(out, GUARD_END)
	out = append(out, "")

	// Note that because we are bind mounting /etc/hosts, we can't do a normal atomic file write
	// (where we write a temp file and rename it)
	// TODO: We should just hold the file open while we read & write it
	err = ioutil.WriteFile(h.Path, []byte(strings.Join(out, "\n")), stat.Mode().Perm())
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", h.Path, err)
	}

	return nil
}

func atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)

	tempFile, err := ioutil.TempFile(dir, ".tmp"+filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("error creating temp file in %q: %v", dir, err)
	}

	mustClose := true
	mustRemove := true

	defer func() {
		if mustClose {
			if err := tempFile.Close(); err != nil {
				glog.Warningf("error closing temp file: %v", err)
			}
		}

		if mustRemove {
			if err := os.Remove(tempFile.Name()); err != nil {
				glog.Warningf("error removing temp file %q: %v", tempFile.Name(), err)
			}
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	mustClose = false

	if err := os.Chmod(tempFile.Name(), perm); err != nil {
		return fmt.Errorf("error changing mode of temp file: %v", err)
	}

	if err := os.Rename(tempFile.Name(), filename); err != nil {
		return fmt.Errorf("error moving temp file %q to %q: %v", tempFile.Name(), filename, err)
	}

	mustRemove = false
	return nil
}
