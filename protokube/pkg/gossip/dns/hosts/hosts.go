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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/golang/glog"
)

const GUARD_BEGIN = "# Begin host entries managed by kops - do not edit"
const GUARD_END = "# End host entries managed by kops"

var hostsFileMutex sync.Mutex

func UpdateHostsFileWithRecords(p string, addrToHosts map[string][]string) error {
	// For safety / sanity, we avoid concurrent updates from one process
	hostsFileMutex.Lock()
	defer hostsFileMutex.Unlock()

	stat, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("error getting file status of %q: %v", p, err)
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading file %q: %v", p, err)
	}

	var out []string
	inGuardBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		k := strings.TrimSpace(line)
		if k == GUARD_BEGIN {
			if inGuardBlock {
				glog.Warningf("/etc/hosts guard-block begin seen while in guard block; will ignore")
			}
			inGuardBlock = true
		}

		if !inGuardBlock {
			out = append(out, line)
		}

		if k == GUARD_END {
			if !inGuardBlock {
				glog.Warningf("/etc/hosts guard-block end seen before guard-block start; will ignore end")
				// Don't output the line
				out = out[:len(out)-1]
			}

			inGuardBlock = false
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

	var block []string
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
		glog.V(2).Infof("skipping update of unchanged /etc/hosts")
		return nil
	}

	// Note that because we are bind mounting /etc/hosts, we can't do a normal atomic file write
	// (where we write a temp file and rename it)
	// TODO: We should just hold the file open while we read & write it
	err = ioutil.WriteFile(p, updated, stat.Mode().Perm())
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", p, err)
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
