/*
Copyright 2024 The Kubernetes Authors.

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

package main

//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 output:dir=config/crds crd:crdVersions=v1 paths=./bootstrap/kops/api/...;./controlplane/kops/api/...

//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 object paths=./snapshot/cluster-api/...;./bootstrap/kops/api/...;./controlplane/kops/api/...
