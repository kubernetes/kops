/*
Copyright 2017 The Kubernetes Authors.

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

package vspheretasks

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"io/ioutil"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AttachISO represents the cloud-init ISO file attached to a VMware VM
//go:generate fitask -type=AttachISO
type AttachISO struct {
	Name            *string
	VM              *VirtualMachine
	IG              *kops.InstanceGroup
	BootstrapScript *model.BootstrapScript
}

var _ fi.HasName = &AttachISO{}
var _ fi.HasDependencies = &AttachISO{}

func (o *AttachISO) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	vmCreateTask := tasks["VirtualMachine/"+*o.VM.Name]
	if vmCreateTask == nil {
		glog.Fatalf("Unable to find create VM task %s dependency for AttachISO %s", *o.VM.Name, *o.Name)
	}
	deps = append(deps, vmCreateTask)
	return deps
}

// GetName returns the Name of the object, implementing fi.HasName
func (o *AttachISO) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *AttachISO) SetName(name string) {
	o.Name = &name
}

func (e *AttachISO) Run(c *fi.Context) error {
	glog.Info("AttachISO.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *AttachISO) Find(c *fi.Context) (*AttachISO, error) {
	glog.Info("AttachISO.Find invoked!")
	return nil, nil
}

func (_ *AttachISO) CheckChanges(a, e, changes *AttachISO) error {
	glog.Info("AttachISO.CheckChanges invoked!")
	return nil
}

func (_ *AttachISO) RenderVSphere(t *vsphere.VSphereAPITarget, a, e, changes *AttachISO) error {
	startupScript, err := changes.BootstrapScript.ResourceNodeUp(changes.IG)
	startupStr, err := startupScript.AsString()
	if err != nil {
		return fmt.Errorf("error rendering startup script: %v", err)
	}
	dir, err := ioutil.TempDir("", *changes.VM.Name)
	defer os.RemoveAll(dir)

	isoFile := createISO(changes, startupStr, dir)
	err = t.Cloud.UploadAndAttachISO(changes.VM.Name, isoFile)
	if err != nil {
		return err
	}

	return nil
}

func createUserData(startupStr string, dir string) {
	// Update the startup script to add the extra spaces for
	// indentation when copied to the user-data file.
	strArray := strings.Split(startupStr, "\n")
	for i, str := range strArray {
		if len(str) > 0 {
			strArray[i] = "       " + str
		}
	}
	startupStr = strings.Join(strArray, "\n")

	data := strings.Replace(userDataTemplate, "$SCRIPT", startupStr, -1)
	userDataFile := filepath.Join(dir, "user-data")
	glog.V(4).Infof("User data file content: %s", data)

	if err := ioutil.WriteFile(userDataFile, []byte(data), 0644); err != nil {
		glog.Fatalf("Unable to write user-data into file %s", userDataFile)
	}
	return
}

func createMetaData(dir string, vmName string) {
	data := strings.Replace(metaDataTemplate, "$INSTANCE_ID", uuid.NewUUID().String(), -1)
	data = strings.Replace(data, "$LOCAL_HOST_NAME", vmName, -1)

	glog.V(4).Infof("Meta data file content: %s", string(data))

	metaDataFile := filepath.Join(dir, "meta-data")
	if err := ioutil.WriteFile(metaDataFile, []byte(data), 0644); err != nil {
		glog.Fatalf("Unable to write meta-data into file %s", metaDataFile)
	}
	return
}

func createISO(changes *AttachISO, startupStr string, dir string) string {
	createUserData(startupStr, dir)
	createMetaData(dir, *changes.VM.Name)

	isoFile := filepath.Join(dir, *changes.VM.Name+".iso")
	var commandName string

	switch os := runtime.GOOS; os {
	case "darwin":
		commandName = "mkisofs"
	case "linux":
		commandName = "genisoimage"

	default:
		glog.Fatalf("Cannot generate ISO file %s. Unsupported operation system (%s)!!!", isoFile, os)
	}
	cmd := exec.Command(commandName, "-o", isoFile, "-volid", "cidata", "-joliet", "-rock", dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		glog.Fatalf("Error %s occurred while executing command %+v", err, cmd)
	}
	glog.V(4).Infof("%s std output : %s\n", commandName, out.String())
	glog.V(4).Infof("%s std error : %s\n", commandName, stderr.String())
	return isoFile
}
