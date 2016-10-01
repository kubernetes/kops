package main

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"os"
)

var (
	// value overwritten during build. This can be used to resolve issues.
	BuildVersion = cloudup.NodeUpVersion
)

func main() {
	Execute()
}

// exitWithError will terminate execution with an error result
// It prints the error to stderr and exits with a non-zero exit code
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "\n%v\n", err)
	os.Exit(1)
}
