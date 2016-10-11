package kops

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/golang/glog"
	"strings"
)

func ParseKubernetesVersion(version string) (*semver.Version, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		glog.Warningf("error parsing kubernetes semver %q, falling back to string matching", version)

		v := strings.Trim(version, "v")
		if strings.HasPrefix(v, "1.3.") {
			sv = semver.Version{Major: 1, Minor: 3}
		} else if strings.HasPrefix(v, "1.4.") {
			sv = semver.Version{Major: 1, Minor: 4}
		} else if strings.HasPrefix(v, "1.5.") {
			sv = semver.Version{Major: 1, Minor: 5}
		} else if strings.Contains(v, "/v1.3.") {
			sv = semver.Version{Major: 1, Minor: 3}
		} else if strings.Contains(v, "/v1.4.") {
			sv = semver.Version{Major: 1, Minor: 4}
		} else if strings.Contains(v, "/v1.5.") {
			sv = semver.Version{Major: 1, Minor: 5}
		} else {
			return nil, fmt.Errorf("unable to parse kubernetes version %q", version)
		}
	}

	return &sv, nil
}
