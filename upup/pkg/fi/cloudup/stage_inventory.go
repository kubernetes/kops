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

	"net/url"

	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
)

const (
	dockerExec     = "docker"
	AssetBinary    = "binary"
	AssetContainer = "container"
)

type AssetTransferer interface {
	Transfer(asset interface{}, t string) error
}

type FileAssetTransferer struct {
	fileRepo string
}

type ContainerAssetTransferer struct {
	containerRepo string
}

type ContainerFileAssetTransferer struct {
	containerRepo string
}

type StageInventory struct {
	assetTransferers map[string]AssetTransferer
	inventory        api.InventorySpec
}

func NewStageInventory(fileRepo string, stageFiles bool, containerRepo string, stageContainers bool, assets *api.Inventory) *StageInventory {

	assetTransferers := make(map[string]AssetTransferer)

	if stageFiles {
		assetTransferers[AssetBinary] = FileAssetTransferer{
			fileRepo: fileRepo,
		}
	}

	if stageContainers {
		assetTransferers[AssetContainer] = ContainerAssetTransferer{
			containerRepo: containerRepo,
		}
	}

	return &StageInventory{
		assetTransferers: assetTransferers,
		inventory:        assets.Spec,
	}
}

func (i *StageInventory) Run() error {

	for _, asset := range i.inventory.ContainerAssets {
		err := i.processAsset(*asset, AssetContainer)
		if err != nil {
			return fmt.Errorf("Error StageInventory.Run - Type:%s Data: %+v, err: %v", AssetContainer, asset, err)
		}
	}

	for _, asset := range i.inventory.ExecutableFileAsset {
		err := i.processAsset(*asset, AssetBinary)
		if err != nil {
			return fmt.Errorf("Error StageInventory.Run - Type:%s Data: %+v, err %v", AssetBinary, asset, err)
		}
	}

	// FIXME channel

	return nil

}

func (i *StageInventory) processAsset(asset interface{}, t string) error {

	assetTransferer := i.assetTransferers[t]

	if assetTransferer == nil {
		glog.Infof("skipping transfer: %#v - asset: %#v\n", assetTransferer, asset)
		return nil
	}
	glog.Infof("processing transfer: %#v - asset: %#v\n", assetTransferer, asset)
	err := assetTransferer.Transfer(asset, t)
	if err != nil {
		return fmt.Errorf("Error Transfering Asset - Type:%s Data:%+v - %v", t, asset, err)
	}

	return nil
}

func (f FileAssetTransferer) TransferSha(location string, t string) error {
	glog.Infof("File asset transfer: %s - %s\n", t, location)

	glog.Infoln("FileAssetTransferer.Transfer: reading data...")
	data, err := vfs.Context.ReadFile(location)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer  unable to read path %q: %v", location, err)
	}

	fileURL, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer  unable to read path %q: %v", location, err)
	}

	s3Path := f.fileRepo + fileURL.Path
	glog.Infof("FileAssetTransferer.Transfer: s3Path: %s\n", s3Path)
	destinationRegistry, err := vfs.Context.BuildVfsPath(s3Path)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer parsing registry path %q: %v", f.fileRepo, err)
	}

	glog.Infoln("FileAssetTransferer.Transfer: writing data...")
	err = destinationRegistry.WriteFile(data)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer destination path %q: %v", f.fileRepo, err)
	}

	return nil
}

func (f FileAssetTransferer) Transfer(i interface{}, t string) error {

	asset := i.(api.ExecutableFileAsset)
	glog.Infof("File asset transfer: %s - %s\n", t, asset.Location)

	glog.Infoln("FileAssetTransferer.Transfer: reading data...")
	data, err := vfs.Context.ReadFile(asset.Location)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer  unable to read path %q: %v", asset.Location, err)
	}

	fileURL, err := url.Parse(asset.Location)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer  unable to read path %q: %v", asset.Location, err)
	}

	s3Path := f.fileRepo + fileURL.Path
	glog.Infof("FileAssetTransferer.Transfer: s3Path: %s\n", s3Path)
	destinationRegistry, err := vfs.Context.BuildVfsPath(s3Path)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer parsing registry path %q: %v", f.fileRepo, err)
	}

	glog.Infoln("FileAssetTransferer.Transfer: writing data...")
	err = destinationRegistry.WriteFile(data)
	if err != nil {
		return fmt.Errorf("Error FileAssetTransferer.Transfer destination path %q: %v", f.fileRepo, err)
	}

	if asset.SHA != "" {
		f.TransferSha(asset.SHA, "sha")
	}

	return nil
}

func (c ContainerAssetTransferer) Transfer(i interface{}, t string) error {
	asset := i.(api.ContainerAsset)
	glog.Infof("ContainerAssetTransferer.Transfer: %s - %v\n", t, asset)

	if asset.Location != "" {
		return c.TransferFile(asset, t)
	}

	// Download image
	location := asset.String
	glog.Infof("Downloading container image %s\n", location)
	args := []string{"pull", location}
	err := performExec(dockerExec, args)
	if err != nil {
		return err
	}

	var name string
	if strings.Contains(asset.Name, "/") {
		split := strings.Split(asset.Name, "/")
		name = split[len(split)-1]
	} else {
		name = asset.Name
	}

	tagName := fmt.Sprintf("%s/%s:%s", c.containerRepo, name, asset.Tag)
	glog.Infof("Tagging local image tagName[-]%s\n", tagName)
	err = tagAndPushToDocker(tagName, location)
	if err != nil {
		return fmt.Errorf("Error pushing docker image with tagName-'%s' baseDockerImageId-'%s': %v", tagName, location, err)
	}

	err = cleanUpDockerImages(tagName, location)
	if err != nil {
		return fmt.Errorf("Error cleanup images with tagName-'%s' baseDockerImageId-'%s': %v", tagName, location, err)
	}

	return nil
}

func (c ContainerAssetTransferer) TransferFile(i interface{}, t string) error {

	asset := i.(api.ContainerAsset)
	// TODO update logging to match kops, no use of method and struct member names
	glog.Infof("ContainerFileAssetTransferer.Transfer starting: %s - %v\n", t, asset)

	uuid, err := NewUUID()
	if err != nil {
		return fmt.Errorf("Error getting UUID for file '%s': %v", asset.Location, err)
	}

	pathParts := strings.Split(asset.Location, "/")
	// TODO get system tmp folder and make a temp diretory
	localFile := fmt.Sprintf("/tmp/%s-%s", uuid, pathParts[len(pathParts)-1])

	glog.Infof("Local file: %s\n", localFile)

	dirMode := os.FileMode(0755)
	err = downloadFile(asset.Location, localFile, dirMode)
	if err != nil {
		return fmt.Errorf("Error downloading file '%s': %v", asset.Location, err)
	}
	// File Cleanup
	defer func() {
		err = os.Remove(localFile)
		if err != nil {
			glog.Warningf("Error Removing file-'%s': %v", localFile, err)
		}

	}()

	// Load the image into docker
	args := []string{"docker", "load", "-i", localFile}
	human := strings.Join(args, " ")

	glog.Infof("ContainerFileAssetTransferer.Transfer  Running command %s\n", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error loading docker image with '%s': %v: %s", human, err, string(output))
	}

	dockerImageId := extractImageIdFromOutput(string(output))
	glog.Infof("ContainerFileAssetTransferer.Transfer Loaded image id: %s\n", dockerImageId)

	tagName := fmt.Sprintf("%s/%s", c.containerRepo, dockerImageId)
	err = tagAndPushToDocker(tagName, dockerImageId)
	if err != nil {
		return fmt.Errorf("Error pushing docker image with tagName-'%s' baseDockerImageId-'%s': %v", tagName, dockerImageId, err)
	}

	err = cleanUpDockerImages(tagName, dockerImageId)
	if err != nil {
		return fmt.Errorf("Error cleanup images with tagName-'%s' baseDockerImageId-'%s': %v", tagName, dockerImageId, err)
	}

	return nil
}

func extractImageIdFromOutput(output string) string {
	// Assumes oputput format is 'Loaded image: <imageId>'
	outputValues := strings.Split(string(output), "Loaded image: ")
	return strings.Trim(outputValues[1], "\n")
}

func downloadFile(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file for download %q: %v", destPath, err)
	}
	defer output.Close()

	glog.Infof("Downloading %q", url)

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error doing HTTP fetch of %q: %v", url, err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error downloading HTTP content from %q: %v", url, err)
	}
	return nil
}

// Stolen from: https://play.golang.org/p/4FkNSiUDMg
// newUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
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

func tagAndPushToDocker(tagName, imageId string) error {

	glog.Infof("Tagging local image tagName[-]%s\n", tagName)
	args := []string{"tag", imageId, tagName}
	err := performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - tagging tagName '%s' - imageId '%s' : %v", tagName, imageId, err)
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

func cleanUpDockerImages(pushedImageId, baseImageId string) error {

	args := []string{"rmi", "-f", pushedImageId}
	glog.Infof("Removing pushed container image %s\n", strings.Join(args, " "))
	err := performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - removing pushed container image '%s': %v", pushedImageId, err)
	}

	args = []string{"rmi", "-f", baseImageId}
	glog.Infof("Removing base container image %s\n", strings.Join(args, " "))
	err = performExec(dockerExec, args)
	if err != nil {
		return fmt.Errorf("Docker Error - removing base container image '%s': %v", baseImageId, err)
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
