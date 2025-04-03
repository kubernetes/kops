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

package elementotasks

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type ServerGroup struct {
	Name      *string
	Lifecycle fi.Lifecycle
	SSHKeys   []*SSHKey
	Network   *Network

	Count      int
	NeedUpdate []string

	Location string
	Size     string
	Image    string

	EnableIPv4 bool
	EnableIPv6 bool

	UserData fi.Resource

	Labels map[string]string
}

func (v *ServerGroup) Find(c *fi.CloudupContext) (*ServerGroup, error) {

}