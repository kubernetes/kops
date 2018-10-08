package envy

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var gil = &sync.Mutex{}
var env = map[string]string{}

func init() {
	Load()
	loadEnv()
}

// Load the ENV variables to the env map
func loadEnv() {
	// Detect the Go version on the user system, not the one that was used to compile the binary
	v := ""
	out, err := exec.Command("go", "version").Output()
	if err == nil {
		// This will break when Go 2 lands
		v = strings.Split(string(out), " ")[2][4:]
	} else {
		v = runtime.Version()[4:]
	}

	goRuntimeVersion, _ := strconv.ParseFloat(runtime.Version()[4:], 64)

	goVersion, err := strconv.ParseFloat(v, 64)
	if err != nil {
		goVersion = goRuntimeVersion
	}

	if os.Getenv("GO_ENV") == "" {
		if flag.Lookup("test.v") != nil {
			Set("GO_ENV", "test")
		}
	}

	// set the GOPATH if using >= 1.8 and the GOPATH isn't set
	if goVersion >= 8 && os.Getenv("GOPATH") == "" {
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err == nil {
			gp := strings.TrimSpace(string(out))
			os.Setenv("GOPATH", gp)
		}
	}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env[pair[0]] = os.Getenv(pair[0])
	}
}

// Reload the ENV variables. Useful if
// an external ENV manager has been used
func Reload() {
	env = map[string]string{}
	loadEnv()
}

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
func Load(files ...string) error {

	// If no files received, load the default one
	if len(files) == 0 {
		err := godotenv.Overload()
		if err == nil {
			Reload()
		}
		return err
	}

	// We received a list of files
	for _, file := range files {

		// Check if it exists or we can access
		if _, err := os.Stat(file); err != nil {
			// It does not exist or we can not access.
			// Return and stop loading
			return err
		}

		// It exists and we have permission. Load it
		if err := godotenv.Overload(file); err != nil {
			return err
		}

		// Reload the env so all new changes are noticed
		Reload()

	}
	return nil
}

// Get a value from the ENV. If it doesn't exist the
// default value will be returned.
func Get(key string, value string) string {
	gil.Lock()
	defer gil.Unlock()
	if v, ok := env[key]; ok {
		return v
	}
	return value
}

// Get a value from the ENV. If it doesn't exist
// an error will be returned
func MustGet(key string) (string, error) {
	gil.Lock()
	defer gil.Unlock()
	if v, ok := env[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("could not find ENV var with %s", key)
}

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
func Set(key string, value string) {
	gil.Lock()
	defer gil.Unlock()
	env[key] = value
}

// MustSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
func MustSet(key string, value string) error {
	gil.Lock()
	defer gil.Unlock()
	err := os.Setenv(key, value)
	if err != nil {
		return err
	}
	env[key] = value
	return nil
}

// Map all of the keys/values set in envy.
func Map() map[string]string {
	gil.Lock()
	defer gil.Unlock()
	return env
}

// Temp makes a copy of the values and allows operation on
// those values temporarily during the run of the function.
// At the end of the function run the copy is discarded and
// the original values are replaced. This is useful for testing.
func Temp(f func()) {
	oenv := env
	env = map[string]string{}
	for k, v := range oenv {
		env[k] = v
	}
	defer func() { env = oenv }()
	f()
}

func GoPath() string {
	return Get("GOPATH", "")
}

// GoPaths returns all possible GOPATHS that are set.
func GoPaths() []string {
	gp := Get("GOPATH", "")
	if runtime.GOOS == "windows" {
		return strings.Split(gp, ";") // Windows uses a different separator
	}
	return strings.Split(gp, ":")
}

func importPath(path string) string {
	for _, gopath := range GoPaths() {
		srcpath := filepath.Join(gopath, "src")
		rel, err := filepath.Rel(srcpath, path)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}

	// fallback to trim
	rel := strings.TrimPrefix(path, filepath.Join(GoPath(), "src"))
	rel = strings.TrimPrefix(rel, string(filepath.Separator))
	return filepath.ToSlash(rel)
}

func CurrentPackage() string {
	pwd, _ := os.Getwd()
	return importPath(pwd)
}

func Environ() []string {
	gil.Lock()
	defer gil.Unlock()
	var e []string
	for k, v := range env {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	return e
}
