package executor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	sshClient *ssh.Client
}

func NewSSH(client *ssh.Client) Executor {
	return &SSHExecutor{sshClient: client}
}

var _ Executor = &SSHExecutor{}

func (e *SSHExecutor) Close() error {
	return e.sshClient.Close()
}

// SCPMkdir executes a mkdir against the SSH target, using SCP
func (s *SSHExecutor) Mkdir(dest string, mode os.FileMode) error {
	glog.Infof("Doing SSH SCP mkdir: %q", dest)
	session, err := s.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("error establishing SSH session: %v", err)
	}
	defer session.Close()

	name := filepath.Base(dest)
	scpBase := filepath.Dir(dest)
	//scpBase = "." + scpBase

	var stdinErr error
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		_, stdinErr = fmt.Fprintln(w, "D0"+toOctal(mode), 0, name)
		if stdinErr != nil {
			return
		}
	}()
	output, err := session.CombinedOutput("/usr/bin/scp -tr " + scpBase)
	if err != nil {
		glog.Warningf("Error output from SCP: %s", output)
		return fmt.Errorf("error doing SCP mkdir: %v", err)
	}
	if stdinErr != nil {
		glog.Warningf("Error output from SCP: %s", output)
		return fmt.Errorf("error doing SCP mkdir (writing to stdin): %v", stdinErr)
	}

	return nil
}

func toOctal(mode os.FileMode) string {
	return strconv.FormatUint(uint64(mode), 8)
}

// SCPPut copies a file to the SSH target, using SCP
func (s *SSHExecutor) Put(dest string, length int, content io.Reader, mode os.FileMode) error {
	glog.Infof("Doing SSH SCP upload: %q", dest)
	session, err := s.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("error establishing SSH session: %v", err)
	}
	defer session.Close()

	name := filepath.Base(dest)
	scpBase := filepath.Dir(dest)
	//scpBase = "." + scpBase

	var stdinErr error
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		_, stdinErr = fmt.Fprintln(w, "C0"+toOctal(mode), length, name)
		if stdinErr != nil {
			return
		}
		_, stdinErr = io.Copy(w, content)
		if stdinErr != nil {
			return
		}
		_, stdinErr = fmt.Fprint(w, "\x00")
		if stdinErr != nil {
			return
		}
	}()
	output, err := session.CombinedOutput("/usr/bin/scp -tr " + scpBase)
	if err != nil {
		glog.Warningf("Error output from SCP: %s", output)
		return fmt.Errorf("error doing SCP put: %v", err)
	}
	if stdinErr != nil {
		glog.Warningf("Error output from SCP: %s", output)
		return fmt.Errorf("error doing SCP put (writing to stdin): %v", stdinErr)
	}

	return nil
}

func (s *SSHExecutor) Run(cmd *CommandExecution) error {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("error establishing SSH session: %v", err)
	}
	defer session.Close()

	return runCommand(cmd, s, func(command []string) ([]byte, error) {
		output, err := session.CombinedOutput(joinCommand(command))
		return output, err
	})
}
