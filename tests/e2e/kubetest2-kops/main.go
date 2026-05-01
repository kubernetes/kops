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

package main

import (
	"os"

	"sigs.k8s.io/kubetest2/pkg/app"

	"k8s.io/kops/tests/e2e/kubetest2-kops/deployer"
)

func main() {
	// Prow's Azure WI preset exports AZURE_STORAGE_ACCOUNT; kOps now rejects it.
	os.Unsetenv("AZURE_STORAGE_ACCOUNT")

	app.Main(deployer.Name, deployer.New)
}
