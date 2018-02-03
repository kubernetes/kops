package imagebuilder

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// ReadFile reads the whole file using ioutil.ReadFile, but does path expansion first
func ReadFile(p string) ([]byte, error) {
	if strings.HasPrefix(p, "~/") {
		p = os.Getenv("HOME") + p[1:]
	}
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", p, err)
	}
	return data, nil
}
