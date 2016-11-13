package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	daemon "github.com/tyranron/daemonigo"
)

// Setting up daemon properties.
func init() {
	daemon.AppName = "golang-sizeof.tips HTTP server"
	daemon.PidFile = "logs/sizeof.pid"

	httpPort := ""
	flag.StringVar(
		&httpPort, "http", DefaultHttpPort, "port to listen http reauests on",
	)

	// Overwriting default daemonigo "start" action.
	daemon.SetAction("start", func() {
		switch isRunning, _, err := daemon.Status(); {
		case err != nil:
			printStatusErr(err)
		case isRunning:
			fmt.Printf(
				"%s is already started and running now\n", daemon.AppName,
			)
		default:
			daemonStart(httpPort)
		}
	})

	// Overwriting default daemonigo "restart" action.
	daemon.SetAction("restart", func() {
		isRunning, process, err := daemon.Status()
		if err != nil {
			printStatusErr(err)
			return
		}
		if isRunning {
			fmt.Printf("Stopping %s...", daemon.AppName)
			if err := daemon.Stop(process); err != nil {
				printFailed(err)
				return
			} else {
				fmt.Println("OK")
			}
		}
		daemonStart(httpPort)
	})
}

// Helper function for custom daemon starting.
func daemonStart(port string) {
	fmt.Printf("Starting %s...", daemon.AppName)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGUSR1)
	cmd, err := daemon.StartCommand()
	if err != nil {
		printFailed(err)
		return
	}
	if port != "" {
		cmd.Env = append(cmd.Env, "_GO_HTTP="+port)
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		printFailed(err)
		return
	}
	if err = cmd.Start(); err != nil {
		printFailed(err)
		return
	}
	select {
	case <-sig: // received "OK" signal from child process
		fmt.Println("OK")
	case <-time.After(10 * time.Second): // timeout for waiting signal
		fmt.Println("TIMEOUTED")
		fmt.Println("Reason: signal from child process not received")
		fmt.Println("Details: check logs/application.log for details")
	case err := <-func() chan error {
		ch := make(chan error)
		go func() {
			msg, _ := ioutil.ReadAll(stdErr)
			err := cmd.Wait()
			if err == nil {
				err = fmt.Errorf("child process unexpectedly stopped")
			}
			if len(msg) > 0 {
				err = fmt.Errorf("%s\nDetails: %s", err.Error(), msg)
			}
			ch <- err
		}()
		return ch
	}(): // child process unexpectedly stopped without sending signal
		printFailed(err)
	}
}

// Helper function for printing error that occurred during
// daemon status checking.
func printStatusErr(e error) {
	fmt.Println("Checking status of " + daemon.AppName + " failed")
	fmt.Println("Details:", e.Error())
}

// Helper function for printing failures of daemon actions.
func printFailed(e error) {
	fmt.Println("FAILED")
	fmt.Println("Details:", e.Error())
}
