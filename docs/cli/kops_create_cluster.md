## kops create cluster

Create cluster

### Synopsis


Creates a k8s cluster.

```
kops create cluster
```

### Options

```
      --admin-access string         Restrict access to admin endpoints (SSH, HTTPS) to this CIDR.  If not set, access will not be restricted by IP.
      --associate-public-ip         Specify --associate-public-ip=[true|false] to enable/disable association of public IP for master ASG and nodes. Default is 'true'. (default true)
      --channel string              Channel for default versions and configuration to use (default "stable")
      --cloud string                Cloud provider to use - gce, aws
      --dns-zone string             DNS hosted zone to use (defaults to longest matching zone)
      --image string                Image to use
      --kubernetes-version string   Version of kubernetes to run (defaults to version in channel)
      --master-size string          Set instance size for masters
      --master-zones string         Zones in which to run masters (must be an odd number)
      --model string                Models to apply (separate multiple models with commas) (default "config,proto,cloudup")
      --network-cidr string         Set to override the default network CIDR
      --networking string           Networking mode to use.  kubenet (default), classic, external, cni, kopeio-vxlan, weave. (default "kubenet")
      --node-count int              Set the number of nodes
      --node-size string            Set instance size for nodes
      --out string                  Path to write any local output
      --project string              Project to use (must be set on GCE)
      --ssh-public-key string       SSH public key to use (default "~/.ssh/id_rsa.pub")
      --target string               Target - direct, terraform (default "direct")
  -t, --topology string             Controls network topology for the cluster. public|private. Default is 'public'. (default "public")
      --vpc string                  Set to use a shared VPC
      --yes                         Specify --yes to immediately create the cluster
      --zones string                Zones in which to run the cluster
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
* [kops create](kops_create.md)	 - Create a resource by filename or stdin

