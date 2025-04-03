/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagKubernetesClusterName         = "kops.k8s.io/cluster"
	TagKubernetesFirewallRole        = "kops.k8s.io/firewall-role"
	TagKubernetesInstanceGroup       = "kops.k8s.io/instance-group"
	TagKubernetesInstanceRole        = "kops.k8s.io/instance-role"
	TagKubernetesInstanceUserData    = "kops.k8s.io/instance-userdata"
	TagKubernetesInstanceNeedsUpdate = "kops.k8s.io/needs-update"
	TagKubernetesVolumeRole          = "kops.k8s.io/volume-role"
	TagKubernetesNodeLabelPrefix     = "node-label.kops.k8s.io."
)

// ElementoCloud exposes all the interfaces required to operate on the Elemento cloud
type ElementoCloud struct {
	fi.Cloud

	// TODO: Detect and add additional fields here
}

var _ fi.Cloud = &elementoCloudImplementation{}

// Interaction with Elemento cloud resources
type elementoCloudImplementation struct {
	Client *ecloud.Client

	region string
	// TODO: Add additional fields here
}

func NewElementoCloud(region string) (ElementoCloud, error) {
	accessToken := os.Getenv("ELEMENTO_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("ELEMENTO_ACCESS_TOKEN is required")
	}

	opts := []ecloud.ClientOption{
		ecloud.WithAccessToken(accessToken),
	}

	client := ecloud.NewClient(opts...)

	return &elementoCloudImplementation{
		Client: client,
		region: region,
	}, nil
}

// TODO: add implementation functions here