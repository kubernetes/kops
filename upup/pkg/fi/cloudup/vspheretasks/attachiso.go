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

// attachiso houses the task that creates cloud-init ISO file, uploads and attaches it to a VM on vSphere cloud.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pborman/uuid"
	"k8s.io/klog"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// AttachISO represents the cloud-init ISO file attached to a VM on vSphere cloud.
//go:generate fitask -type=AttachISO
type AttachISO struct {
	Name            *string
	VM              *VirtualMachine
	IG              *kops.InstanceGroup
	BootstrapScript *model.BootstrapScript
	Cluster         *kops.Cluster
}

var _ fi.HasName = &AttachISO{}
var _ fi.HasDependencies = &AttachISO{}

// GetDependencies returns map of tasks on which this task depends.
func (o *AttachISO) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	vmCreateTask := tasks["VirtualMachine/"+*o.VM.Name]
	if vmCreateTask == nil {
		klog.Fatalf("Unable to find create VM task %s dependency for AttachISO %s", *o.VM.Name, *o.Name)
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

// Run invokes DefaultDeltaRunMethod for this task.
func (e *AttachISO) Run(c *fi.Context) error {
	klog.Info("AttachISO.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

// Find is a no-op for this task.
func (e *AttachISO) Find(c *fi.Context) (*AttachISO, error) {
	klog.Info("AttachISO.Find invoked!")
	return nil, nil
}

// CheckChanges is a no-op for this task.
func (_ *AttachISO) CheckChanges(a, e, changes *AttachISO) error {
	klog.Info("AttachISO.CheckChanges invoked!")
	return nil
}

// RenderVSphere executes the actual task logic, for vSphere cloud.
func (_ *AttachISO) RenderVSphere(t *vsphere.VSphereAPITarget, a, e, changes *AttachISO) error {
	startupScript, err := e.BootstrapScript.ResourceNodeUp(e.IG, e.Cluster)
	if err != nil {
		return fmt.Errorf("error on resource nodeup: %v", err)
	}
	startupStr, err := startupScript.AsString()
	if err != nil {
		return fmt.Errorf("error rendering startup script: %v", err)
	}
	dir, err := ioutil.TempDir("", *e.VM.Name)
	if err != nil {
		return fmt.Errorf("error creating tempdir: %v", err)
	}

	defer os.RemoveAll(dir)

	// Need this in cloud config file for vSphere CloudProvider
	vmUUID, err := t.Cloud.FindVMUUID(e.VM.Name)
	if err != nil {
		return err
	}

	isoFile, err := createISO(e, startupStr, dir, t.Cloud.CoreDNSServer, vmUUID)
	if err != nil {
		klog.Errorf("Failed to createISO for vspheretasks, err: %v", err)
		return err
	}

	err = t.Cloud.UploadAndAttachISO(e.VM.Name, isoFile)
	if err != nil {
		return err
	}

	return nil
}

func createUserData(e *AttachISO, startupStr string, dir string, dnsServer string, vmUUID string) error {

	// Populate nodeup initialization script.

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

	// Populate script to update nameserver for the VM.
	dnsURL, err := url.Parse(dnsServer)
	if err != nil {
		return err
	}
	dnsHost, _, err := net.SplitHostPort(dnsURL.Host)
	if err != nil {
		return err
	}
	var lines []string
	lines = append(lines, "       echo \"nameserver "+dnsHost+"\" >> /etc/resolvconf/resolv.conf.d/head")
	lines = append(lines, "       resolvconf -u")
	dnsUpdateStr := strings.Join(lines, "\n")
	data = strings.Replace(data, "$DNS_SCRIPT", dnsUpdateStr, -1)

	// Populate VM UUID information.
	vmUUIDStr := "       " + vmUUID + "\n"
	data = strings.Replace(data, "$VM_UUID", vmUUIDStr, -1)

	// Populate volume metadata.
	data, err = createVolumeScript(e, data)
	if err != nil {
		return err
	}

	userDataFile := filepath.Join(dir, "user-data")
	klog.V(4).Infof("User data file content: %s", data)

	if err = ioutil.WriteFile(userDataFile, []byte(data), 0644); err != nil {
		klog.Errorf("Unable to write user-data into file %s", userDataFile)
		return err
	}

	return nil
}

func createVolumeScript(e *AttachISO, data string) (string, error) {
	if e.IG.Spec.Role != kops.InstanceGroupRoleMaster {
		return strings.Replace(data, "$VOLUME_SCRIPT", "       No volume metadata needed for "+string(e.IG.Spec.Role)+".", -1), nil
	}

	volsString, err := getVolMetadata(e)

	if err != nil {
		return "", err
	}

	return strings.Replace(data, "$VOLUME_SCRIPT", "       "+volsString, -1), nil
}

func getVolMetadata(e *AttachISO) (string, error) {
	var volsMetadata []vsphere.VolumeMetadata

	// Creating vsphere.VolumeMetadata using clusters EtcdClusterSpec
	for i, etcd := range e.Cluster.Spec.EtcdClusters {
		volMetadata := vsphere.VolumeMetadata{}
		volMetadata.EtcdClusterName = etcd.Name
		volMetadata.VolumeId = vsphere.GetVolumeId(i + 1)

		var members []vsphere.EtcdMemberSpec
		var thisNode string
		for _, member := range etcd.Members {
			if *member.InstanceGroup == e.IG.Name {
				thisNode = member.Name
			}
			etcdMember := vsphere.EtcdMemberSpec{
				Name:          member.Name,
				InstanceGroup: *member.InstanceGroup,
			}
			members = append(members, etcdMember)
		}

		if thisNode == "" {
			return "", fmt.Errorf("Failed to construct volume metadata for %v InstanceGroup.", e.IG.Name)
		}

		volMetadata.EtcdNodeName = thisNode
		volMetadata.Members = members
		volsMetadata = append(volsMetadata, volMetadata)
	}

	klog.V(4).Infof("Marshaling master vol metadata : %v", volsMetadata)
	volsString, err := vsphere.MarshalVolumeMetadata(volsMetadata)
	klog.V(4).Infof("Marshaled master vol metadata: %v", volsString)
	if err != nil {
		return "", err
	}
	return volsString, nil
}

func createMetaData(dir string, vmName string) error {
	data := strings.Replace(metaDataTemplate, "$INSTANCE_ID", uuid.NewUUID().String(), -1)
	data = strings.Replace(data, "$LOCAL_HOST_NAME", vmName, -1)

	klog.V(4).Infof("Meta data file content: %s", string(data))

	metaDataFile := filepath.Join(dir, "meta-data")
	if err := ioutil.WriteFile(metaDataFile, []byte(data), 0644); err != nil {
		klog.Errorf("Unable to write meta-data into file %s", metaDataFile)
		return err
	}
	return nil
}

func createISO(e *AttachISO, startupStr string, dir string, dnsServer, vmUUID string) (string, error) {
	err := createUserData(e, startupStr, dir, dnsServer, vmUUID)

	if err != nil {
		return "", err
	}
	err = createMetaData(dir, *e.VM.Name)
	if err != nil {
		return "", err
	}

	isoFile := filepath.Join(dir, *e.VM.Name+".iso")
	var commandName string

	switch os := runtime.GOOS; os {
	case "darwin":
		commandName = "mkisofs"
	case "linux":
		commandName = "genisoimage"

	default:
		return "", fmt.Errorf("Cannot generate ISO file %s. Unsupported operation system (%s)!!!", isoFile, os)
	}
	cmd := exec.Command(commandName, "-o", isoFile, "-volid", "cidata", "-joliet", "-rock", dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		klog.Errorf("Error %s occurred while executing command %+v", err, cmd)
		return "", err
	}
	klog.V(4).Infof("%s std output : %s\n", commandName, out.String())
	klog.V(4).Infof("%s std error : %s\n", commandName, stderr.String())
	return isoFile, nil
}
