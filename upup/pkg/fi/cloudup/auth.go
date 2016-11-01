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
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

type AuthTaskBuilder struct {
}

var _ TaskBuilder = &AuthTaskBuilder{}

var _ loader.OptionsBuilder = &AuthTaskBuilder{}

func (b *AuthTaskBuilder) BuildTasks(l *Loader) error {
	return nil
}

func (b *AuthTaskBuilder) BuildOptions(options interface{}) error {
	cluster := options.(*api.ClusterSpec)

	if cluster.KubeAPIServer == nil {
		cluster.KubeAPIServer = &api.KubeAPIServerConfig{}
	}

	if cluster.KubeAPIServer.RuntimeConfig == nil {
		cluster.KubeAPIServer.RuntimeConfig = make(map[string]string)
	}

	cluster.KubeAPIServer.RuntimeConfig["rbac.authorization.k8s.io/v1alpha1"] = "true"

	cluster.KubeAPIServer.AuthorizationMode = fi.String("RBAC")

	cluster.KubeAPIServer.OIDCIssuerURL = fi.String("https://accounts.google.com")
	cluster.KubeAPIServer.OIDCClientID = fi.String("841205377713-dlmq2pe0n1ftkevna3r18livjljm5uct.apps.googleusercontent.com")
	cluster.KubeAPIServer.OIDCUsernameClaim = fi.String("email")

	cluster.KubeAPIServer.AuthorizationRBACSuperUser = fi.String("admin")

	return nil
}
