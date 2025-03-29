package credentials

import (
	"net"
	"net/url"
	"strings"

	"github.com/docker/cli/cli/config/types"
)

type store interface {
	Save() error
	GetAuthConfigs() map[string]types.AuthConfig
	GetFilename() string
}

// fileStore implements a credentials store using
// the docker configuration file to keep the credentials in plain text.
type fileStore struct {
	file store
}

// NewFileStore creates a new file credentials store.
func NewFileStore(file store) Store {
	return &fileStore{file: file}
}

// Erase removes the given credentials from the file store.This function is
// idempotent and does not update the file if credentials did not change.
func (c *fileStore) Erase(serverAddress string) error {
	if _, exists := c.file.GetAuthConfigs()[serverAddress]; !exists {
		// nothing to do; no credentials found for the given serverAddress
		return nil
	}
	delete(c.file.GetAuthConfigs(), serverAddress)
	return c.file.Save()
}

// Get retrieves credentials for a specific server from the file store.
func (c *fileStore) Get(serverAddress string) (types.AuthConfig, error) {
	authConfig, ok := c.file.GetAuthConfigs()[serverAddress]
	if !ok {
		// Maybe they have a legacy config file, we will iterate the keys converting
		// them to the new format and testing
		for r, ac := range c.file.GetAuthConfigs() {
			if serverAddress == ConvertToHostname(r) {
				return ac, nil
			}
		}

		authConfig = types.AuthConfig{}
	}
	return authConfig, nil
}

func (c *fileStore) GetAll() (map[string]types.AuthConfig, error) {
	return c.file.GetAuthConfigs(), nil
}

// Store saves the given credentials in the file store. This function is
// idempotent and does not update the file if credentials did not change.
func (c *fileStore) Store(authConfig types.AuthConfig) error {
	authConfigs := c.file.GetAuthConfigs()
	if oldAuthConfig, ok := authConfigs[authConfig.ServerAddress]; ok && oldAuthConfig == authConfig {
		// Credentials didn't change, so skip updating the configuration file.
		return nil
	}
	authConfigs[authConfig.ServerAddress] = authConfig
	return c.file.Save()
}

func (c *fileStore) GetFilename() string {
	return c.file.GetFilename()
}

func (c *fileStore) IsFileStore() bool {
	return true
}

// ConvertToHostname converts a registry url which has http|https prepended
// to just an hostname.
// Copied from github.com/docker/docker/registry.ConvertToHostname to reduce dependencies.
func ConvertToHostname(maybeURL string) string {
	stripped := maybeURL
	if strings.Contains(stripped, "://") {
		u, err := url.Parse(stripped)
		if err == nil && u.Hostname() != "" {
			if u.Port() == "" {
				return u.Hostname()
			}
			return net.JoinHostPort(u.Hostname(), u.Port())
		}
	}
	hostName, _, _ := strings.Cut(stripped, "/")
	return hostName
}
