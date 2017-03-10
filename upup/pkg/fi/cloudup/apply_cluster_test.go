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
	"fmt"
	"strings"
	"testing"

	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup"
	"k8s.io/kops/util/pkg/vfs"
)

func testRenderNodeUPConfig(ig *api.InstanceGroup) (*nodeup.NodeUpConfig, error) {
	if ig == nil {
		return nil, fmt.Errorf("instanceGroup cannot be nil")
	}

	role := ig.Spec.Role
	if role == "" {
		return nil, fmt.Errorf("cannot determine role for instance group: %v", ig.ObjectMeta.Name)
	}

	cluster := buildMinimalCluster()

	cluster.Spec.Networking = &api.NetworkingSpec{
		Weave: &api.WeaveNetworkingSpec{},
	}

	clusterTags, err := buildCloudupTags(cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to build cluster tags")
	}
	nodeUpTags, err := buildNodeupTags(role, cluster, clusterTags)
	if err != nil {
		return nil, err
	}

	config := &nodeup.NodeUpConfig{}
	for _, tag := range nodeUpTags.List() {
		config.Tags = append(config.Tags, tag)
	}

	config.Assets = []string{}

	config.ClusterName = "clusterName"

	config.ConfigBase = fi.String("path")

	config.InstanceGroupName = ig.ObjectMeta.Name

	var images []*nodeup.Image

	{
		location := ProtokubeImageSource()

		hash, err := findHash(location)
		if err != nil {
			return nil, err
		}

		config.ProtokubeImage = &nodeup.Image{
			Name:   kops.DefaultProtokubeImageName(),
			Source: location,
			Hash:   hash.Hex(),
		}
	}

	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("error parsing config base %q: %v", cluster.Spec.ConfigBase, err)
	}

	config.Images = images
	config.Channels = []string{
		configBase.Join("addons", "bootstrap-channel.yaml").Path(),
	}

	return config, nil
}

func TestBootstrapScript(t *testing.T) {

	g := buildMinimalNodeInstanceGroup()

	bootstrapScriptBuilder := &model.BootstrapScript{
		NodeUpConfigBuilder: testRenderNodeUPConfig,
		NodeUpSourceHash:    "",
		NodeUpSource:        NodeUpLocation(),
	}

	holder, err := bootstrapScriptBuilder.ResourceNodeUp(g)

	if err != nil {
		t.Errorf("error calling boostrapScriptBuilder: %v", err)
	}

	s, err := holder.AsString()

	if err != nil {
		t.Errorf("error calling holder.AsString(): %v", err)
	}

	if strings.Contains(s, "KOPS_FEATURE_FLAGS=") {
		t.Errorf("feature flags should not be enabled")
	}

}

func TestBootstrapScript_Feature_Flag(t *testing.T) {

	featureflag.ParseFlags("+ExperimentalCriticalPodAnnotation")

	if !featureflag.ExperimentalCriticalPodAnnotation.Enabled() {
		t.Errorf("feature flag should be enabled")
	}

	g := buildMinimalNodeInstanceGroup()

	bootstrapScriptBuilder := &model.BootstrapScript{
		NodeUpConfigBuilder: testRenderNodeUPConfig,
		NodeUpSourceHash:    "",
		NodeUpSource:        NodeUpLocation(),
	}

	holder, err := bootstrapScriptBuilder.ResourceNodeUp(g)

	if err != nil {
		t.Errorf("error calling boostrapScriptBuilder: %v", err)
	}

	s, err := holder.AsString()

	if err != nil {
		t.Errorf("error calling holder.AsString(): %v", err)
	}

	if !strings.Contains(s, "KOPS_FEATURE_FLAGS=\"+ExperimentalCriticalPodAnnotation\"") {
		t.Errorf("feature flag should be enabled in the bootstrap script\n %s", s)
	}

}
