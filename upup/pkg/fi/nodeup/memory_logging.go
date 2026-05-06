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

package nodeup

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

const (
	nodeupLogMemoryEnv         = "NODEUP_LOG_MEMORY"
	nodeupLogMemoryIntervalEnv = "NODEUP_LOG_MEMORY_INTERVAL"
	nodeupLogMemoryInterval    = 30 * time.Second
)

var nodeupCgroupMemoryStatKeys = []string{
	"anon",
	"file",
	"kernel",
	"kernel_stack",
	"pagetables",
	"sock",
	"shmem",
	"slab",
	"file_mapped",
	"file_dirty",
	"file_writeback",
	"inactive_file",
	"active_file",
	"rss",
	"cache",
	"mapped_file",
	"swap",
	"total_rss",
	"total_cache",
	"total_mapped_file",
}

type nodeupMemoryLogger struct {
	enabled  bool
	interval time.Duration
	start    time.Time
}

func newNodeupMemoryLogger(enabled bool) *nodeupMemoryLogger {
	enabled = enabled || nodeupMemoryLoggingEnabledFromEnv()
	interval := nodeupLogMemoryInterval
	if enabled {
		interval = nodeupMemoryLoggingIntervalFromEnv()
	}

	return &nodeupMemoryLogger{
		enabled:  enabled,
		interval: interval,
		start:    time.Now(),
	}
}

func nodeupMemoryLoggingEnabledFromEnv() bool {
	value := strings.TrimSpace(os.Getenv(nodeupLogMemoryEnv))
	if value == "" {
		return false
	}

	enabled, err := strconv.ParseBool(value)
	if err != nil {
		klog.Warningf("ignoring invalid %s=%q: %v", nodeupLogMemoryEnv, value, err)
		return false
	}
	return enabled
}

func nodeupMemoryLoggingIntervalFromEnv() time.Duration {
	value := strings.TrimSpace(os.Getenv(nodeupLogMemoryIntervalEnv))
	if value == "" {
		return nodeupLogMemoryInterval
	}

	interval, err := time.ParseDuration(value)
	if err != nil {
		klog.Warningf("ignoring invalid %s=%q: %v", nodeupLogMemoryIntervalEnv, value, err)
		return nodeupLogMemoryInterval
	}
	if interval < 0 {
		klog.Warningf("ignoring negative %s=%q", nodeupLogMemoryIntervalEnv, value)
		return nodeupLogMemoryInterval
	}
	return interval
}

func (l *nodeupMemoryLogger) Log(phase string, taskSummary *nodeupMemoryTaskSummary) {
	if l == nil || !l.enabled {
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	cgroupMemory := readNodeupCgroupMemory()

	fields := []string{
		fmt.Sprintf("phase=%q", phase),
		fmt.Sprintf("elapsed=%s", time.Since(l.start).Round(time.Millisecond)),
		fmt.Sprintf("go_heap_alloc_bytes=%d", mem.HeapAlloc),
		fmt.Sprintf("go_heap_inuse_bytes=%d", mem.HeapInuse),
		fmt.Sprintf("go_heap_sys_bytes=%d", mem.HeapSys),
		fmt.Sprintf("go_stack_inuse_bytes=%d", mem.StackInuse),
		fmt.Sprintf("go_sys_bytes=%d", mem.Sys),
		fmt.Sprintf("go_mallocs=%d", mem.Mallocs),
		fmt.Sprintf("go_frees=%d", mem.Frees),
		fmt.Sprintf("go_num_gc=%d", mem.NumGC),
		fmt.Sprintf("cgroup_version=%q", cgroupMemory.version),
		fmt.Sprintf("cgroup_path=%q", cgroupMemory.path),
		fmt.Sprintf("cgroup_current_bytes=%d", cgroupMemory.current),
		fmt.Sprintf("cgroup_peak_bytes=%d", cgroupMemory.peak),
		fmt.Sprintf("cgroup_stat=%q", formatNodeupCgroupMemoryStats(cgroupMemory.stat)),
	}
	if taskSummary != nil {
		fields = append(fields, taskSummary.logFields()...)
	}

	klog.Infof("nodeup memory stats %s", strings.Join(fields, " "))
}

func (l *nodeupMemoryLogger) StartPeriodic(phase string, taskSummary *nodeupMemoryTaskSummary) func() {
	if l == nil || !l.enabled || l.interval == 0 {
		return func() {}
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Log(phase, taskSummary)
			case <-stop:
				return
			}
		}
	}()

	return func() {
		close(stop)
		<-done
	}
}

type nodeupMemoryTaskSummary struct {
	tasks          int
	files          int
	fileContents   int
	directories    int
	services       int
	packages       int
	issueCerts     int
	kubeconfigs    int
	loadImages     int
	bootstrapTasks int
}

func summarizeNodeupMemoryTasks(tasks map[string]fi.NodeupTask) nodeupMemoryTaskSummary {
	summary := nodeupMemoryTaskSummary{
		tasks: len(tasks),
	}

	for _, task := range tasks {
		switch task := task.(type) {
		case *nodetasks.File:
			summary.files++
			if task.Type == nodetasks.FileType_Directory {
				summary.directories++
			}
			if task.Contents != nil {
				summary.fileContents++
			}
		case *nodetasks.Service:
			summary.services++
		case *nodetasks.Package:
			summary.packages++
		case *nodetasks.IssueCert:
			summary.issueCerts++
		case *nodetasks.KubeConfig:
			summary.kubeconfigs++
		case *nodetasks.LoadImageTask:
			summary.loadImages++
		case *nodetasks.BootstrapClientTask:
			summary.bootstrapTasks++
		}
	}

	return summary
}

func (s *nodeupMemoryTaskSummary) logFields() []string {
	return []string{
		fmt.Sprintf("tasks=%d", s.tasks),
		fmt.Sprintf("file_tasks=%d", s.files),
		fmt.Sprintf("file_content_tasks=%d", s.fileContents),
		fmt.Sprintf("directory_tasks=%d", s.directories),
		fmt.Sprintf("service_tasks=%d", s.services),
		fmt.Sprintf("package_tasks=%d", s.packages),
		fmt.Sprintf("issue_cert_tasks=%d", s.issueCerts),
		fmt.Sprintf("kubeconfig_tasks=%d", s.kubeconfigs),
		fmt.Sprintf("load_image_tasks=%d", s.loadImages),
		fmt.Sprintf("bootstrap_client_tasks=%d", s.bootstrapTasks),
	}
}

type nodeupCgroupMemory struct {
	version string
	path    string
	current int64
	peak    int64
	stat    map[string]int64
}

type nodeupCgroupMemoryCandidate struct {
	version string
	path    string
}

func readNodeupCgroupMemory() nodeupCgroupMemory {
	result := nodeupCgroupMemory{
		version: "unavailable",
		current: -1,
		peak:    -1,
	}

	candidates := nodeupCgroupMemoryCandidates()
	for _, candidate := range candidates {
		current, currentOK := readNodeupInt64File(path.Join(candidate.path, nodeupCgroupMemoryCurrentFile(candidate.version)))
		peak, peakOK := readNodeupInt64File(path.Join(candidate.path, nodeupCgroupMemoryPeakFile(candidate.version)))
		stat, statOK := readNodeupCgroupMemoryStatFile(path.Join(candidate.path, "memory.stat"))
		if !currentOK && !peakOK && !statOK {
			continue
		}

		result.version = candidate.version
		result.path = candidate.path
		if currentOK {
			result.current = current
		}
		if peakOK {
			result.peak = peak
		}
		if statOK {
			result.stat = stat
		}
		return result
	}

	return result
}

func nodeupCgroupMemoryCurrentFile(version string) string {
	if version == "v1" {
		return "memory.usage_in_bytes"
	}
	return "memory.current"
}

func nodeupCgroupMemoryPeakFile(version string) string {
	if version == "v1" {
		return "memory.max_usage_in_bytes"
	}
	return "memory.peak"
}

func nodeupCgroupMemoryCandidates() []nodeupCgroupMemoryCandidate {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return fallbackNodeupCgroupMemoryCandidates()
	}

	candidates := nodeupCgroupMemoryCandidatesFromProcCgroup(string(data))
	if len(candidates) == 0 {
		return fallbackNodeupCgroupMemoryCandidates()
	}
	return candidates
}

func nodeupCgroupMemoryCandidatesFromProcCgroup(data string) []nodeupCgroupMemoryCandidate {
	var candidates []nodeupCgroupMemoryCandidate
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		tokens := strings.SplitN(line, ":", 3)
		if len(tokens) != 3 {
			continue
		}

		hierarchyID, controllers, cgroupPath := tokens[0], tokens[1], tokens[2]
		switch {
		case hierarchyID == "0" && controllers == "":
			candidates = append(candidates, nodeupCgroupMemoryCandidate{
				version: "v2",
				path:    path.Join("/sys/fs/cgroup", cleanNodeupCgroupPath(cgroupPath)),
			})
		case nodeupCgroupControllersIncludeMemory(controllers):
			candidates = append(candidates, nodeupCgroupMemoryCandidate{
				version: "v1",
				path:    path.Join("/sys/fs/cgroup/memory", cleanNodeupCgroupPath(cgroupPath)),
			})
		}
	}

	candidates = append(candidates, fallbackNodeupCgroupMemoryCandidates()...)
	return dedupeNodeupCgroupMemoryCandidates(candidates)
}

func nodeupCgroupControllersIncludeMemory(controllers string) bool {
	for _, controller := range strings.Split(controllers, ",") {
		if controller == "memory" {
			return true
		}
	}
	return false
}

func cleanNodeupCgroupPath(cgroupPath string) string {
	cleaned := path.Clean("/" + cgroupPath)
	if cleaned == "/" {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func fallbackNodeupCgroupMemoryCandidates() []nodeupCgroupMemoryCandidate {
	return []nodeupCgroupMemoryCandidate{
		{version: "v2", path: "/sys/fs/cgroup"},
		{version: "v1", path: "/sys/fs/cgroup/memory"},
	}
}

func dedupeNodeupCgroupMemoryCandidates(candidates []nodeupCgroupMemoryCandidate) []nodeupCgroupMemoryCandidate {
	var deduped []nodeupCgroupMemoryCandidate
	seen := make(map[nodeupCgroupMemoryCandidate]bool)
	for _, candidate := range candidates {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		deduped = append(deduped, candidate)
	}
	return deduped
}

func readNodeupInt64File(filePath string) (int64, bool) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, false
	}
	value := strings.TrimSpace(string(data))
	if value == "" || value == "max" {
		return 0, false
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func readNodeupCgroupMemoryStatFile(filePath string) (map[string]int64, bool) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}
	return parseNodeupCgroupMemoryStat(string(data)), true
}

func parseNodeupCgroupMemoryStat(data string) map[string]int64 {
	stats := make(map[string]int64)
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		stats[fields[0]] = value
	}
	return stats
}

func formatNodeupCgroupMemoryStats(stats map[string]int64) string {
	var parts []string
	for _, key := range nodeupCgroupMemoryStatKeys {
		value, ok := stats[key]
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%d", key, value))
	}
	return strings.Join(parts, ",")
}
