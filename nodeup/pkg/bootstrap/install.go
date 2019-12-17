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

package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
)

type Installation struct {
	FSRoot          string
	CacheDir        string
	RunTasksOptions fi.RunTasksOptions
	Command         []string
}

func (i *Installation) Run() error {
	distribution, err := distros.FindDistribution(i.FSRoot)
	if err != nil {
		return fmt.Errorf("error determining OS distribution: %v", err)
	}

	tags := sets.NewString()
	tags.Insert(distribution.BuildTags()...)

	tasks := make(map[string]fi.Task)

	buildContext := &fi.ModelBuilderContext{
		Tasks: tasks,
	}
	i.Build(buildContext)

	// If there is a package task, we need an update packages task
	for _, t := range tasks {
		if _, ok := t.(*nodetasks.Package); ok {
			klog.Infof("Package task found; adding UpdatePackages task")
			tasks["UpdatePackages"] = nodetasks.NewUpdatePackages()
			break
		}
	}

	if tasks["UpdatePackages"] == nil {
		klog.Infof("No package task found; won't update packages")
	}

	var configBase vfs.Path
	var cloud fi.Cloud
	var keyStore fi.Keystore
	var secretStore fi.SecretStore

	target := &local.LocalTarget{
		CacheDir: i.CacheDir,
		Tags:     tags,
	}

	checkExisting := true
	context, err := fi.NewContext(target, nil, cloud, keyStore, secretStore, configBase, checkExisting, tasks)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(i.RunTasksOptions)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	err = target.Finish(tasks)
	if err != nil {
		return fmt.Errorf("error finishing target: %v", err)
	}

	return nil
}
func (i *Installation) Build(c *fi.ModelBuilderContext) {
	c.AddTask(i.buildSystemdJob())
}

func (i *Installation) buildSystemdJob() *nodetasks.Service {
	command := strings.Join(i.Command, " ")

	serviceName := "kops-configuration.service"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Run kops bootstrap (nodeup)")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	var buffer bytes.Buffer

	if os.Getenv("AWS_REGION") != "" {
		buffer.WriteString("\"AWS_REGION=")
		buffer.WriteString(os.Getenv("AWS_REGION"))
		buffer.WriteString("\" ")
	}

	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		buffer.WriteString("\"GOSSIP_DNS_CONN_LIMIT=")
		buffer.WriteString(os.Getenv("GOSSIP_DNS_CONN_LIMIT"))
		buffer.WriteString("\" ")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("S3_ENDPOINT") != "" {
		buffer.WriteString("\"S3_ENDPOINT=")
		buffer.WriteString(os.Getenv("S3_ENDPOINT"))
		buffer.WriteString("\" ")
		buffer.WriteString("\"S3_REGION=")
		buffer.WriteString(os.Getenv("S3_REGION"))
		buffer.WriteString("\" ")
		buffer.WriteString("\"S3_ACCESS_KEY_ID=")
		buffer.WriteString(os.Getenv("S3_ACCESS_KEY_ID"))
		buffer.WriteString("\" ")
		buffer.WriteString("\"S3_SECRET_ACCESS_KEY=")
		buffer.WriteString(os.Getenv("S3_SECRET_ACCESS_KEY"))
		buffer.WriteString("\" ")
	}

	// Pass in required credentials when using user-defined swift endpoint
	if os.Getenv("OS_AUTH_URL") != "" {
		for _, envVar := range []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_USERNAME",
			"OS_PASSWORD",
			"OS_AUTH_URL",
			"OS_REGION_NAME",
		} {
			buffer.WriteString("'")
			buffer.WriteString(envVar)
			buffer.WriteString("=")
			buffer.WriteString(os.Getenv(envVar))
			buffer.WriteString("' ")
		}
	}

	if os.Getenv("DIGITALOCEAN_ACCESS_TOKEN") != "" {
		buffer.WriteString("\"DIGITALOCEAN_ACCESS_TOKEN=")
		buffer.WriteString(os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))
		buffer.WriteString("\" ")
	}

	if os.Getenv("OSS_REGION") != "" {
		buffer.WriteString("\"OSS_REGION=")
		buffer.WriteString(os.Getenv("OSS_REGION"))
		buffer.WriteString("\" ")
	}

	if os.Getenv("ALIYUN_ACCESS_KEY_ID") != "" {
		buffer.WriteString("\"ALIYUN_ACCESS_KEY_ID=")
		buffer.WriteString(os.Getenv("ALIYUN_ACCESS_KEY_ID"))
		buffer.WriteString("\" ")
		buffer.WriteString("\"ALIYUN_ACCESS_KEY_SECRET=")
		buffer.WriteString(os.Getenv("ALIYUN_ACCESS_KEY_SECRET"))
		buffer.WriteString("\" ")
	}

	if buffer.String() != "" {
		manifest.Set("Service", "Environment", buffer.String())
	}

	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStart", command)
	manifest.Set("Service", "Type", "oneshot")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", serviceName, manifestString)

	service := &nodetasks.Service{
		Name:       serviceName,
		Definition: fi.String(manifestString),
	}

	service.InitDefaults()

	return service
}
