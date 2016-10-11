package kops

import (
	"testing"
)

func Test_ParseKubernetesVersion(t *testing.T) {
	grid := map[string]string{
		"1.3.7":         "1.3.7",
		"v1.4.0-beta.8": "1.4.0-beta.8",
		"1.5.0":         "1.5.0",
		"https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/": "1.4.0",
	}
	for version, expected := range grid {
		sv, err := ParseKubernetesVersion(version)
		if err != nil {
			t.Errorf("ParseKubernetesVersion error parsing %q: %v", version, err)
		}

		actual := sv.String()
		if actual != expected {
			t.Errorf("version mismatch: %q -> %q but expected %q", version, actual, expected)
		}
	}

}
