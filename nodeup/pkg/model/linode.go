/*
Copyright 2026 The Kubernetes Authors.

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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// LinodeBuilder writes the Akamai (Linode)-specific configuration
type LinodeBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &LinodeBuilder{}

// Build configures Akamai (Linode)-specific node settings including swap disabling.
func (b *LinodeBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.CloudProvider() != kops.CloudProviderLinode {
		return nil
	}

	// Akamai (Linode) instances ship with swap enabled. Disable it when MemorySwapBehavior is unset,
	// since the default kubelet swap mode (NoSwap) requires swap to be off.
	if b.NodeupConfig.KubeletConfig.MemorySwapBehavior == "" {
		if err := b.disableSwap(c); err != nil {
			return fmt.Errorf("error disabling swap: %v", err)
		}
	}

	return nil
}

// disableSwap disables swap at the OS level and removes swap entries from /etc/fstab
func (b *LinodeBuilder) disableSwap(c *fi.NodeupModelBuilderContext) error {
	// Read current /etc/fstab
	fstabPath := "/etc/fstab"
	fstabBytes, err := os.ReadFile(fstabPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", fstabPath, err)
	}

	// Filter out swap entries
	var filteredLines []string
	scanner := bufio.NewScanner(bytes.NewReader(fstabBytes))
	for scanner.Scan() {
		line := scanner.Text()
		// Keep lines that don't contain " swap " or "\tswap\t"
		if !strings.Contains(line, " swap ") && !strings.Contains(line, "\tswap\t") {
			filteredLines = append(filteredLines, line)
		} else {
			klog.V(2).Infof("Removing swap entry from fstab: %s", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading fstab: %w", err)
	}

	// Write filtered fstab back
	filteredContent := strings.Join(filteredLines, "\n") + "\n"
	c.AddTask(&nodetasks.File{
		Path:     fstabPath,
		Contents: fi.NewStringResource(filteredContent),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
		// Execute swapoff after updating fstab
		OnChangeExecute: [][]string{{"swapoff", "-a"}},
	})

	return nil
}
