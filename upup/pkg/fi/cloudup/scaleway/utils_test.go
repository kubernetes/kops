/*
Copyright 2023 The Kubernetes Authors.

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

package scaleway

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"sigs.k8s.io/yaml"
)

type ScalewayProfile struct {
	AccessKey        string `json:"access_key"`
	SecretKey        string `json:"secret_key"`
	DefaultProjectID string `json:"default_project_id"`
}

func createScalewayConfigFile() error {
	scalewayDefaultProfile := ScalewayProfile{
		AccessKey:        "SCWAAAAAAAAAAAAAAAAA",
		SecretKey:        "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		DefaultProjectID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	}
	scalewayProfiles := map[string]ScalewayProfile{
		"default": scalewayDefaultProfile,
	}
	scalewayConfigFile := map[string]map[string]ScalewayProfile{
		"profiles": scalewayProfiles,
	}

	out, err := yaml.Marshal(scalewayConfigFile)
	if err != nil {
		return fmt.Errorf("error marshalling yaml file: %w", err)
	}
	err = os.WriteFile("./scw_config_test.yaml", out, 0644)
	if err != nil {
		return fmt.Errorf("error writing yaml file: %w", err)
	}
	return nil
}

func TestCreateValidScalewayProfile(t *testing.T) {
	tests := []struct {
		testDescription string
		loadConfigFile  bool
		environment     map[string]string
		expectedConfig  ScalewayProfile
	}{
		{
			testDescription: "Only environment set",
			loadConfigFile:  false,
			environment: map[string]string{
				"SCW_ACCESS_KEY":         "SCWBBBBBBBBBBBBBBBBB",
				"SCW_SECRET_KEY":         "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"SCW_DEFAULT_PROJECT_ID": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
			expectedConfig: ScalewayProfile{
				AccessKey:        "SCWBBBBBBBBBBBBBBBBB",
				SecretKey:        "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				DefaultProjectID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
		},
		{
			testDescription: "Only profile set",
			loadConfigFile:  true,
			environment: map[string]string{
				"SCW_PROFILE":     "default",
				"SCW_CONFIG_PATH": "./scw_config_test.yaml",
			},
			expectedConfig: ScalewayProfile{
				AccessKey:        "SCWAAAAAAAAAAAAAAAAA",
				SecretKey:        "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				DefaultProjectID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
		},
		{
			testDescription: "Environment should override profile's default project ID",
			loadConfigFile:  true,
			environment: map[string]string{
				"SCW_PROFILE":            "default",
				"SCW_CONFIG_PATH":        "./scw_config_test.yaml",
				"SCW_DEFAULT_PROJECT_ID": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
			expectedConfig: ScalewayProfile{
				AccessKey:        "SCWAAAAAAAAAAAAAAAAA",
				SecretKey:        "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				DefaultProjectID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
		},
		{
			testDescription: "Environment should override whole profile",
			loadConfigFile:  true,
			environment: map[string]string{
				"SCW_PROFILE":            "default",
				"SCW_CONFIG_PATH":        "./scw_config_test.yaml",
				"SCW_ACCESS_KEY":         "SCWBBBBBBBBBBBBBBBBB",
				"SCW_SECRET_KEY":         "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"SCW_DEFAULT_PROJECT_ID": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
			expectedConfig: ScalewayProfile{
				AccessKey:        "SCWBBBBBBBBBBBBBBBBB",
				SecretKey:        "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				DefaultProjectID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			},
		},
		{
			testDescription: "Empty profile for integration tests",
			loadConfigFile:  false,
			environment: map[string]string{
				"SCW_PROFILE": "REDACTED",
			},
			expectedConfig: ScalewayProfile{},
		},
	}

	for _, test := range tests {
		t.Run(test.testDescription, func(t *testing.T) {
			if test.loadConfigFile {
				err := createScalewayConfigFile()
				if err != nil {
					t.Fatal(err)
				}
			}

			for k, v := range test.environment {
				err := os.Setenv(k, v)
				if err != nil {
					t.Error(err)
				}
			}

			defer t.Cleanup(func() {
				// Delete config file
				err := os.Remove("./scw_config_test.yaml")
				if err != nil {
					if !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("error deleting yaml file: %v", err)
					}
				}
				// Unset environment variables
				envToUnset := []string{"SCW_PROFILE", "SCW_CONFIG_PATH", "SCW_ACCESS_KEY", "SCW_SECRET_KEY", "SCW_DEFAULT_PROJECT_ID"}
				for _, key := range envToUnset {
					err := os.Unsetenv(key)
					if err != nil {
						t.Fatalf("error unsetting environment: %v", err)
					}
				}
			})

			actualProfile, err := CreateValidScalewayProfile()
			if err != nil {
				t.Error(err)
			}
			actualConfig := ScalewayProfile{
				AccessKey:        fi.ValueOf(actualProfile.AccessKey),
				SecretKey:        fi.ValueOf(actualProfile.SecretKey),
				DefaultProjectID: fi.ValueOf(actualProfile.DefaultProjectID),
			}

			if actualConfig != test.expectedConfig {
				t.Errorf("config differs, expected %+v, got %+v", test.expectedConfig, actualConfig)
			}
		})
	}
}
