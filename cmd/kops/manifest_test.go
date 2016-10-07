package main

// Manifest CRUD tests
// Note: The CRUD operations are temporarily skipped until Kops can support a testing environment
// to enable these tests remove the `t.SkipNow()` methods, and flesh out manifest_example.yml with
// a valid test environment configuration.
// - or -
// Create a new manifest testing .yml file and update the manifest var or use KOPS_TEST_MANIFEST ENV VAR

import (
	"testing"
	"os"
)

var default_manifest = "../../manifest_example.yml"

func getManifestString() string {
	env := os.Getenv("KOPS_TEST_MANIFEST")
	if env != "" {
		return env
	}
	return default_manifest
}

// Will test that getManifest() works on our example .yml file.
// Also verifies the .yml file is meaningful
func TestManifestParsingHappy(t *testing.T) {
	v, err := getManifest(getManifestString())
	if err != nil {
		t.Error(err)
	}
	if len(v.AllKeys()) < 1 {
		t.Error("Empty or invalid manifest_example.yml")
	}
}

// Sad path
func TestManifestParsingSad(t *testing.T) {
	_, err := getManifest("bad_manifest")
	if err == nil {
		t.Error("Sad path failure " + err.Error())
	}
}

// Will test a kops create with a manifest
func TestCreateClusterWithManifest(t *testing.T) {
	t.SkipNow()
	cluster := CreateClusterCmd{}
	cluster.Filename = getManifestString()
	var args []string
	err := cluster.Run(args)
	if err != nil {
		t.Error(err)
	}

}

// Will test a kops get with a manifest
func TestGetClusterWithManifest(t *testing.T) {
	t.SkipNow()
	cluster := GetClustersCmd{}
	cluster.Filename = getManifestString()
	var args []string
	err := cluster.Run(args)
	if err != nil {
		t.Error(err)
	}
}

// Will test a kops update with a manifest
func TestUpdateClusterWithManifest(t *testing.T) {
	t.SkipNow()
	cluster := UpdateClusterCmd{}
	cluster.Filename = getManifestString()
	var args []string
	err := cluster.Run(args)
	if err != nil {
		t.Error(err)
	}
}

// Will test a kops delete with a manifest
func TestDeleteClusterWithManifest(t *testing.T) {
	t.SkipNow()
	cluster := DeleteClusterCmd{}
	cluster.Filename = getManifestString()
	var args []string
	err := cluster.Run(args)
	if err != nil {
		t.Error(err)
	}
}
