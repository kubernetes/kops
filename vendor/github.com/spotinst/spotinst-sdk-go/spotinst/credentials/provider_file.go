package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// FileCredentialsProviderName provides a name of File provider.
	FileCredentialsProviderName = "FileCredentialsProvider"

	// FileCredentialsEnvVarFile specifies the name of the environment variable
	// points to the location of the credentials file.
	FileCredentialsEnvVarFile = "SPOTINST_CREDENTIALS_FILE"
)

var (
	// ErrFileCredentialsHomeNotFound is emitted when the user directory
	// cannot be found.
	ErrFileCredentialsHomeNotFound = errors.New("spotinst: user home directory not found")

	// ErrFileCredentialsLoadFailed is emitted when the provider is unable to
	// load credentials from the credentials file.
	ErrFileCredentialsLoadFailed = errors.New("spotinst: failed to load credentials file")

	// ErrFileCredentialsTokenNotFound is emitted when the loaded credentials
	// did not contain a valid token.
	ErrFileCredentialsTokenNotFound = errors.New("spotinst: credentials did not contain token")
)

// A FileProvider retrieves credentials from the current user's home
// directory.
type FileProvider struct {
	// Path to the credentials file.
	//
	// If empty will look for FileCredentialsEnvVarFile env variable. If the
	// env value is empty will default to current user's home directory.
	// Linux/OSX: "$HOME/.spotinst/credentials.json"
	// Windows:   "%USERPROFILE%\.spotinst\credentials.json"
	Filename string

	// retrieved states if the credentials have been successfully retrieved.
	retrieved bool
}

// NewFileCredentials returns a pointer to a new Credentials object
// wrapping the file provider.
func NewFileCredentials(filename string) *Credentials {
	return NewCredentials(&FileProvider{
		Filename: filename,
	})
}

// Retrieve reads and extracts the shared credentials from the current
// users home directory.
func (p *FileProvider) Retrieve() (Value, error) {
	p.retrieved = false

	filename, err := p.filename()
	if err != nil {
		return Value{ProviderName: FileCredentialsProviderName}, err
	}

	creds, err := p.loadCredentials(filename)
	if err != nil {
		return Value{ProviderName: FileCredentialsProviderName}, err
	}

	if len(creds.ProviderName) == 0 {
		creds.ProviderName = FileCredentialsProviderName
	}

	p.retrieved = true
	return creds, nil
}

func (p *FileProvider) String() string {
	return FileCredentialsProviderName
}

// filename returns the filename to use to read Spotinst credentials.
//
// Will return an error if the user's home directory path cannot be found.
func (p *FileProvider) filename() (string, error) {
	if p.Filename == "" {
		if p.Filename = os.Getenv(FileCredentialsEnvVarFile); p.Filename != "" {
			return p.Filename, nil
		}

		homeDir := os.Getenv("HOME") // *nix
		if homeDir == "" {           // Windows
			homeDir = os.Getenv("USERPROFILE")
		}
		if homeDir == "" {
			return "", ErrFileCredentialsHomeNotFound
		}

		p.Filename = filepath.Join(homeDir, ".spotinst", "credentials")
	}

	return p.Filename, nil
}

// loadCredentials loads the credentials from the file pointed to by filename.
// The credentials retrieved from the profile will be returned or error. Error will be
// returned if it fails to read from the file, or the data is invalid.
func (p *FileProvider) loadCredentials(filename string) (Value, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Value{ProviderName: FileCredentialsProviderName},
			fmt.Errorf("%s: %s", ErrFileCredentialsLoadFailed.Error(), err)
	}
	defer f.Close()

	var value Value
	if err := json.NewDecoder(f).Decode(&value); err != nil {
		return Value{ProviderName: FileCredentialsProviderName},
			fmt.Errorf("%s: %s", ErrFileCredentialsLoadFailed.Error(), err)
	}
	if token := value.Token; len(token) == 0 {
		return Value{ProviderName: FileCredentialsProviderName},
			ErrFileCredentialsTokenNotFound
	}

	return value, nil
}
