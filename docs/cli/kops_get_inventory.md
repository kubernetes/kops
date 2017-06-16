## kops get inventory

Output a list of IoTk - inventory of all things kops. 

### Synopsis


Output a list of IoTk - inventory of all things kops.  Bill of materials (BOM) for a kops installation; containers, binaries, etc.

```
kops get inventory
```

### Examples

```
  # Get a inventory list from a YAML file
  kops get inventory -f k8s.example.com.yaml --state s3://k8s.example.com
  
  # Get a inventory list from a cluster
  kops get inventory k8s.example.com --state s3://k8s.example.com
  
  # Get a inventory list from a cluster as YAML
  kops get inventory k8s.example.com --state s3://k8s.example.com -o YAML
```

### Options

```
      --channel string              Channel for default versions and configuration to use (default "stable")
  -f, --filename stringSlice        Filename to use to create the resource
      --kubernetes-version string   Version of kubernetes to run (defaults to version in channel)
  -o, --output string               output format.  One of: yaml, json, table (default "table")
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
* [kops get](kops_get.md)	 - Get one or many resources.

