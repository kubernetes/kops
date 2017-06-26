## kops rolling-update cluster

Rolling update a cluster.

### Synopsis


This command updates a kubernetes cluster to match the cloud, and kops specifications. 

To perform rolling update, you need to update the cloud resources first with "kops update cluster" 

Note: terraform users will need run the following commands all from the same directory "kops update cluster --target=terraform" then "terraform plan" then "terraform apply" prior to running "kops rolling-update cluster" 

Use export KOPS FEATURE FLAGS="+DrainAndValidateRollingUpdate" to use beta code that drains the nodes and validates the cluster.  New flags for Drain and Validation operations will be shown when the environment variable is set. 

Node Replacement Strategies Alpha Feature 

We are now including three new strategies that influence node replacement. All masters and bastions are rolled sequentially before the nodes, and this flag does not influence their replacement.  These strategies utilize the feature flag mentioned above. 

  1. "default" - A node is drained then deleted.  The cloud then replaces the node automatically.  
  2. "create-all-new-ig-first" - All node instance groups are duplicated first; then all old nodes are cordoned.  
  3. "create-new-by-ig" - As each node instance group rolls, the instance group is duplicated, then all old nodes are cordoned.  

The second and third options create new instance groups. In order to use this ALPHA feature you need to enable +DrainAndValidateRollingUpdate,+RollingUpdateStrategies feature flags.

```
kops rolling-update cluster
```

### Examples

```
  # Roll the currently selected kops cluster
  kops rolling-update cluster --yes
  
  # Roll the k8s-cluster.example.com kops cluster
  # use the new drain an validate functionality
  export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --fail-on-validate-error="false" \
  --master-interval=8m \
  --node-interval=8m
  
  
  # Roll the k8s-cluster.example.com kops cluster
  # only roll the node instancegroup
  # use the new drain an validate functionality
  export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --fail-on-validate-error="false" \
  --node-interval 8m \
  --instance-group nodes
  
  # Roll the k8s-cluster.example.com kops cluster, and only roll the instancegroup named "foo".
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --fail-on-validate-error="false" \
  --node-interval 8m \
  --instance-group foo
  
  # Use the create-new-by-ig node strategy. Master(s) are update in series, and then
  # each instance groups is updated in a loop. A new instance group is created, cluster is validated,
  # and then the old nodes are cordon, drained and deleted. This process is repeated
  # for every node instance group.
  export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate,+RollingUpdateStrategies"
  kops rolling-update cluster k8s-cluster.example.com --yes \
  --strategy create-new-by-ig --master-interval=8m \
  --node-interval=8m
```

### Options

```
      --bastion-interval duration    Time to wait between restarting bastions (default 5m0s)
      --cloudonly                    Perform rolling update without confirming progress with k8s
      --force                        Force rolling update, even if no changes
      --instance-group stringSlice   List of instance groups to update (defaults to all if not specified)
      --master-interval duration     Time to wait between restarting masters (default 5m0s)
      --node-interval duration       Time to wait between restarting nodes (default 2m0s)
  -y, --yes                          perform rolling update without confirmation
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
* [kops rolling-update](kops_rolling-update.md)	 - Rolling update a cluster.

