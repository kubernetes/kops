/*
Copyright 2019 The Kubernetes Authors.

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

package protokube

//
//import (
//	"encoding/json"
//	"fmt"
//	"k8s.io/klog"
//	"io"
//	"io/ioutil"
//	"os"
//	"os/exec"
//	"path"
//	"strings"
//	"sync"
//	"time"
//)
//
//const BootstrapDir = "/etc/kubernetes/bootstrap"
//
//type BootstrapTask struct {
//	Command []string `json:"command"`
//}
//
//func (b *BootstrapTask) String() string {
//	return DebugString(b)
//}
//
//// RunKubelet runs the bootstrap tasks, and watches them until they exit
//// Currently only one task is supported / will work properly
//func (k *KubeBoot) RunBootstrapTasks() error {
//	bootstrapDir := pathFor(BootstrapDir)
//
//	var dirs []os.FileInfo
//	var err error
//
//	for {
//		dirs, err = ioutil.ReadDir(bootstrapDir)
//		if err != nil {
//			if os.IsNotExist(err) {
//				dirs = nil
//			} else {
//				return fmt.Errorf("error listing %q: %v", bootstrapDir, err)
//			}
//		}
//		if len(dirs) != 0 {
//			break
//		}
//		klog.Infof("No entries found in %q", BootstrapDir)
//		time.Sleep(10 * time.Second)
//	}
//
//	for _, dir := range dirs {
//		if !dir.IsDir() {
//			continue
//		}
//
//		p := path.Join(bootstrapDir, dir.Name())
//		files, err := ioutil.ReadDir(p)
//		if err != nil {
//			return fmt.Errorf("error listing %q: %v", p, err)
//		}
//
//		if len(files) == 0 {
//			klog.Infof("No files in %q; ignoring", p)
//			continue
//		}
//
//		// TODO: Support more than one bootstrap task?
//
//		// TODO: Have multiple proto-kubelet configurations to support recovery?
//		// i.e. launch newest version that stays up?
//
//		fp := path.Join(p, files[0].Name())
//		err = k.runBootstrapTask(fp)
//		if err != nil {
//			return fmt.Errorf("error running bootstrap task %q: %v", fp, err)
//		}
//	}
//
//	return nil
//}
//
//// RunKubelet runs a bootstrap task and watches it until it exits
//func (k *KubeBoot) runBootstrapTask(path string) error {
//	// TODO: Use a file lock or similar to only start proto-kubelet if real-kubelet is not running?
//
//	data, err := ioutil.ReadFile(path)
//	if err != nil {
//		return fmt.Errorf("error reading task %q: %v", path, err)
//	}
//
//	task := &BootstrapTask{}
//
//	err = json.Unmarshal(data, task)
//	if err != nil {
//		return fmt.Errorf("error parsing task %q: %v", path, err)
//	}
//
//	name := task.Command[0]
//	args := task.Command[1:]
//
//	cmd := exec.Command(name, args...)
//
//	stdout, err := cmd.StdoutPipe()
//	if err != nil {
//		return fmt.Errorf("error building stdout pipe: %v", err)
//	}
//	stderr, err := cmd.StderrPipe()
//	if err != nil {
//		return fmt.Errorf("error building stderr pipe: %v", err)
//	}
//
//	wg := new(sync.WaitGroup)
//	wg.Add(2)
//
//	err = cmd.Start()
//	if err != nil {
//		return fmt.Errorf("error starting command %q: %v", strings.Join(task.Command, " "), err)
//	}
//
//	go copyStream(os.Stdout, stdout, wg)
//	go copyStream(os.Stderr, stderr, wg)
//
//	wg.Wait()
//
//	err = cmd.Wait()
//	if err != nil {
//		return fmt.Errorf("error from command %q: %v", task.Command, err)
//	}
//
//	return nil
//}
//
//func copyStream(dst io.Writer, src io.ReadCloser, waitGroup *sync.WaitGroup) {
//	_, err := io.Copy(dst, src)
//	if err != nil {
//		// Not entirely sure if we need to do something special in this case?
//		klog.Warningf("error copying stream: %v", err)
//	}
//	waitGroup.Done()
//}
