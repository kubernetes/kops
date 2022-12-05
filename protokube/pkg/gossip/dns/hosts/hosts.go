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

package hosts

import (
	"bytes"
	"fmt"
	math_rand "math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

const (
	GUARD_BEGIN = "# Begin host entries managed by kops - do not edit"
	GUARD_END   = "# End host entries managed by kops"
)

var hostsFileMutex sync.Mutex

// HostMap holds a set of host to address mappings, a simplification of /etc/hosts.
type HostMap struct {
	records []hostMapRecord
}

// ParseHostMap parses lines from /etc/hosts (expected to be our guarded block) into a HostMap structure.
// It parses as much as it can, and returns invalid lines (which should ideally be empty)
func (m *HostMap) Parse(existing []string) []string {
	var badLines []string

	for _, line := range existing {
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		if strings.HasPrefix(tokens[0], "#") {
			// Comments shouldn't really happen in our guarded block
			if line == GUARD_BEGIN || line == GUARD_END {
				klog.Warningf("ignoring extra guard line in /etc/hosts: %q", line)
			} else {
				badLines = append(badLines, line)
			}
			continue
		}

		if len(tokens) == 1 {
			badLines = append(badLines, line)
			continue
		}

		address := tokens[0]
		for _, hostname := range tokens[1:] {
			m.records = append(m.records, hostMapRecord{
				Address:  address,
				Hostname: hostname,
			})
		}
	}

	return badLines
}

// hostMap holds a single host-name to address mapping.
type hostMapRecord struct {
	Hostname string
	Address  string
}

// ReplaceRecords replaces all the addresses for the given hostname.
func (m *HostMap) ReplaceRecords(hostname string, addresses []string) {
	var newRecords []hostMapRecord

	for _, address := range addresses {
		newRecords = append(newRecords, hostMapRecord{
			Hostname: hostname,
			Address:  address,
		})
	}

	for _, record := range m.records {
		if record.Hostname == hostname {
			continue
		}
		newRecords = append(newRecords, record)
	}

	m.records = newRecords
}

// UpdateHostsFileWithRecords updates /etc/hosts by applying the given mutation function.
func UpdateHostsFileWithRecords(p string, mutator func(guarded []string) (*HostMap, error)) error {
	// For safety / sanity, we avoid concurrent updates from one process
	hostsFileMutex.Lock()
	defer hostsFileMutex.Unlock()

	stat, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("error getting file status of %q: %v", p, err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading file %q: %v", p, err)
	}

	var guarded []string
	var out []string
	inGuardBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		k := strings.TrimSpace(line)
		if k == GUARD_BEGIN {
			if inGuardBlock {
				klog.Warningf("/etc/hosts guard-block begin seen while in guard block; will ignore")
			}
			inGuardBlock = true
			// don't output guard lines into guarded block
			continue
		}

		if k == GUARD_END {
			if !inGuardBlock {
				klog.Warningf("/etc/hosts guard-block end seen before guard-block start; will ignore end")
			}
			inGuardBlock = false
			// don't output guard lines into guarded block
			continue
		}

		if inGuardBlock {
			guarded = append(guarded, line)
		} else {
			out = append(out, line)
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

	hosts, err := mutator(guarded)
	if err != nil {
		return err
	}

	var block []string
	addrToHosts := make(map[string][]string)
	for _, record := range hosts.records {
		addrToHosts[record.Address] = append(addrToHosts[record.Address], record.Hostname)
	}
	for addr, hosts := range addrToHosts {
		sort.Strings(hosts)
		block = append(block, addr+"\t"+strings.Join(hosts, " "))
	}
	// Sort into a consistent order to minimize updates
	sort.Strings(block)

	out = append(out, GUARD_BEGIN)
	out = append(out, block...)
	out = append(out, GUARD_END)
	out = append(out, "")

	updated := []byte(strings.Join(out, "\n"))

	if bytes.Equal(updated, data) {
		klog.V(2).Infof("skipping update of unchanged /etc/hosts")
		return nil
	}

	// Note that because we are bind mounting /etc/hosts, we can't do a normal atomic file write
	// (where we write a temp file and rename it)
	if err := pseudoAtomicWrite(p, updated, stat.Mode()); err != nil {
		return fmt.Errorf("error writing file %q: %v", p, err)
	}

	return nil
}

// Because we are bind-mounting /etc/hosts, we can't do a normal
// atomic file write (where we write a temp file and rename it);
// instead we write the file, pause, re-read and see if anyone else
// wrote in the meantime; if so we rewrite again.  By pausing for a
// random amount of time, eventually we'll win the write race and
// exit.  This doesn't guarantee fairness, but it should mean that the
// end-result is not malformed (i.e. partial writes).
func pseudoAtomicWrite(p string, b []byte, mode os.FileMode) error {
	attempt := 0
	for {
		attempt++
		if attempt > 10 {
			return fmt.Errorf("failed to consistently write file %q - too many retries", p)
		}

		if err := os.WriteFile(p, b, mode); err != nil {
			klog.Warningf("error writing file %q: %v", p, err)
			continue
		}

		n := 1 + math_rand.Intn(20)
		time.Sleep(time.Duration(n) * time.Millisecond)

		contents, err := os.ReadFile(p)
		if err != nil {
			klog.Warningf("error re-reading file %q: %v", p, err)
			continue
		}

		if bytes.Equal(contents, b) {
			return nil
		}

		klog.Warningf("detected concurrent write to file %q, will retry", p)
	}
}
