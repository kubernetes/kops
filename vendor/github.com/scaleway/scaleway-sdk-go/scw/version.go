package scw

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// TODO: versioning process
const (
	defaultVersion = "v1.0.0-beta.7+dev"
	path           = "github.com/scaleway/scaleway-sdk-go"
)

var cachedVersion = (*string)(nil)

func getVersion() string {
	if cachedVersion == nil {
		debugVersion := ""
		b, ok := debug.ReadBuildInfo()
		if ok {
			for _, dep := range b.Deps {
				if dep.Path == path {
					debugVersion = dep.Version
				}
			}
		}

		cachedVersion = &debugVersion
	}

	if *cachedVersion != "" {
		return *cachedVersion
	}

	return defaultVersion
}

var userAgent = fmt.Sprintf("scaleway-sdk-go/%s (%s; %s; %s)", getVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
