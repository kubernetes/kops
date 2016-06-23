package fi

import (
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"os"
	"strings"
)

type StateStore interface {
	// VFSPath returns the path where the StateStore is stored
	VFSPath() vfs.Path

	CA() CAStore
	Secrets() SecretStore

	ReadConfig(config interface{}) error
	WriteConfig(config interface{}) error
}

type VFSStateStore struct {
	location vfs.Path
	ca       CAStore
	secrets  SecretStore
}

var _ StateStore = &VFSStateStore{}

func NewVFSStateStore(location vfs.Path, dryrun bool) (*VFSStateStore, error) {
	s := &VFSStateStore{
		location: location,
	}
	var err error
	s.ca, err = NewVFSCAStore(location.Join("pki"), dryrun)
	if err != nil {
		return nil, fmt.Errorf("error building CA store: %v", err)
	}
	s.secrets, err = NewVFSSecretStore(location.Join("secrets"))
	if err != nil {
		return nil, fmt.Errorf("error building secret store: %v", err)
	}

	return s, nil
}

func (s *VFSStateStore) CA() CAStore {
	return s.ca
}

func (s *VFSStateStore) VFSPath() vfs.Path {
	return s.location
}

func (s *VFSStateStore) Secrets() SecretStore {
	return s.secrets
}

func (s *VFSStateStore) ReadConfig(config interface{}) error {
	configPath := s.location.Join("config")
	data, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading configuration file %s: %v", configPath, err)
	}

	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err = utils.YamlUnmarshal([]byte(configString), config)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

func (s *VFSStateStore) WriteConfig(config interface{}) error {
	configPath := s.location.Join("config")

	data, err := utils.YamlMarshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %v", err)
	}

	err = configPath.WriteFile(data)
	if err != nil {
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}
