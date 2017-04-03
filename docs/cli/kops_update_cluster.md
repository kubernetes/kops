## kops update cluster

Update cluster

### Synopsis


Updates a k8s cluster.

```
kops update cluster
```

### Options

```
      --create-kube-config      Will control automatically creating the kube config file on your local filesystem (default true)
      --model string            Models to apply (separate multiple models with commas) (default "config,proto,cloudup")
      --out string              Path to write any local output
      --ssh-public-key string   SSH public key to use (deprecated: use kops create secret instead)
      --target string           Target - direct, terraform (default "direct")
      --yes                     Actually create cloud resources
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
* [kops update](kops_update.md)	 - update clusters

