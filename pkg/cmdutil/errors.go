package cmdutil

import (
	"fmt"
	"os"
)

func CheckErr(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Unexpected error: %v", err)
	os.Exit(1)
}
