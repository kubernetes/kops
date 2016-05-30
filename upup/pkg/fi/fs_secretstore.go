package fi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type FilesystemSecretStore struct {
	basedir string
}

var _ SecretStore = &FilesystemSecretStore{}

func NewFilesystemSecretStore(basedir string) (SecretStore, error) {
	c := &FilesystemSecretStore{
		basedir: basedir,
	}
	err := os.MkdirAll(path.Join(basedir), 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}
	return c, nil
}

func (c *FilesystemSecretStore) buildSecretPath(id string) string {
	return path.Join(c.basedir, id)
}

func (c *FilesystemSecretStore) FindSecret(id string) (*Secret, error) {
	p := c.buildSecretPath(id)
	s, err := c.loadSecret(p)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (c *FilesystemSecretStore) ListSecrets() ([]string, error) {
	files, err := ioutil.ReadDir(c.basedir)
	if err != nil {
		return nil, fmt.Errorf("error listing secrets directory: %v", err)
	}
	var ids []string
	for _, f := range files {
		id := f.Name()
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *FilesystemSecretStore) Secret(id string) (*Secret, error) {
	s, err := c.FindSecret(id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("Secret not found: %q", id)
	}
	return s, nil
}

func (c *FilesystemSecretStore) CreateSecret(id string) (*Secret, error) {
	p := c.buildSecretPath(id)

	s, err := CreateSecret()
	if err != nil {
		return nil, err
	}

	err = c.storeSecret(s, p)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	return c.loadSecret(p)
}

func (c *FilesystemSecretStore) loadSecret(p string) (*Secret, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	s := &Secret{}
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, fmt.Errorf("error parsing secret from %q: %v", p, err)
	}
	return s, nil
}

func (c *FilesystemSecretStore) storeSecret(s *Secret, p string) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error serializing secret: %v", err)
	}
	return c.writeFile(data, p)
}

func (c *FilesystemSecretStore) writeFile(data []byte, p string) error {
	// TODO: concurrency?
	err := ioutil.WriteFile(p, data, 0600)
	if err != nil {
		// TODO: Delete file on disk?  Write a temp file and move it atomically?
		return fmt.Errorf("error writing certificate/key data to path %q: %v", p, err)
	}
	return nil
}
