package ulogd

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/weaveworks/weave/common"
)

func waitForExit(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		common.Log.Fatalf("ulogd terminated: %v", err)
	}
	common.Log.Fatal("ulogd terminated normally")
}

func Start() error {
	cmd := exec.Command("/usr/sbin/ulogd", "-v")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	stdout, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)
	go waitForExit(cmd)
	return nil
}
