## kops completion

Output shell completion code for the given shell (bash)

### Synopsis


Output shell completion code for the given shell (bash).

This command prints shell code which must be evaluation to provide interactive
completion of kops commands.

```
kops completion
```

### Examples

```

# load in the kops completion code for bash (depends on the bash-completion framework).
source <(kops completion bash)
```

### Options

```
      --shell string   target shell (bash).
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --config string                    config file (default is $HOME/.kops.yaml)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default false)
      --name string                      Name of cluster
      --state string                     Location of state storage
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [kops](kops.md)	 - kops is kubernetes ops

