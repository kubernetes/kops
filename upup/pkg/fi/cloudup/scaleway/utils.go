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

package scaleway

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// isHTTPCodeError returns true if err is an http error with code statusCode
func isHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if errors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}
	return false
}

// is404Error returns true if err is an HTTP 404 error
func is404Error(err error) bool {
	notFoundError := &scw.ResourceNotFoundError{}
	return isHTTPCodeError(err, http.StatusNotFound) || errors.As(err, &notFoundError)
}

func ParseZoneFromClusterSpec(clusterSpec kops.ClusterSpec) (scw.Zone, error) {
	zone := ""
	for _, subnet := range clusterSpec.Networking.Subnets {
		if zone == "" {
			zone = subnet.Zone
		} else if zone != subnet.Zone {
			return "", fmt.Errorf("scaleway currently only supports clusters in the same zone")
		}
	}
	return scw.Zone(zone), nil
}

func ParseRegionFromZone(zone scw.Zone) (region scw.Region, err error) {
	region, err = scw.ParseRegion(strings.TrimRight(string(zone), "-123"))
	if err != nil {
		return "", fmt.Errorf("could not determine region from zone %s: %w", zone, err)
	}
	return region, nil
}

func ClusterNameFromTags(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, TagClusterName) {
			return strings.TrimPrefix(tag, TagClusterName+"=")
		}
	}
	return ""
}

func InstanceGroupNameFromTags(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, TagInstanceGroup) {
			return strings.TrimPrefix(tag, TagInstanceGroup+"=")
		}
	}
	return ""
}

func InstanceRoleFromTags(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, TagNameRolePrefix) {
			return strings.TrimPrefix(tag, TagNameRolePrefix+"=")
		}
	}
	return ""
}

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
		return nil, fmt.Errorf(errMsg)
	}

	return &profile, nil
}

func GetIPAMPublicIP(ipamAPI *ipam.API, serverID string, zone scw.Zone) (string, error) {
	if ipamAPI == nil {
		profile, err := CreateValidScalewayProfile()
		if err != nil {
			return "", err
		}
		scwClient, err := scw.NewClient(
			scw.WithProfile(profile),
			scw.WithUserAgent(KopsUserAgentPrefix+kopsv.Version),
		)
		if err != nil {
			return "", fmt.Errorf("creating client for Scaleway IPAM checking: %w", err)
		}
		ipamAPI = ipam.NewAPI(scwClient)
	}

	region, err := zone.Region()
	if err != nil {
		return "", fmt.Errorf("converting zone %s to region: %w", zone, err)
	}

	ips, err := ipamAPI.ListIPs(&ipam.ListIPsRequest{
		Region:     region,
		IsIPv6:     fi.PtrTo(false),
		ResourceID: &serverID,
		Zonal:      fi.PtrTo(zone.String()), // not useful according to tests made on this route with the CLI
	}, scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("listing IPs for server %s: %w", serverID, err)
	}

	if len(ips.IPs) < 1 {
		return "", fmt.Errorf("could not find IP for server %s", serverID)
	}
	if len(ips.IPs) > 1 {
		klog.V(10).Infof("Found more than 1 IP for server %s, using %s", serverID, ips.IPs[0].Address.IP.String())
	}

	return ips.IPs[0].Address.IP.String(), nil
}
