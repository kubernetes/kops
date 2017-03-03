## Changing a cluster configuration

(This procedure is currently unnecessarily convoluted.  Expect it to get streamlined!)

* Edit the cluster spec: `kops edit cluster ${NAME}`

* View the changes you are going to apply `kops update cluster ${NAME}`

* Apply the changes for real `kops update cluster ${NAME} --yes`

* See which nodes need to be restarted `kops rolling-update cluster ${NAME}`

* Apply the rolling-update `kops rolling-update cluster ${NAME} --yes`

NOTE: rolling-update does not yet perform a real rolling update - it just shuts down machines in sequence with a delay;
 there will be downtime [Issue #37](https://github.com/kubernetes/kops/issues/37)
We have implemented a new feature that does drain and validate nodes.  This feature is experimental, and you can use the new feature by setting `export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"`.

