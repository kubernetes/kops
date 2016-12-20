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

package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/nodeup/pkg/model"
	"os"
	"path"
	"strings"
)

// FindDistribution identifies the distribution on which we are running
// We will likely remove this when everything is containerized
func FindDistribution(rootfs string) (model.Distribution, error) {
	// Ubuntu has /etc/lsb-release (and /etc/debian_version)
	lsbRelease, err := ioutil.ReadFile(path.Join(rootfs, "etc/lsb-release"))
	if err == nil {
		for _, line := range strings.Split(string(lsbRelease), "\n") {
			line = strings.TrimSpace(line)
			if line == "DISTRIB_CODENAME=xenial" {
				return model.DistributionXenial, nil
			}
		}
		glog.Warningf("unhandled lsb-release info %q", string(lsbRelease))
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/lsb-release: %v", err)
	}

	// Debian has /etc/debian_version
	debianVersionBytes, err := ioutil.ReadFile(path.Join(rootfs, "etc/debian_version"))
	if err == nil {
		debianVersion := strings.TrimSpace(string(debianVersionBytes))
		if strings.HasPrefix(debianVersion, "8.") {
			return model.DistributionJessie, nil
		} else {
			return "", fmt.Errorf("unhandled debian version %q", debianVersion)
		}
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/debian_version: %v", err)
	}

	// Redhat has /etc/redhat-release
	// Centos has /etc/centos-release
	redhatRelease, err := ioutil.ReadFile(path.Join(rootfs, "etc/redhat-release"))
	if err == nil {
		for _, line := range strings.Split(string(redhatRelease), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Red Hat Enterprise Linux Server release 7.") {
				return model.DistributionRhel7, nil
			}
			if strings.HasPrefix(line, "CentOS Linux release 7.") {
				return model.DistributionCentos7, nil
			}
		}
		glog.Warningf("unhandled redhat-release info %q", string(lsbRelease))
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/redhat-release: %v", err)
	}

	return "", fmt.Errorf("cannot identify distro")
}

//// FindCloudTags infers tags from the cloud environment
//func FindCloudTags(rootfs string) ([]string, error) {
//	productVersionBytes, err := ioutil.ReadFile(path.Join(rootfs, "sys/class/dmi/id/product_version"))
//	if err == nil {
//		productVersion := strings.TrimSpace(string(productVersionBytes))
//		switch productVersion {
//		case "amazon":
//			return findCloudTagsAWS(rootfs)
//		default:
//			glog.V(2).Infof("Unknown /sys/class/dmi/id/product_version %q", productVersion)
//		}
//	} else if !os.IsNotExist(err) {
//		glog.Infof("error reading /sys/class/dmi/id/product_version: %v", err)
//	}
//	return nil, fmt.Errorf("cannot identify cloud")
//}
//
//type awsIAMInfo struct {
//	Code               string
//	LastUpdated        string
//	InstanceProfileArn string
//	InstanceProfileId  string
//}
//
//// findAWSCloudTags infers cloud tags once we have determined we are on AWS
//func findCloudTagsAWS(rootfs string) ([]string, error) {
//	tags := []string{"_aws"}
//
//	// We can't get the tags, annoyingly
//
//	iamInfoBytes, err := vfs.Context.ReadFile("http://169.254.169.254/2016-04-19/meta-data/iam/info")
//	if err != nil {
//		return nil, fmt.Errorf("error querying for iam info: %v", err)
//	}
//
//	iamInfo := &awsIAMInfo{}
//	if err := json.Unmarshal(iamInfoBytes, iamInfo); err != nil {
//		glog.Infof("Invalid IAM info: %q", string(iamInfoBytes))
//		return nil, fmt.Errorf("error decoding iam info: %v", err)
//	}
//
//	arn := iamInfo.InstanceProfileArn
//	if strings.HasSuffix(arn, "-masters") {
//		tags = append(tags, "_master")
//	} else if strings.HasSuffix(arn, "-nodes") {
//		tags = append(tags, "_master")
//	} else {
//		return nil, fmt.Errorf("unexpected IAM role name %q", arn)
//	}
//
//	return tags, nil
//}
//
//
