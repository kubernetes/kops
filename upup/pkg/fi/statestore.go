package fi

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"os"
	"strings"
)

type WriteOption string

const (
	WriteOptionCreate       WriteOption = "Create"
	WriteOptionOnlyIfExists WriteOption = "IfExists"
)

type StateStore interface {
	// VFSPath returns the path where the StateStore is stored
	VFSPath() vfs.Path

	CA() CAStore
	Secrets() SecretStore

	ReadConfig(path string, config interface{}) error
	WriteConfig(path string, config interface{}, options ...WriteOption) error

	// ListChildren returns a list of all (direct) children of the specified path
	// It only returns the raw names, not the prefixes
	ListChildren(pathPrefix string) ([]string, error)
}

type VFSStateStore struct {
	location vfs.Path
	keystore CAStore
	secrets  SecretStore
}

var _ StateStore = &VFSStateStore{}

func NewVFSStateStore(base vfs.Path, clusterName string) *VFSStateStore {
	location := base.Join(clusterName)
	s := &VFSStateStore{
		location: location,
	}
	s.keystore = NewVFSCAStore(location.Join("pki"))
	s.secrets = NewVFSSecretStore(location.Join("secrets"))
	return s
}

func (s *VFSStateStore) CA() CAStore {
	return s.keystore
}

func (s *VFSStateStore) VFSPath() vfs.Path {
	return s.location
}

func (s *VFSStateStore) Secrets() SecretStore {
	return s.secrets
}

func (s *VFSStateStore) ListChildren(pathPrefix string) ([]string, error) {
	vfsPath := s.location.Join(pathPrefix)
	children, err := vfsPath.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing children of %s: %v", pathPrefix, err)
	}

	var names []string
	for _, child := range children {
		names = append(names, child.Base())
	}
	return names, nil
}

func (s *VFSStateStore) ReadConfig(path string, config interface{}) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	configPath := s.location.Join(path)
	data, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return err
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

func (s *VFSStateStore) WriteConfig(path string, config interface{}, writeOptions ...WriteOption) error {
	configPath := s.location.Join(path)

	data, err := utils.YamlMarshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %v", err)
	}

	create := false
	for _, writeOption := range writeOptions {
		switch writeOption {
		case WriteOptionCreate:
			create = true
		case WriteOptionOnlyIfExists:
			_, err = configPath.ReadFile()
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("cannot update configuration file %s: does not exist", configPath)
				}
				return fmt.Errorf("error checking if configuration file %s exists already: %v", configPath, err)
			}
		default:
			return fmt.Errorf("unknown write option: %q", writeOption)
		}
	}

	if create {
		err = configPath.CreateFile(data)
	} else {
		err = configPath.WriteFile(data)
	}
	if err != nil {
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}
