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

package cloudup

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"k8s.io/kops/util/pkg/vfs"

	"github.com/golang/glog"
)

const (
	dockerExec = "docker"
)

// TransferExecutableFileAssets transfers file contents from source to destination
func TransferExecutableFileAssets(source string, destination string) error {

	glog.Infof("TransferExecutableFileAssets: %s ---> %s\n", source, destination)

	glog.Infoln("TransferExecutableFileAssets: reading from source...")
	data, err := vfs.Context.ReadFile(source)
	if err != nil {
		return fmt.Errorf("Error unable to read path %q: %v", source, err)
	}

	filePath := strings.Split(source, "/")
	s3Path := fmt.Sprintf("%s/%s", destination, filePath[len(filePath)-1])
	glog.Infof("TransferExecutableFileAssets: s3Path: %s\n", s3Path)
	destinationRegistry, err := vfs.Context.BuildVfsPath(s3Path)
	if err != nil {
		return fmt.Errorf("Error parsing registry path %q: %v", destination, err)
	}

	glog.Infoln("TransferExecutableFileAssets: writing data...")
	err = destinationRegistry.WriteFile(data)
	if err != nil {
		return fmt.Errorf("Error destination path %q: %v", destination, err)
	}

	return nil
}

// TransferContainerAssets transfers container image from source docker registry to destination docker registry.
// This method assumes the docker client is installed.
func TransferContainerAssets(source string, destination string) error {

	glog.Infof("TransferContainerAssets: %s --> %s\n", source, destination)

	// Download image
	glog.Infof("TransferContainerAssets: Downloading container image %s\n", source)
	args := []string{"pull", source}
	err := performExec(dockerExec, args)
	if err != nil {
		return err
	}

	//Tag image with new Repo
	dockerImageVersion := strings.Split(source, ":")
	originalImageName := dockerImageVersion[0]
	imageVersion := dockerImageVersion[1]
	imagePath := strings.Split(originalImageName, "/")
	tagName := fmt.Sprintf("%s/%s:%s", destination, imagePath[len(imagePath)-1], imageVersion)
	glog.Infof("TransferContainerAssets: Tagging local image tagName[-]%s\n", tagName)
	err = tagAndPushToDocker(tagName, source)
	if err != nil {
		return fmt.Errorf("Error pushing docker image with tagName-'%s' baseDockerImageId-'%s': %v", tagName, source, err)
	}

	err = cleanUpDockerImages(tagName, source)
	if err != nil {
		return fmt.Errorf("Error cleanup images with tagName-'%s' baseDockerImageId-'%s': %v", tagName, source, err)
	}

	return nil
}

// CompressedFileAssets transfers compressed image from source to destination docker registry
// This method assumes the docker client is installed.
func TransferCompressedFileAssets(source string, destination string) error {

	glog.Infof("TransferCompressedFileAssets starting: %s --> %s\n", source, destination)

	uuid, err := newUUID()
	if err != nil {
		return fmt.Errorf("TransferCompressedFileAssets: Error getting UUID for file '%s': %v", source, err)
	}

	pathParts := strings.Split(source, "/")
	localFile := fmt.Sprintf(os.TempDir(), uuid, pathParts[len(pathParts)-1])

	glog.Infof("TransferCompressedFileAssets: Local file: %s\n", localFile)

	dirMode := os.FileMode(0755)
	err = downloadFile(source, localFile, dirMode)
	if err != nil {
		return fmt.Errorf("Error downloading file '%s': %v", source, err)
	}
	// File Cleanup
	defer func() {
		err = os.RemoveAll(localFile)
		if err != nil {
			glog.Warningf("Error Removing temporary directory %q: %v", localFile, err)
		}

	}()

	// Load the image into docker
	args := []string{"docker", "load", "-i", localFile}
	human := strings.Join(args, " ")

	glog.Infof("TransferCompressedFileAssets:  Running command %s\n", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error loading docker image with '%s': %v: %s", human, err, string(output))
	}

	dockerImageID := extractImageIDFromOutput(string(output))
	glog.Infof("TransferCompressedFileAssets: Loaded image id: %s\n", dockerImageID)

	tagName := fmt.Sprintf("%s/%s", destination, dockerImageID)
	err = tagAndPushToDocker(tagName, dockerImageID)
	if err != nil {
		return fmt.Errorf("Error pushing docker image with tagName-'%s' baseDockerImageID-'%s': %v", tagName, dockerImageID, err)
	}

	err = cleanUpDockerImages(tagName, dockerImageID)
	if err != nil {
		return fmt.Errorf("Error cleanup images with tagName-'%s' baseDockerImageID-'%s': %v", tagName, dockerImageID, err)
	}

	return nil
}

func extractImageIDFromOutput(output string) string {
	// Assumes oputput format is 'Loaded image: <imageId>'
	outputValues := strings.Split(string(output), "Loaded image: ")
	return strings.Trim(outputValues[1], "\n")
}

func downloadFile(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("Error creating directories for destination file %q: %v", destPath, err)
	}

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Error creating file for download %q: %v", destPath, err)
	}
	defer output.Close()

	glog.Infof("Downloading %q", url)

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error doing HTTP fetch of %q: %v", url, err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("Error downloading HTTP content from %q: %v", url, err)
	}
	return nil
}

func newUUID() (string, error) {
	// Stolen from: https://play.golang.org/p/4FkNSiUDMg
	// newUUID generates a random UUID according to RFC 4122
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func tagAndPushToDocker(tagName, imageID string) error {

	glog.Infof("Tagging local image tagName[-]%s\n", tagName)
	args := []string{"tag", imageID, tagName}
	err := performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - tagging tagName '%s' - imageID '%s' : %v", tagName, imageID, err)
	}

	// Push image to new Repo
	glog.Infof("Pushing image tagName[-]%s\n", tagName)
	args = []string{"push", tagName}
	err = performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - pushing tagName '%s': %v", tagName, err)
	}

	return nil
}

func cleanUpDockerImages(pushedImageID, baseImageID string) error {

	args := []string{"rmi", "-f", pushedImageID}
	glog.Infof("Removing pushed container image %s\n", strings.Join(args, " "))
	err := performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - removing pushed container image '%s': %v", pushedImageID, err)
	}

	args = []string{"rmi", "-f", baseImageID}
	glog.Infof("Removing base container image %s\n", strings.Join(args, " "))
	err = performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - removing base container image '%s': %v", baseImageID, err)
	}

	return nil
}

func performExec(cmdStr string, args []string) error {

	binary, err := exec.LookPath(cmdStr)
	if err != nil {
		return fmt.Errorf("Error finding executable file: %s - %v", cmdStr, err)
	}

	cmd := exec.Command(binary, args...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v -- %s\n", err, errOut.String())
		return err
	}
	return nil
}
