/*
Copyright 2022 The Kubernetes Authors.

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

import (
	"context"
	"fmt"
	"net"

	"cloud.google.com/go/compute/metadata"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipyandex "k8s.io/kops/protokube/pkg/gossip/yandex"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
)

// TODO(YuraBeznos): rewrite in GCP/GCE style with discovery

// YandexCloudProvider defines the Yandex Cloud volume implementation.
type YandexCloudProvider struct {
	client *ycsdk.SDK
	server *compute.Instance

	//folderID     string
	//zone         string
	//region       string
	//clusterName  string
	instanceName string
	internalIP   net.IP
}

var _ CloudProvider = &YandexCloudProvider{}

// NewYandexCloudProvider returns a new Yandex Cloud provider.
func NewYandexCloudProvider() (*YandexCloudProvider, error) {

	sdk, err := ycsdk.Build(context.TODO(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}

	instanceId, err := metadata.InstanceID()
	if err != nil {
		return nil, err
	}
	instanceName, err := metadata.InstanceName()
	if err != nil {
		return nil, err
	}
	internalIP, err := metadata.InternalIP()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("error querying InternalIP from Yandex: %v", err)
	}
	if internalIP == "" {
		return nil, fmt.Errorf("InternalIP from metadata was empty")
	}
	internalIPNetFormat := net.ParseIP(internalIP)
	if internalIPNetFormat == nil {
		return nil, fmt.Errorf("InternalIP from metadata was not parseable(%q)", internalIP)
	}
	klog.Infof("Found internalIP=%q", internalIPNetFormat)

	instance, err := sdk.Compute().Instance().Get(context.TODO(), &compute.GetInstanceRequest{
		InstanceId: instanceId,
	})
	if err != nil {
		return nil, err
	}

	klog.Infof("Yandex instance: %q", instance.Id)

	yc := &YandexCloudProvider{
		client:       sdk,
		server:       instance,
		instanceName: instanceName,
		internalIP:   internalIPNetFormat,
	}

	return yc, nil
}

func (yc YandexCloudProvider) InstanceInternalIP() net.IP {
	return yc.internalIP
}

func (yc *YandexCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	clusterName, ok := yc.server.Labels[yandex.TagKubernetesClusterName]
	if !ok {
		return nil, fmt.Errorf("failed to find cluster name label for running server: %v", yc.server.Labels)
	}
	return gossipyandex.NewSeedProvider(yc.client, clusterName)
}

func (yc *YandexCloudProvider) InstanceID() string {
	return yc.instanceName
}
