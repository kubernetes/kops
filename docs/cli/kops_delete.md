## kops delete

Delete clusters,instancegroups, or secrets.

### Synopsis


Delete clusters, instancegroups, or secrets.

```
kops delete -f FILENAME [--yes]
```

### Examples

```
  # Create a cluster using a file
  kops delete -f my-cluster.yaml
  
  # Delete a cluster in AWS.
  kops delete cluster --name=k8s.example.com --state=s3://kops-state-1234
```

### Options

```
  -f, --filename stringSlice   Filename to use to delete the resource
  -y, --yes                    Specify --yes to delete the resource
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
* [kops](kops.md)	 - kops is Kubernetes ops.
* [kops delete cluster](kops_delete_cluster.md)	 - Delete a cluster.
* [kops delete instancegroup](kops_delete_instancegroup.md)	 - Delete instancegroup
* [kops delete secret](kops_delete_secret.md)	 - Delete a secret

