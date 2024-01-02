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

package dump

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	podLogDumpConcurrency = 20
)

type podLogDumper struct {
	k8sClient    *kubernetes.Clientset
	artifactsDir string
}

type podLogDumpResult struct {
	err error
}

func NewPodLogDumper(k8sConfig *rest.Config, artifactsDir string) (*podLogDumper, error) {
	k8sConfig.QPS = 50
	k8sConfig.Burst = 100
	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}
	return &podLogDumper{
		k8sClient:    clientSet,
		artifactsDir: artifactsDir,
	}, nil
}

func (d *podLogDumper) DumpLogs(ctx context.Context) error {
	klog.Info("Dumping k8s pod logs")

	allPods, err := d.k8sClient.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing pods: %w", err)
	}

	jobs := make(chan v1.Pod, len(allPods.Items))
	results := make(chan podLogDumpResult, len(allPods.Items))

	for i := 0; i < podLogDumpConcurrency; i++ {
		go d.getPodLogs(ctx, jobs, results)
	}

	var dumpErr error

	for _, pod := range allPods.Items {
		jobs <- pod
	}
	close(jobs)

	for i := 0; i < len(allPods.Items); i++ {
		result := <-results
		if result.err != nil {
			errors.Join(dumpErr, result.err)
		}
	}
	close(results)
	return dumpErr
}

func (d *podLogDumper) getPodLogs(ctx context.Context, pods chan v1.Pod, results chan podLogDumpResult) {
	for pod := range pods {
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
			resPath := path.Join(d.artifactsDir, "cluster-info", pod.Namespace, pod.Name, container.Name)

			err := os.MkdirAll(path.Dir(resPath), 0755)
			if err != nil {
				results <- podLogDumpResult{
					err: fmt.Errorf("creating directory %q: %w", resPath, err),
				}
				continue
			}

			{
				resp, err := d.k8sClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{Container: container.Name, Previous: true}).Do(ctx).Raw()
				var statusErr *k8sErrors.StatusError
				if errors.As(err, &statusErr) {
					if statusErr.ErrStatus.Code != 400 {
						results <- podLogDumpResult{
							err: fmt.Errorf("getting pod logs for the previous instance of %v/%v: %w", pod.Namespace, pod.Name, err),
						}
					}
				} else {
					err := writeContainerLogs(resPath+".previous.log", resp)
					if err != nil {
						results <- podLogDumpResult{
							err: err,
						}
					}
				}
			}

			{
				resp, err := d.k8sClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{Container: container.Name}).Do(ctx).Raw()
				var statusErr *k8sErrors.StatusError
				if errors.As(err, &statusErr) {
					if statusErr.ErrStatus.Code != 400 {
						results <- podLogDumpResult{
							err: fmt.Errorf("getting pod logs for the current instance of %v/%v: %w", pod.Namespace, pod.Name, err),
						}
						continue
					}
				} else {
					err := writeContainerLogs(resPath+".log", resp)
					if err != nil {
						results <- podLogDumpResult{
							err: err,
						}
					}
				}
			}
		}

		results <- podLogDumpResult{}
	}
}

func writeContainerLogs(filePath string, contents []byte) error {
	resFile, err := os.Create(filePath)
	defer func(resFile *os.File) {
		_ = resFile.Close()
	}(resFile)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", filePath, err)
	}
	_, err = resFile.Write(contents)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", filePath, err)
	}
	return nil
}
