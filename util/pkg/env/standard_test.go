/*
Copyright 2020 The Kubernetes Authors.

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

package env

import (
	"os"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestAddEnvVariable(t *testing.T) {
	tc := struct {
		TestName       string
		EnvKey         string
		EnvValue       string
		ExpectedEnvVar []v1.EnvVar
	}{
		TestName: "Test adding an environment variable",
		EnvKey:   "fake_var_name",
		EnvValue: "fake_var_value",
		ExpectedEnvVar: []v1.EnvVar{
			{
				Name:  "fake_var_name",
				Value: "fake_var_value",
			},
		},
	}

	envVar := EnvVars{}
	t.Run(tc.TestName, func(t *testing.T) {
		os.Setenv(tc.EnvKey, tc.EnvValue)
		envVar.addEnvVariableIfExist(tc.EnvKey)
		envVars := envVar.ToEnvVars()
		if !reflect.DeepEqual(envVars, tc.ExpectedEnvVar){
			t.Errorf("Test failed, unexpected environment varibale %s", tc.EnvKey)
		}
	})
}
