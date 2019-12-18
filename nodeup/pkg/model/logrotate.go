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

package model

import (
	"fmt"
	"strings"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
)

// LogrotateBuilder installs logrotate.d and configures log rotation for kubernetes logs
type LogrotateBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &LogrotateBuilder{}

// Build is responsible for configuring logrotate
func (b *LogrotateBuilder) Build(c *fi.ModelBuilderContext) error {

	switch b.Distribution {
	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install logrotate")
		return nil
	case distros.DistributionCoreOS:
		klog.Infof("Detected CoreOS; won't install logrotate")
	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install logrotate")
	default:
		c.AddTask(&nodetasks.Package{Name: "logrotate"})
	}

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
	b.addLogRotate(c, "etcd", "/var/log/etcd.log", logRotateOptions{})
	b.addLogRotate(c, "etcd-events", "/var/log/etcd-events.log", logRotateOptions{})

	if b.Cluster.Spec.KubeAPIServer.AuditLogPath != nil {
		b.addLogRotate(c, "kube-audit", *b.Cluster.Spec.KubeAPIServer.AuditLogPath, logRotateOptions{})
	}
	
	if err := b.addLogrotateService(c); err != nil {
		return err
	}

	// Add timer to run hourly.
	{
		unit := &systemd.Manifest{}
		unit.Set("Unit", "Description", "Hourly Log Rotation")
		unit.Set("Timer", "OnCalendar", "hourly")

		service := &nodetasks.Service{
			Name:       "logrotate.timer", // Override (by name) any existing timer
			Definition: s(unit.Render()),
		}

		service.InitDefaults()

		c.AddTask(service)
	}

	return nil
}

// addLogrotateService creates a logrotate systemd task to act as target for the timer, if one is needed
func (b *LogrotateBuilder) addLogrotateService(c *fi.ModelBuilderContext) error {
	switch b.Distribution {
	case distros.DistributionCoreOS, distros.DistributionFlatcar, distros.DistributionContainerOS:
		// logrotate service already exists
		return nil
	}

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Rotate and Compress System Logs")
	manifest.Set("Service", "ExecStart", "/usr/sbin/logrotate /etc/logrotate.conf")

	service := &nodetasks.Service{
		Name:       "logrotate.service",
		Definition: s(manifest.Render()),
	}
	service.InitDefaults()
	c.AddTask(service)

	return nil
}

type logRotateOptions struct {
	MaxSize    string
	DateFormat string
}

func (b *LogrotateBuilder) addLogRotate(c *fi.ModelBuilderContext, name, path string, options logRotateOptions) {
	if options.MaxSize == "" {
		options.MaxSize = "100M"
	}

	// CoreOS sets "dateext" options, and maxsize-based rotation will fail if
	// the file has been previously rotated on the same calendar date.
	if b.Distribution == distros.DistributionCoreOS {
		options.DateFormat = "-%Y%m%d-%s"
	}

	// Flatcar sets "dateext" options, and maxsize-based rotation will fail if
	// the file has been previously rotated on the same calendar date.
	if b.Distribution == distros.DistributionFlatcar {
		options.DateFormat = "-%Y%m%d-%s"
	}

	lines := []string{
		path + "{",
		"  rotate 5",
		"  copytruncate",
		"  missingok",
		"  notifempty",
		"  delaycompress",
		"  maxsize " + options.MaxSize,
	}

	if options.DateFormat != "" {
		lines = append(lines, "  dateformat "+options.DateFormat)
	}

	lines = append(
		lines,
		"  daily",
		"  create 0644 root root",
		"}",
	)

	contents := strings.Join(lines, "\n") + "\n"

	c.AddTask(&nodetasks.File{
		Path:     "/etc/logrotate.d/" + name,
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
	})
}
