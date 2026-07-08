/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaymetadata

import (
	"fmt"
	"os"

	"github.com/scaleway/scaleway-sdk-go/scw"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kops/upup/pkg/fi"
)

func getScalewayProfile() (*scw.Profile, error) {
	scwProfileName := os.Getenv("SCW_PROFILE")
	if scwProfileName == "" {
		return nil, nil
	}
	config, err := scw.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading Scaleway config file: %w", err)
	}
	profile, ok := config.Profiles[scwProfileName]
	if !ok {
		return nil, fmt.Errorf("could not find Scaleway profile %q", scwProfileName)
	}
	return profile, nil
}

func checkCredentials(accessKey, secretKey, projectID string) []error {
	errList := []error(nil)
	if accessKey == "" {
		errList = append(errList, fmt.Errorf("SCW_ACCESS_KEY has to be set"))
	}
	if secretKey == "" {
		errList = append(errList, fmt.Errorf("SCW_SECRET_KEY has to be set"))
	}
	if projectID == "" {
		errList = append(errList, fmt.Errorf("SCW_DEFAULT_PROJECT_ID has to be set"))
	}
	return errList
}

func CreateValidScalewayProfile() (*scw.Profile, error) {
	var profile scw.Profile

	// If the profile is REDACTED, we're running integration tests so no need to return any credentials
	if profileName := os.Getenv("SCW_PROFILE"); profileName == "REDACTED" {
		return &profile, nil
	}

	// We load the credentials form the profile first
	profileFromScwConfig, err := getScalewayProfile()
	if err != nil {
		return nil, err
	}
	// We load the credentials from the environment second
	var profileFromEnv scw.Profile
	if accessKey := os.Getenv("SCW_ACCESS_KEY"); accessKey != "" {
		profileFromEnv.AccessKey = &accessKey
	}
	if secretKey := os.Getenv("SCW_SECRET_KEY"); secretKey != "" {
		profileFromEnv.SecretKey = &secretKey
	}
	if projectID := os.Getenv("SCW_DEFAULT_PROJECT_ID"); projectID != "" {
		profileFromEnv.DefaultProjectID = &projectID
	}

	// We merge the profiles: the environment will override the values from the profile
	if profileFromScwConfig == nil {
		profile = profileFromEnv
	} else {
		profile = *scw.MergeProfiles(profileFromScwConfig, &profileFromEnv)
	}

	// We check that the profile has an access key, a secret key and a default project ID
	if errList := checkCredentials(fi.ValueOf(profile.AccessKey), fi.ValueOf(profile.SecretKey), fi.ValueOf(profile.DefaultProjectID)); errList != nil {
		errMsg := k8serrors.NewAggregate(errList).Error()
		if scwProfileName := os.Getenv("SCW_PROFILE"); scwProfileName != "" {
			errMsg += fmt.Sprintf(" in profile %q", scwProfileName)
		} else {
			errMsg += " in a Scaleway profile or as an environment variable"
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	return &profile, nil
}
