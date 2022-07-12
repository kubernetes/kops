package scw

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"gopkg.in/yaml.v2"
)

// configV1 is a Scaleway CLI configuration file
type configV1 struct {
	// Organization is the identifier of the Scaleway organization
	Organization string `json:"organization"`

	// Token is the authentication token for the Scaleway organization
	Token string `json:"token"`

	// Version is the actual version of scw CLI
	Version string `json:"version"`
}

func unmarshalConfV1(content []byte) (*configV1, error) {
	var config configV1
	err := json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, err
}

func (v1 *configV1) toV2() *Config {
	return &Config{
		Profile: Profile{
			DefaultOrganizationID: &v1.Organization,
			DefaultProjectID:      &v1.Organization, // v1 config is not aware of project, so default project is set to organization ID
			SecretKey:             &v1.Token,
			// ignore v1 version
		},
	}
}

// MigrateLegacyConfig will migrate the legacy config to the V2 when none exist yet.
// Returns a boolean set to true when the migration happened.
// TODO: get accesskey from account?
func MigrateLegacyConfig() (bool, error) {
	// STEP 1: try to load config file V2
	v2Path, v2PathOk := getConfigV2FilePath()
	if !v2PathOk || fileExist(v2Path) {
		return false, nil
	}

	// STEP 2: try to load config file V1
	v1Path, v1PathOk := getConfigV1FilePath()
	if !v1PathOk {
		return false, nil
	}
	file, err := ioutil.ReadFile(v1Path)
	if err != nil {
		return false, nil
	}
	confV1, err := unmarshalConfV1(file)
	if err != nil {
		return false, errors.Wrap(err, "content of config file %s is invalid json", v1Path)
	}

	// STEP 3: create dir
	err = os.MkdirAll(filepath.Dir(v2Path), 0700)
	if err != nil {
		return false, errors.Wrap(err, "mkdir did not work on %s", filepath.Dir(v2Path))
	}

	// STEP 4: marshal yaml config
	newConfig := confV1.toV2()
	file, err = yaml.Marshal(newConfig)
	if err != nil {
		return false, err
	}

	// STEP 5: save config
	err = ioutil.WriteFile(v2Path, file, defaultConfigPermission)
	if err != nil {
		return false, errors.Wrap(err, "cannot write file %s", v2Path)
	}

	// STEP 6: log success
	logger.Warningf("migrated existing config to %s", v2Path)
	return true, nil
}
