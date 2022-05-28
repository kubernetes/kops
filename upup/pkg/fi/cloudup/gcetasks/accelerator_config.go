package gcetasks

import (
	"k8s.io/kops/upup/pkg/fi"
)

// AcceleratorConfig defines an accelerator config
type AcceleratorConfig struct {
	AcceleratorCount int64  `json:"acceleratorCount,omitempty"`
	AcceleratorType  string `json:"acceleratorType,omitempty"`
}

var (
	_ fi.HasDependencies = &AcceleratorConfig{}
)

func (a *AcceleratorConfig) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}
