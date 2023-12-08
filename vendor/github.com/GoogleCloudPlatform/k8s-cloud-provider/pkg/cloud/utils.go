/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

var (
	domainPrefix          = "https://www.googleapis.com"
	computePrefix         = "https://www.googleapis.com/compute"
	networkServicesPrefix = "https://www.googleapis.com/networkservices"
)

// SetAPIDomain sets the root of the URL for the API. The default domain is
// "https://www.googleapis.com".
func SetAPIDomain(domain string) {
	domainPrefix = domain
	computePrefix = domain + "/compute"
	networkServicesPrefix = domain + "/networkservices"
}

// ResourceID identifies a GCE resource as parsed from compute resource URL.
type ResourceID struct {
	ProjectID string
	// APIGroup identifies the API Group of the resource.
	APIGroup meta.APIGroup
	Resource string
	Key      *meta.Key
}

// Equal returns true if two resource IDs are equal.
func (r *ResourceID) Equal(other *ResourceID) bool {
	switch {
	case r == nil && other == nil:
		return true
	case r == nil || other == nil:
		return false
	case r.ProjectID != other.ProjectID || r.Resource != other.Resource || r.APIGroup != other.APIGroup:
		return false
	case r.Key != nil && other.Key != nil:
		return *r.Key == *other.Key
	case r.Key == nil && other.Key == nil:
		return true
	default:
		return false
	}
}

// ResourceMapKey is a flat ResourceID that can be used as a key in maps.
type ResourceMapKey struct {
	ProjectID string
	APIGroup  meta.APIGroup
	Resource  string
	Name      string
	Zone      string
	Region    string
}

func (rk ResourceMapKey) ToID() *ResourceID {
	return &ResourceID{
		ProjectID: rk.ProjectID,
		APIGroup:  rk.APIGroup,
		Resource:  rk.Resource,
		Key:       &meta.Key{Name: rk.Name, Zone: rk.Zone, Region: rk.Region},
	}
}

// MapKey returns a flat key that can be used for referencing in maps.
func (r *ResourceID) MapKey() ResourceMapKey {
	return ResourceMapKey{
		ProjectID: r.ProjectID,
		APIGroup:  r.APIGroup,
		Resource:  r.Resource,
		Name:      r.Key.Name,
		Zone:      r.Key.Zone,
		Region:    r.Key.Region,
	}
}

// RelativeResourceName returns the relative resource name string
// representing this ResourceID.
// Deprecated: Use SelfLink instead
func (r *ResourceID) RelativeResourceName() string {
	return RelativeResourceName(r.ProjectID, r.Resource, r.Key)
}

// ResourcePath returns the resource path representing this ResourceID.
// Deprecated: Use SelfLink instead
func (r *ResourceID) ResourcePath() string {
	return ResourcePath(r.Resource, r.Key)
}

// SelfLink returns a URL representing the resource and defaults to Compute API
// Group if no API Group is specified.
func (r *ResourceID) SelfLink(ver meta.Version) string {
	apiGroup := r.APIGroup
	if apiGroup == "" {
		apiGroup = meta.APIGroupCompute
	}
	return SelfLinkWithGroup(apiGroup, ver, r.ProjectID, r.Resource, r.Key)
}

func (r *ResourceID) String() string {
	prefix := fmt.Sprintf("%s:%s", r.Resource, r.ProjectID)
	if r.APIGroup != "" {
		prefix = fmt.Sprintf("%s/%s", r.APIGroup, prefix)
	}
	switch r.Key.Type() {
	case meta.Zonal:
		return fmt.Sprintf("%s/%s/%s", prefix, r.Key.Zone, r.Key.Name)
	case meta.Regional:
		return fmt.Sprintf("%s/%s/%s", prefix, r.Key.Region, r.Key.Name)
	}
	return fmt.Sprintf("%s/%s", prefix, r.Key.Name)
}

// apiGroupRegex is used to extract the API Group out of a Resource URL.
// This regex expects API Group to be followed ine one of 2 patterns:
// <ver>/projects/ path or legacy one <api_group>.googleapis.com/<ver>/projects/.
// Unfortunately it cannot predict what comes before the API
// group since that is configurable via SetAPIDomain.
// legacyApiGroupRegex is used to extract API Group from legacy path in format
var apiGroupRegex = regexp.MustCompile(`([a-z]*)(\.googleapis\.com)?\/(alpha|beta|v1|v1alpha1|v1beta1)/projects`)

// ParseResourceURL parses resource URLs of the following formats:
//
//	global/<res>/<name>
//	regions/<region>/<res>/<name>
//	zones/<zone>/<res>/<name>
//	projects/<proj>
//	projects/<proj>/global/<res>/<name>
//	projects/<proj>/regions/<region>/<res>/<name>
//	projects/<proj>/zones/<zone>/<res>/<name>
//	[https://www.googleapis.com/<apigroup>/<ver>]/projects/<proj>/global/<res>/<name>
//	[https://www.googleapis.com/<apigroup>/<ver>]/projects/<proj>/regions/<region>/<res>/<name>
//	[https://www.googleapis.com/<apigroup>/<ver>]/projects/<proj>/zones/<zone>/<res>/<name>
//	[https://<apigroup>.googleapis.com/<ver>]/projects/<proj>/global/<res>/<name>
//	[https://<apigroup>.googleapis.com/<ver>]/projects/<proj>/regions/<region>/<res>/<name>
//	[https://<apigroup>.googleapis.com/<ver>]/projects/<proj>/zones/<zone>/<res>/<name>
//
// Note that ParseResourceURL can't round trip partial paths that do not
// include an API Group.
func ParseResourceURL(url string) (*ResourceID, error) {
	matches := apiGroupRegex.FindStringSubmatch(url)
	apiGroup, err := apiGroupFromMatches(matches)
	if err != nil {
		return nil, fmt.Errorf("ParseResourceURL(%q) returned error: %v", url, err)
	}
	return parseURL(url, apiGroup)
}

func apiGroupFromMatches(matches []string) (meta.APIGroup, error) {
	if len(matches) < 2 {
		return meta.APIGroup(""), nil
	}

	switch matches[1] {
	case "compute":
		return meta.APIGroupCompute, nil
	case "networkservices":
		return meta.APIGroupNetworkServices, nil
	}
	return meta.APIGroup(""), fmt.Errorf("matches does not contain a supported API Group: %v", matches)
}

func parseURL(url string, apiGroup meta.APIGroup) (*ResourceID, error) {
	errNotValid := fmt.Errorf("%q is not a valid resource URL", url)
	// Trim prefix off URL leaving "projects/..."
	projectsIndex := strings.Index(url, "/projects/")
	if projectsIndex >= 0 {
		url = url[projectsIndex+1:]
	}

	parts := strings.Split(url, "/")
	if len(parts) < 2 || len(parts) > 6 {
		return nil, errNotValid
	}

	ret := &ResourceID{APIGroup: apiGroup}
	scopedName := parts
	if parts[0] == "projects" {
		ret.Resource = "projects"
		ret.ProjectID = parts[1]
		scopedName = parts[2:]

		if len(scopedName) == 0 {
			return ret, nil
		}
	}

	switch scopedName[0] {
	case "global":
		if len(scopedName) != 3 {
			return nil, errNotValid
		}
		ret.Resource = scopedName[1]
		ret.Key = meta.GlobalKey(scopedName[2])
		return ret, nil
	case "regions":
		switch len(scopedName) {
		case 2:
			ret.Resource = "regions"
			ret.Key = meta.GlobalKey(scopedName[1])
			return ret, nil
		case 4:
			ret.Resource = scopedName[2]
			ret.Key = meta.RegionalKey(scopedName[3], scopedName[1])
			return ret, nil
		default:
			return nil, errNotValid
		}
	case "zones":
		switch len(scopedName) {
		case 2:
			ret.Resource = "zones"
			ret.Key = meta.GlobalKey(scopedName[1])
			return ret, nil
		case 4:
			ret.Resource = scopedName[2]
			ret.Key = meta.ZonalKey(scopedName[3], scopedName[1])
			return ret, nil
		default:
			return nil, errNotValid
		}
	}
	return nil, errNotValid
}

func copyViaJSON(dest, src interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

// ResourcePath returns the path starting from the location.
// Example: regions/us-central1/subnetworks/my-subnet
// Deprecated: Use SelfLinkWithGroup instead
func ResourcePath(resource string, key *meta.Key) string {
	switch resource {
	case "zones", "regions":
		return fmt.Sprintf("%s/%s", resource, key.Name)
	case "projects":
		return "invalid-resource"
	}

	switch key.Type() {
	case meta.Zonal:
		return fmt.Sprintf("zones/%s/%s/%s", key.Zone, resource, key.Name)
	case meta.Regional:
		return fmt.Sprintf("regions/%s/%s/%s", key.Region, resource, key.Name)
	case meta.Global:
		return fmt.Sprintf("global/%s/%s", resource, key.Name)
	}
	return "invalid-key-type"
}

// RelativeResourceName returns the path starting from project.
// Example: projects/my-project/regions/us-central1/subnetworks/my-subnet
// Deprecated: Use SelfLinkWithGroup instead
func RelativeResourceName(project, resource string, key *meta.Key) string {
	switch resource {
	case "projects":
		return fmt.Sprintf("projects/%s", project)
	default:
		return fmt.Sprintf("projects/%s/%s", project, ResourcePath(resource, key))
	}
}

// SelfLink returns a URL representing the resource and assumes Compute API Group.
// Deprecated: Use SelfLinkWithGroup instead
func SelfLink(ver meta.Version, project, resource string, key *meta.Key) string {
	return SelfLinkWithGroup(meta.APIGroupCompute, ver, project, resource, key)
}

// SelfLinkWithGroup returns the self link URL for the given object.
func SelfLinkWithGroup(apiGroup meta.APIGroup, ver meta.Version, project, resource string, key *meta.Key) string {
	var prefix string

	switch apiGroup {
	case meta.APIGroupCompute:
		prefix = computePrefix
	case meta.APIGroupNetworkServices:
		prefix = networkServicesPrefix
	default:
		prefix = domainPrefix + "/invalid-apigroup"
	}

	switch ver {
	case meta.VersionAlpha:
		prefix = prefix + "/alpha"
	case meta.VersionBeta:
		if apiGroup == meta.APIGroupNetworkServices {
			prefix = prefix + "/v1beta1"
		} else {
			prefix = prefix + "/beta"
		}
	case meta.VersionGA:
		prefix = prefix + "/v1"
	default:
		prefix = "invalid-version"
	}

	return fmt.Sprintf("%s/%s", prefix, RelativeResourceName(project, resource, key))
}

// aggregatedListKey return the aggregated list key based on the resource key.
func aggregatedListKey(k *meta.Key) string {
	switch k.Type() {
	case meta.Regional:
		return fmt.Sprintf("regions/%s", k.Region)
	case meta.Zonal:
		return fmt.Sprintf("zones/%s", k.Zone)
	case meta.Global:
		return "global"
	default:
		return "unknownScope"
	}
}
