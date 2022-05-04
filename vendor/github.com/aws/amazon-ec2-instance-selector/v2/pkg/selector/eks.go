// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package selector

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

const (
	eksAMIRepoURL               = "https://github.com/awslabs/amazon-eks-ami"
	eksFallbackLatestAMIVersion = "v20210125"
	eksInstanceTypesFile        = "eni-max-pods.txt"
)

// EKS is a Service type for a custom service filter transform
type EKS struct {
	AMIRepoURL string
}

// Filters implements the Service interface contract for EKS
func (e *EKS) Filters(version string) (Filters, error) {
	if e.AMIRepoURL == "" {
		e.AMIRepoURL = eksAMIRepoURL
	}
	var filters Filters

	if version == "" {
		var err error
		version, err = e.getLatestAMIVersion()
		if err != nil {
			log.Printf("There was a problem fetching the latest EKS AMI version, using hardcoded fallback version %s\n", eksFallbackLatestAMIVersion)
			version = eksFallbackLatestAMIVersion
		}
	}
	supportedInstanceTypes, err := e.getSupportedInstanceTypes(version)
	if err != nil {
		log.Printf("Unable to retrieve EKS supported instance types for version %s: %v", version, err)
		return filters, err
	}
	filters.InstanceTypes = &supportedInstanceTypes
	filters.VirtualizationType = aws.String("hvm")
	return filters, nil
}

func (e *EKS) getSupportedInstanceTypes(version string) ([]string, error) {
	supportedInstanceTypes := []string{}
	resp, err := http.Get(fmt.Sprintf("%s/archive/%s.zip", e.AMIRepoURL, version))
	if err != nil {
		return supportedInstanceTypes, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return supportedInstanceTypes, fmt.Errorf("Unable to retrieve EKS supported instance types, got non-200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return supportedInstanceTypes, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return supportedInstanceTypes, err
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		filePathParts := strings.Split(zipFile.Name, "/")
		fileName := filePathParts[len(filePathParts)-1]
		if fileName == eksInstanceTypesFile {
			unzippedFileBytes, err := readZipFile(zipFile)
			if err != nil {
				log.Println(err)
				continue
			}
			supportedInstanceTypesFileBody := string(unzippedFileBytes)
			for _, line := range strings.Split(strings.Replace(supportedInstanceTypesFileBody, "\r\n", "\n", -1), "\n") {
				if !strings.HasPrefix(line, "#") {
					instanceType := strings.Split(line, " ")[0]
					supportedInstanceTypes = append(supportedInstanceTypes, instanceType)
				}
			}
		}
	}
	return supportedInstanceTypes, nil
}

func (e EKS) getLatestAMIVersion() (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// Get latest version
	resp, err := client.Get(fmt.Sprintf("%s/releases/latest", e.AMIRepoURL))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("Can't retrieve latest release from github because redirect was not sent")
	}
	versionRedirect := resp.Header.Get("location")
	pathParts := strings.Split(versionRedirect, "/")
	return pathParts[len(pathParts)-1], nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
