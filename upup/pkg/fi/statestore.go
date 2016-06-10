package fi

import (
	"fmt"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
)

type StateStore interface {
	CA() CAStore
	Secrets() SecretStore
}

type VFSStateStore struct {
	location vfs.Path
	ca       CAStore
	secrets  SecretStore
}

var _ StateStore = &VFSStateStore{}

func NewVFSStateStore(location vfs.Path) (*VFSStateStore, error) {
	s := &VFSStateStore{
		location: location,
	}
	var err error
	s.ca, err = NewVFSCAStore(location.Join("pki"))
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

func (s *VFSStateStore) Secrets() SecretStore {
	return s.secrets
}
