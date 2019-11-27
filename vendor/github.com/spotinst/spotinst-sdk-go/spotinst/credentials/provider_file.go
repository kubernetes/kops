package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-ini/ini"
)

const (
	// FileCredentialsProviderName provides a name of File provider.
	FileCredentialsProviderName = "FileCredentialsProvider"

	// FileCredentialsEnvVarFile specifies the name of the environment variable
	// points to the location of the credentials file.
	FileCredentialsEnvVarFile = "SPOTINST_CREDENTIALS_FILE"

	// FileCredentialsEnvVarProfile specifies the name of the environment variable
	// points to a profile name to use when loading credentials.
	FileCredentialsEnvVarProfile = "SPOTINST_CREDENTIALS_PROFILE"
)

var (
	// ErrFileCredentialsLoadFailed is emitted when the provider is unable to load
	// credentials from the credentials file.
	ErrFileCredentialsLoadFailed = errors.New("spotinst: failed to load credentials file")

	// ErrFileCredentialsTokenNotFound is emitted when the loaded credentials
	// do not contain a valid token.
	ErrFileCredentialsTokenNotFound = errors.New("spotinst: credentials do not contain token")
)

// DefaultProfile returns the SDK's default profile name to use when loading
// credentials.
func DefaultProfile() string {
	return "default"
}

// DefaultFilename returns the SDK's default file path for the credentials file.
//
// Builds the config file path based on the OS's platform.
//   - Linux/Unix : $HOME/.spotinst/credentials
//   - Windows    : %USERPROFILE%\.spotinst\credentials
func DefaultFilename() string {
	return filepath.Join(userHomeDir(), ".spotinst", "credentials")
}

// A FileProvider retrieves credentials from the current user's home directory.
type FileProvider struct {
	// Profile to load.
	Profile string

	// Path to the credentials file.
	//
	// If empty will look for FileCredentialsEnvVarFile env variable. If the
	// env value is empty will default to current user's home directory.
	// - Linux/Unix : $HOME/.spotinst/credentials
	// - Windows    : %USERPROFILE%\.spotinst\credentials
	Filename string

	// retrieved states if the credentials have been successfully retrieved.
	retrieved bool
}

// NewFileCredentials returns a pointer to a new Credentials object wrapping the
// file provider.
func NewFileCredentials(profile, filename string) *Credentials {
	return NewCredentials(&FileProvider{
		Profile:  profile,
		Filename: filename,
	})
}

// Retrieve reads and extracts the shared credentials from the current users home
// directory.
func (p *FileProvider) Retrieve() (Value, error) {
	p.retrieved = false

	value, err := p.loadCredentials(p.profile(), p.filename())
	if err != nil {
		return value, err
	}

	if len(value.ProviderName) == 0 {
		value.ProviderName = FileCredentialsProviderName
	}

	p.retrieved = true
	return value, nil
}

func (p *FileProvider) String() string {
	return FileCredentialsProviderName
}

// profile returns the profile to use to read Spotinst credentials.
func (p *FileProvider) profile() string {
	if p.Profile == "" {
		if p.Profile = os.Getenv(FileCredentialsEnvVarProfile); p.Profile != "" {
			return p.Profile
		}

		p.Profile = DefaultProfile()
	}

	return p.Profile
}

// filename returns the filename to use to read Spotinst credentials.
func (p *FileProvider) filename() string {
	if p.Filename == "" {
		if p.Filename = os.Getenv(FileCredentialsEnvVarFile); p.Filename != "" {
			return p.Filename
		}

		p.Filename = DefaultFilename()
	}

	return p.Filename
}

// loadCredentials loads the credentials from the file pointed to by filename.
// The credentials retrieved from the profile will be returned or error. Error
// will be returned if it fails to read from the file, or the data is invalid.
func (p *FileProvider) loadCredentials(profile, filename string) (Value, error) {
	var value Value
	var iniErr, jsonErr error

	if value, iniErr = p.loadCredentialsINI(profile, filename); iniErr != nil {
		if value, jsonErr = p.loadCredentialsJSON(profile, filename); jsonErr != nil {
			return value, fmt.Errorf("%v: %v", ErrFileCredentialsLoadFailed, iniErr)
		}
	}

	if len(value.Token) == 0 {
		return value, ErrFileCredentialsTokenNotFound
	}

	return value, nil
}

func (p *FileProvider) loadCredentialsINI(profile, filename string) (Value, error) {
	var value Value

	config, err := ini.Load(filename)
	if err != nil {
		return value, err
	}

	section, err := config.GetSection(profile)
	if err != nil {
		return value, err
	}

	value.Token = section.Key("token").String()
	value.Account = section.Key("account").String()

	return value, nil
}

func (p *FileProvider) loadCredentialsJSON(profile, filename string) (Value, error) {
	var value Value

	f, err := os.Open(filename)
	if err != nil {
		return value, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
