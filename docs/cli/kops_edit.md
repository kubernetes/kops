## kops edit

Edit resource

### Synopsis


Edit a resource configuration.
	
This command changes the cloud specification in the registry.

It does not update the cloud resources, to apply the changes use "kops update cluster".

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
* [kops edit cluster](kops_edit_cluster.md)	 - Edit cluster
* [kops edit federation](kops_edit_federation.md)	 - Edit federation
* [kops edit instancegroup](kops_edit_instancegroup.md)	 - Edit instancegroup

