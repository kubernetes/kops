## kops toolbox convert-imported

Convert an imported cluster into a kops cluster

### Synopsis


Convert an imported cluster into a kops cluster

```
kops toolbox convert-imported
```

### Options

```
      --channel string   Channel to use for upgrade (default "stable")
      --newname string   new cluster name
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --config string                    config file (default is $HOME/.kops.yaml)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default false)
      --name string                      Name of cluster
      --state string                     Location of state storage (default "s3://oscar-ai-k8s-dev")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [kops toolbox](kops_toolbox.md)	 - Misc infrequently used commands.

