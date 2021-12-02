/*
Copyright 2020 The Kubernetes Authors.

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

package templater

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver/v4"
	"k8s.io/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/util/pkg/architectures"
)

// templateFuncsMap returns a map if the template functions for this template
func (r *Templater) templateFuncsMap(tm *template.Template) template.FuncMap {
	// grab the template functions from sprig which are pretty awesome
	funcs := sprig.TxtFuncMap()

	funcs["indent"] = indentContent
	// @step: as far as i can see there's no native way in sprig in include external snippets of code
	funcs["include"] = func(name string, context map[string]interface{}) string {
		content, err := includeSnippet(tm, name, context)
		if err != nil {
			panic(err.Error())
		}

		return content
	}

	funcs["ChannelRecommendedKubernetesUpgradeVersion"] = func(version string) string {
		parsed, err := util.ParseKubernetesVersion(version)
		if err != nil {
			panic(err.Error())
		}

		versionInfo := kopsapi.FindKubernetesVersionSpec(r.channel.Spec.KubernetesVersions, *parsed)
		recommended, err := versionInfo.FindRecommendedUpgrade(*parsed)
		if err != nil {
			panic(err.Error())
		}
		return recommended.String()
	}

	funcs["ChannelRecommendedKopsKubernetesVersion"] = func() string {
		return kopsapi.RecommendedKubernetesVersion(r.channel, kops.Version).String()
	}

	funcs["ChannelRecommendedImage"] = func(cloud, k8sVersion string, architecture string) string {
		ver, _ := semver.ParseTolerant(k8sVersion)
		imageSpec := r.channel.FindImage(kopsapi.CloudProviderID(cloud), ver, architectures.Architecture(architecture))
		return imageSpec.Name
	}

	return funcs
}
