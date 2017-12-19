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

package cloudup

import "k8s.io/apimachinery/pkg/util/sets"

// Phase is a portion of work that kops completes.
type Phase string

const (
	// PhaseStageAssets uploads various assets such as containers in a private registry
	PhaseStageAssets Phase = "assets"
	// PhaseNetwork creates network infrastructure.
	PhaseNetwork Phase = "network"
	// PhaseSecurity creates IAM profiles and roles, security groups and firewalls
	PhaseSecurity Phase = "security"
	// PhaseCluster creates the servers, and load-alancers
	PhaseCluster Phase = "cluster"
)

// Phases are used for validation and cli help.
var Phases = sets.NewString(
	string(PhaseStageAssets),
	string(PhaseSecurity),
	string(PhaseNetwork),
	string(PhaseCluster),
)
