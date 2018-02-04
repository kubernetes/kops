package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

type LocalhostExecutor struct {
}

var _ Executor = &LocalhostExecutor{}

func (e *LocalhostExecutor) Close() error {
	return nil
}

func (s *LocalhostExecutor) Mkdir(dest string, mode os.FileMode) error {
	return os.Mkdir(dest, mode)
}

func (s *LocalhostExecutor) Put(dest string, length int, content io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("error opening file %q: %v", dest, err)
	}

	defer func() {
		if f != nil {
			err := f.Close()
			if err != nil {
				glog.Warningf("error closing file %q: %v", dest, err)
			}
		}
	}()

	_, err = io.Copy(f, content)
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", dest, err)
	}

	err = f.Close()
	f = nil // Don't close in defer block
	if err != nil {
		return fmt.Errorf("error closing file %q: %v", dest, err)
	}

	return nil
}

func (s *LocalhostExecutor) Run(cmd *CommandExecution) error {
	return runCommand(cmd, s, func(command []string) ([]byte, error) {
		name := command[0]
		args := []string{}
		if len(command) > 1 {
			args = command[1:]
		}

		output, err := exec.Command(name, args...).CombinedOutput()
		return output, err
	})
}
