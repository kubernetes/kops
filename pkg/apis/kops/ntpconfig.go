/*
Copyright 2021 The Kubernetes Authors.

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

package kops

// NTPConfig is the configuration for NTP.
type NTPConfig struct {
	// Managed controls if the NTP configuration is managed by kOps.
	// The NTP configuration task is skipped if this is set to false.
	Managed *bool `json:"managed,omitempty"`
}
