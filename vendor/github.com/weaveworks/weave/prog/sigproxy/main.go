package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("USAGE: sigproxy <command> [arguments ...]")
	}

	// Install signal handler as soon as possible - channel is buffered so
	// we'll catch signals that arrive whilst child process is starting
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	cmd := exec.Command(os.Args[1], os.Args[2:]...)

	// These default to /dev/null, so set them explicitly to ours
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Only begin delivering signals after the child has started
	go func() {
		for {
			// Signalling PID 0 delivers to our process group
			syscall.Kill(0, (<-sc).(syscall.Signal))
		}
	}()

	if err := cmd.Wait(); err != nil {
		// Exit status is platform specific so not directly accessible - casts
		// required to access system-dependent exit information
		if exitErr, ok := err.(*exec.ExitError); ok {
			waitStatus := exitErr.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		}
		os.Exit(1)
	}
	os.Exit(0)
}
