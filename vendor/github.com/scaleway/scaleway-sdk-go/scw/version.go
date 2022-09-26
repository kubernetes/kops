package scw

import (
	"fmt"
	"runtime"
)

// TODO: versioning process
const version = "v1.0.0-beta.7+dev"

var userAgent = fmt.Sprintf("scaleway-sdk-go/%s (%s; %s; %s)", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
