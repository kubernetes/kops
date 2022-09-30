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
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/vfs"
)

type Installation struct {
	CacheDir        string
	RunTasksOptions fi.RunTasksOptions
	Command         []string
}

func (i *Installation) Run() error {
	_, err := distributions.FindDistribution("/")
	if err != nil {
		return fmt.Errorf("error determining OS distribution: %v", err)
	}

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
	c.AddTask(i.buildEnvFile())
	c.AddTask(i.buildSystemdJob())
}

func (i *Installation) buildEnvFile() *nodetasks.File {
	envVars := make(map[string]string)

	if os.Getenv("AWS_REGION") != "" {
		envVars["AWS_REGION"] = os.Getenv("AWS_REGION")
	}

	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		envVars["GOSSIP_DNS_CONN_LIMIT"] = os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	}

	// Pass in required credentials when using user-defined s3 endpoint
	if os.Getenv("S3_ENDPOINT") != "" {
		envVars["S3_ENDPOINT"] = os.Getenv("S3_ENDPOINT")
		envVars["S3_REGION"] = os.Getenv("S3_REGION")
		envVars["S3_ACCESS_KEY_ID"] = os.Getenv("S3_ACCESS_KEY_ID")
		envVars["S3_SECRET_ACCESS_KEY"] = os.Getenv("S3_SECRET_ACCESS_KEY")
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
			"OS_APPLICATION_CREDENTIAL_ID",
			"OS_APPLICATION_CREDENTIAL_SECRET",
		} {
			envVars[envVar] = os.Getenv(envVar)
		}
	}

	if os.Getenv("DIGITALOCEAN_ACCESS_TOKEN") != "" {
		envVars["DIGITALOCEAN_ACCESS_TOKEN"] = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}

	if os.Getenv("HCLOUD_TOKEN") != "" {
		envVars["HCLOUD_TOKEN"] = os.Getenv("HCLOUD_TOKEN")
	}

	if os.Getenv("OSS_REGION") != "" {
		envVars["OSS_REGION"] = os.Getenv("OSS_REGION")
	}

	if os.Getenv("ALIYUN_ACCESS_KEY_ID") != "" {
		envVars["ALIYUN_ACCESS_KEY_ID"] = os.Getenv("ALIYUN_ACCESS_KEY_ID")
		envVars["ALIYUN_ACCESS_KEY_SECRET"] = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	}

	if os.Getenv("AZURE_STORAGE_ACCOUNT") != "" {
		envVars["AZURE_STORAGE_ACCOUNT"] = os.Getenv("AZURE_STORAGE_ACCOUNT")
	}

	if os.Getenv("SCW_SECRET_KEY") != "" {
		envVars["SCW_ACCESS_KEY"] = os.Getenv("SCW_ACCESS_KEY")
		envVars["SCW_SECRET_KEY"] = os.Getenv("SCW_SECRET_KEY")
		envVars["SCW_DEFAULT_PROJECT_ID"] = os.Getenv("SCW_DEFAULT_PROJECT_ID")
		envVars["SCW_DEFAULT_REGION"] = os.Getenv("SCW_DEFAULT_REGION")
		envVars["SCW_DEFAULT_ZONE"] = os.Getenv("SCW_DEFAULT_ZONE")
	}

	if os.Getenv("YANDEX_CLOUD_CREDENTIAL_FILE") != "" {
		envVars["YANDEX_CLOUD_CREDENTIAL_FILE"] = os.Getenv("YANDEX_CLOUD_CREDENTIAL_FILE")
	}

	sysconfig := ""
	for key, value := range envVars {
		sysconfig += key + "=" + value + "\n"
	}

	task := &nodetasks.File{
		Path:     "/etc/sysconfig/kops-configuration",
		Contents: fi.NewStringResource(sysconfig),
		Type:     nodetasks.FileType_File,
	}

	return task
}

func (i *Installation) buildSystemdJob() *nodetasks.Service {
	command := strings.Join(i.Command, " ")

	serviceName := "kops-configuration.service"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Run kOps bootstrap (nodeup)")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/kops-configuration")
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
