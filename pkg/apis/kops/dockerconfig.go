/*
Copyright 2016 The Kubernetes Authors.

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

import (
	"k8s.io/kubernetes/pkg/util/validation/field"
)

type DockerConfig struct {
	Bridge   *string `json:"bridge,omitempty" flag:"bridge"`
	LogLevel *string `json:"logLevel,omitempty" flag:"log-level"`
	IPTables *bool   `json:"ipTables,omitempty" flag:"iptables"`
	IPMasq   *bool   `json:"ipMasq,omitempty" flag:"ip-masq"`

	// Storage maps to the docker storage flag
	// But nodeup will also process a comma-separate list, selecting the first supported option
	Storage *string `json:"storage,omitempty" flag:"storage-driver"`

	InsecureRegistry *string `json:"insecureRegistry,omitempty" flag:"insecure-registry"`
	MTU              *int32  `json:"mtu,omitempty" flag:"mtu"`

	// The version of docker to install
	// Be careful if changing this; not all docker versions are validated, and they will break in bad ways.
	Version *string `json:"version,omitempty"`
}

var validDockerConfigStorageValues = []string{"aufs", "btrfs", "devicemapper", "overlay", "overlay2", "zfs"}

func ValidateDockerConfig(config *DockerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath.Child("storage"), config.Storage, validDockerConfigStorageValues)...)
	return allErrs
}
