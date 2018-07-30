package bundles

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
)

type Bundle struct {
	base string
	spec bundleSpec
}

type bundleSpec struct {
	// Components specifies the components we recommend
	Components []ComponentSpec `json:"components,omitempty"`
}

type ComponentSpec struct {
	Name string `json:"name,omitempty"`

	Release string `json:"release,omitempty"`

	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}

// LoadBundle loads and parses a Bundle object from the specified VFS location
func LoadBundle(clusterSpec *kops.ClusterSpec, location string) (*Bundle, error) {
	contents, err := LoadBundleManifest(clusterSpec, location)
	if err != nil {
		return nil, err
	}

	spec, err := parseBundleSpec(contents)
	if err != nil {
		return nil, fmt.Errorf("error parsing release %q: %v", location, err)
	}
	glog.V(8).Infof("release contents: %s", string(contents))

	base := location
	lastSlash := strings.LastIndex(base, "/")
	if lastSlash != -1 {
		base = base[:lastSlash]
	}
	rs := &Bundle{
		base: base,
		spec: *spec,
	}
	return rs, nil
}

// parseBundleSpec parses bytes as a Bundle object
func parseBundleSpec(content []byte) (*bundleSpec, error) {
	r := &bundleSpec{}
	err := kops.ParseRawYaml(content, r)
	if err != nil {
		return nil, fmt.Errorf("error parsing release %v", err)
	}

	return r, nil
}

type Component struct {
	bundle *Bundle
	spec   *ComponentSpec
}

// FindComponent returns the recommended release for the component, or nil if not found
func (r *Bundle) FindComponent(componentName string, kubernetesVersion semver.Version) *Component {
	var matches []*ComponentSpec

	for i := range r.spec.Components {
		component := &r.spec.Components[i]

		if component.Name != componentName {
			continue
		}
		if component.KubernetesVersion != "" {
			versionRange, err := semver.ParseRange(component.KubernetesVersion)
			if err != nil {
				glog.Warningf("cannot parse KubernetesVersion=%q", component.KubernetesVersion)
				continue
			}

			if !versionRange(kubernetesVersion) {
				glog.V(2).Infof("Kubernetes version %q does not match range: %s", kubernetesVersion, component.KubernetesVersion)
				continue
			}
		}
		matches = append(matches, component)
	}

	if len(matches) == 0 {
		glog.V(2).Infof("No matching components for %q", componentName)
		return nil
	}

	if len(matches) != 1 {
		glog.Warningf("Multiple matching components for %q", componentName)
	}
	return &Component{
		bundle: r,
		spec:   matches[0],
	}
}

func (c *Component) Location() string {
	return strings.TrimSuffix(c.bundle.base, "/") + "/" + c.spec.Release
}
