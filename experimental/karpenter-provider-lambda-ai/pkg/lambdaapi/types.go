/*
Copyright The Kubernetes Authors.

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

package lambdaapi

type Instance struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	IP              string       `json:"ip"`
	Status          string       `json:"status"`
	InstanceType    InstanceType `json:"instance_type"`
	Region          Region       `json:"region"`
	SSHKeyNames     []string     `json:"ssh_key_names"`
	FileSystemNames []string     `json:"file_system_names"`
}

type InstanceType struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	PriceCentsPerHour int    `json:"price_cents_per_hour"`
	Specs             Specs  `json:"specs"`
}

type Specs struct {
	VCPUs        int `json:"vcpus"`
	MemoryGib    int `json:"memory_gib"`
	GPUs         int `json:"gpus"`
	GPUMemoryGib int `json:"gpu_memory_gib"`
}

type Region struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LaunchRequest struct {
	RegionName       string   `json:"region_name"`
	InstanceTypeName string   `json:"instance_type_name"`
	SSHKeyNames      []string `json:"ssh_key_names"`
	FileSystemNames  []string `json:"file_system_names,omitempty"`
	Quantity         int      `json:"quantity,omitempty"`
	Name             string   `json:"name,omitempty"`
}

type LaunchResponse struct {
	Data struct {
		InstanceIDs []string `json:"instance_ids"`
	} `json:"data"`
}

type TerminateRequest struct {
	InstanceIDs []string `json:"instance_ids"`
}

type TerminateResponse struct {
	Data struct {
		TerminatedInstances []Instance `json:"terminated_instances"`
	} `json:"data"`
}

type ListInstancesResponse struct {
	Data []Instance `json:"data"`
}

type ListInstanceTypesResponse struct {
	Data map[string]InstanceType `json:"data"`
}

type ListRegionsResponse struct {
	Data []Region `json:"data"`
}

type SSHKey struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type ListSSHKeysResponse struct {
	Data []SSHKey `json:"data"`
}
