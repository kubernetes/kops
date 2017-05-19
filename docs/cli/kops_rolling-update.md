## kops rolling-update

Rolling update a cluster.

### Synopsis


This command updates a Kubernetes cluster to match the cloud, and kops specifications. 

The Examples below include a series of command for Terraform users.  The workflow includes using Terraform. 

Use export KOPS FEATURE FLAGS="+DrainAndValidateRollingUpdate" to use beta code that drains the nodes and validates the cluster.  New flags for Drain and Validation operations will be shown when the environment variable is set. 

Node Replacement Algorithm Alpa Feature 

We are now including three new algorithms that influence node replacement. All masters and bastions are rolled sequentially before the nodes, and this flag does not influence their replacement.  These algorithms utilize the feature flag mentioned above. 

  1. "asg" - A node is drained then deleted.  The cloud then replaces the node automatically. (default)  
  2. "pre-create" - All node instance groups are duplicated first; then all old nodes are cordoned.  
  3. "create" - As each node instance group rolls, the instance group is duplicated, then all old nodes are cordoned.  

The second and third options create new instance groups; next, the old nodes are cardoned. The old nodes are drained, and then the instance group(s) is deleted.

### Examples

```
  # Roll the currently selected kops cluster
  kops rolling-update cluster --yes
  
  # Update instructions for Terraform users
  kops update cluster --target=terraform
  terraform plan
  terraform apply
  kops rolling-update cluster --yes
  
  # Roll the k8s-cluster.example.com kops cluster and use the new drain an validate functionality
  export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --fail-on-validate-error="false" \
  --master-interval=8m \
  --node-interval=8m
  
  # Use the pre-create node algorithm. First all new nodes are created
  # The old nodes are cordon, drained and deleted.
  export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --algorithm pre-create
  
  # Roll the k8s-cluster.example.com kops cluster, and only roll the instancegroup named "foo".
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --fail-on-validate-error="false" \
  --node-interval 8m \
  --instance-group foo
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
* [kops rolling-update cluster](kops_rolling-update_cluster.md)	 - Rolling update a cluster.

