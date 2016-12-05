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

package test

import (
	"fmt"
	"testing"
)

const UPGRADE_CLUSTER = `upgrade cluster \
--name %s \
--state %s \
--yes \
-v %c`

const ROLLING_UPDATE_CLUSTER = `rolling-update cluster \
--name %s \
--state %s \
--yes \
-v %c`

func TestRollingUpdate(t *testing.T) {

	kopsUpgradeCommand := fmt.Sprintf(UPGRADE_CLUSTER, TestClusterName, TestStateStore, TestVerbosity)
	stdout, stderr := ExecOuput(KopsPath, kopsUpgradeCommand,[]string{})
	if stderr != nil {
		t.Errorf("Unable to delete cluster: %v\n%s", stderr, stdout)
	}

	kopsRollingUpdateCommand := fmt.Sprintf(ROLLING_UPDATE_CLUSTER, TestClusterName, TestStateStore, TestVerbosity)
	stdout, stderr = ExecOuput(KopsPath, kopsRollingUpdateCommand,[]string{})
	if stderr != nil {
		t.Errorf("Unable to delete cluster: %v\n%s", stderr, stdout)
	}

	err := Validate()
	if err != nil {
		t.Error(err)
	}

}
