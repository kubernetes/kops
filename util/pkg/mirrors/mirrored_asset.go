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

package mirrors

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops"
	"k8s.io/kops/util/pkg/hashing"
)

type MirroredAsset struct {
	Locations []string
	Hash      *hashing.Hash
}

// BuildMirroredAsset checks to see if this is a file under the standard base location, and if so constructs some mirror locations
func BuildMirroredAsset(u *url.URL, hash *hashing.Hash) *MirroredAsset {
	baseURLString := fmt.Sprintf(defaultKopsMirrorBase, kops.Version)
	if !strings.HasSuffix(baseURLString, "/") {
		baseURLString += "/"
	}

	a := &MirroredAsset{
		Hash: hash,
	}

	a.Locations = []string{u.String()}
	if strings.HasPrefix(u.String(), baseURLString) {
		if hash == nil {
			klog.Warningf("not using mirrors for asset %s as it does not have a known hash", u.String())
		} else {
			a.Locations = FindUrlMirrors(u.String())
		}
	}

	return a
}

func (a *MirroredAsset) CompactString() string {
	var s string
	if a.Hash != nil {
		s = a.Hash.Hex()
	}
	s += "@" + strings.Join(a.Locations, ",")
	return s
}
