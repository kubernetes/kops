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

package kops

// Version can be replaced by build tooling
var Version = KOPS_RELEASE_VERSION

// These constants are parsed by build tooling - be careful about changing the formats
const (
	KOPS_RELEASE_VERSION = "1.30.1"
	KOPS_CI_VERSION      = "1.30.2"
)

// GitVersion should be replaced by the makefile
var GitVersion = ""
