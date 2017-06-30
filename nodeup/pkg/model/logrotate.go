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

package model

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"strings"
)

// LogrotateBuilder installs logrotate.d and configures log rotation for kubernetes logs
type LogrotateBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &LogrotateBuilder{}

func (b *LogrotateBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Distribution == distros.DistributionCoreOS {
		glog.Infof("Detected CoreOS; won't install logrotate")
		return nil
	}

	if b.Distribution == distros.DistributionContainerOS {
		glog.Infof("Detected ContainerOS; won't install logrotate")
		return nil
	}

	c.AddTask(&nodetasks.Package{Name: "logrotate"})

	k8sVersion, err := util.ParseKubernetesVersion(b.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return fmt.Errorf("unable to parse KubernetesVersion %q", b.Cluster.Spec.KubernetesVersion)
	}

	if k8sVersion.Major == 1 && k8sVersion.Minor < 6 {
		// In version 1.6, we move log rotation to docker, but prior to that we need a logrotate rule
		b.addLogRotate(c, "docker-containers", "/var/lib/docker/containers/*/*-json.log", logRotateOptions{MaxSize: "10M"})
	}

	b.addLogRotate(c, "docker", "/var/log/docker.log", logRotateOptions{})
	b.addLogRotate(c, "kube-addons", "/var/log/kube-addons.log", logRotateOptions{})
	b.addLogRotate(c, "kube-apiserver", "/var/log/kube-apiserver.log", logRotateOptions{})
	b.addLogRotate(c, "kube-controller-manager", "/var/log/kube-controller-manager.log", logRotateOptions{})
	b.addLogRotate(c, "kube-proxy", "/var/log/kube-proxy.log", logRotateOptions{})
	b.addLogRotate(c, "kube-scheduler", "/var/log/kube-scheduler.log", logRotateOptions{})
	b.addLogRotate(c, "kubelet", "/var/log/kubelet.log", logRotateOptions{})

	// Add cron job to run hourly
	{
		script := `#!/bin/sh
logrotate /etc/logrotate.conf`

		t := &nodetasks.File{
			Path:     "/etc/cron.hourly/logrotate",
			Contents: fi.NewStringResource(script),
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	return nil
}

type logRotateOptions struct {
	MaxSize string
}

func (b *LogrotateBuilder) addLogRotate(c *fi.ModelBuilderContext, name, path string, options logRotateOptions) {
	if options.MaxSize == "" {
		options.MaxSize = "100M"
	}

	lines := []string{
		path + "{",
		"  rotate 5",
		"  copytruncate",
		"  missingok",
		"  notifempty",
		"  delaycompress",
		"  maxsize " + options.MaxSize,
		"  daily",
		"  create 0644 root root",
		"}",
	}

	contents := strings.Join(lines, "\n")

	t := &nodetasks.File{
		Path:     "/etc/logrotate.d/" + name,
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
	}
	c.AddTask(t)
}
