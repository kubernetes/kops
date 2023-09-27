package scaleway

import (
	"os"
	"testing"
)

func TestGetServerPrivateIP(t *testing.T) {
	err := os.Setenv("SCW_PROFILE", "default")
	if err != nil {
		t.Errorf("setting env: %v", err)
	}
	scwCloud, err := NewScwCloud(nil)
	//scwCloud, err := NewScwCloud(map[string]string{"zone": "fr-par-1", "region": "fr-par"})
	if err != nil {
		t.Error(err)
	}
	ip, err := scwCloud.GetServerPrivateIP("scw-wizardly-mestorf", "fr-par-1")
	if err != nil {
		t.Error(err)
	}
	if ip != "10.76.76.41" {
		t.Errorf("expected 10.76.76.41, got %s", ip)
	}
}
