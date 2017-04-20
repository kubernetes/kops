package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var (
	ErrNoCommandSpecified = errors.New("No command specified")
)

func main() {
	var (
		args = os.Args[1:]
	)

	checkErr(checkNetwork())

	if len(args) == 0 {
		checkErr(ErrNoCommandSpecified)
	}

	binary, err := exec.LookPath(args[0])
	checkErr(err)

	checkErr(syscall.Exec(binary, args, os.Environ()))
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
