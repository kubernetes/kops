package app

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/gophergala/golang-sizeof.tips/internal/log"

	daemon "github.com/tyranron/daemonigo"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Run() (exitCode int) {
	switch isDaemon, err := daemon.Daemonize(); {
	case !isDaemon:
		return
	case err != nil:
		log.StdErr("could not start daemon, reason -> %s", err.Error())
		return 1
	}

	var err error
	appLog, err = log.NewApplicationLogger()
	if err != nil {
		log.StdErr("could not create access log, reason -> %s", err.Error())
		return 1
	}

	if err = prepareTemplates(); err != nil {
		log.StdErr("could not parse html templates, reason -> %s", err.Error())
		return 1
	}

	httpPort := os.Getenv("_GO_HTTP")
	if httpPort == "" {
		httpPort = DefaultHttpPort
	}

	bindHttpHandlers()
	canExit, httpErr := make(chan sig, 1), make(chan error, 1)
	go func() {
		defer close(canExit)
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			httpErr <- fmt.Errorf(
				"creating HTTP server on port '%s' FAILED, reason -> %s",
				httpPort, err.Error(),
			)
		}
	}()
	select {
	case err = <-httpErr:
		appLog.Error(err.Error())
		log.StdErr(err.Error())
		return 1
	case <-time.After(300 * time.Millisecond):
	}

	notifyParentProcess()

	<-canExit
	return
}

// Notifies parent process that everything is OK.
func notifyParentProcess() {
	if err := syscall.Kill(os.Getppid(), syscall.SIGUSR1); err != nil {
		appLog.Error(
			"Notifying parent process FAILED, reason -> %s", err.Error(),
		)
	} else {
		appLog.Info("Notifying parent process SUCCEED")
	}
}
