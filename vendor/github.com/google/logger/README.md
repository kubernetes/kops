# logger #
Logger is a simple cross platform Go logging library for Windows, Linux, FreeBSD, and
macOS, it can log to the Windows event log, Linux/macOS syslog, and an io.Writer.

This is not an official Google product.

## Usage ##

Set up the default logger to log the system log (event log or syslog) and a
file, include a flag to turn up verbosity:

```go
import (
  "flag"
  "os"

  "github.com/google/logger"
)

const logPath = "/some/location/example.log"

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

func main() {
  flag.Parse()

  lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
  if err != nil {
    logger.Fatalf("Failed to open log file: %v", err)
  }
  defer lf.Close()

  defer logger.Init("LoggerExample", *verbose, true, lf).Close()

  logger.Info("I'm about to do something!")
  if err := doSomething(); err != nil {
    logger.Errorf("Error running doSomething: %v", err)
  }
}
```

The Init function returns a logger so you can setup multiple instances if you
wish, only the first call to Init will set the default logger:

```go
lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
if err != nil {
  logger.Fatalf("Failed to open log file: %v", err)
}
defer lf.Close()

// Log to system log and a log file, Info logs don't write to stdout.
loggerOne := logger.Init("LoggerExample", false, true, lf)
defer loggerOne.Close()
// Don't to system log or a log file, Info logs write to stdout..
loggerTwo := logger.Init("LoggerExample", true, false, ioutil.Discard)
defer loggerTwo.Close()

loggerOne.Info("This will log to the log file and the system log")
loggerTwo.Info("This will only log to stdout")
logger.Info("This is the same as using loggerOne")

```

## Custom Format ##

| Code                                 | Example                                                  |
|--------------------------------------|----------------------------------------------------------|
| `logger.SetFlags(log.Ldate)`         | ERROR: 2018/11/11 Error running Foobar: message          |
| `logger.SetFlags(log.Ltime)`         | ERROR: 09:42:45 Error running Foobar: message            |
| `logger.SetFlags(log.Lmicroseconds)` | ERROR: 09:42:50.776015 Error running Foobar: message     |
| `logger.SetFlags(log.Llongfile)`     | ERROR: /src/main.go:31: Error running Foobar: message    |
| `logger.SetFlags(log.Lshortfile)`    | ERROR: main.go:31: Error running Foobar: message         |
| `logger.SetFlags(log.LUTC)`          | ERROR: Error running Foobar: message                     |
| `logger.SetFlags(log.LstdFlags)`     | ERROR: 2018/11/11 09:43:12 Error running Foobar: message |

```go
func main() {
    lf, err := os.OpenFile(logPath, …, 0660)
    defer logger.Init("foo", *verbose, true, lf).Close()
    logger.SetFlags(log.LstdFlags)
}
```

More info: https://golang.org/pkg/log/#pkg-constants
