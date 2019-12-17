## Changing a cluster configuration

(This procedure is currently unnecessarily convoluted.  Expect it to get streamlined!)

* Edit the cluster spec: `kops edit cluster ${NAME}`

* View the changes you are going to apply `kops update cluster ${NAME}`

* Apply the changes for real `kops update cluster ${NAME} --yes`

* See which nodes need to be restarted `kops rolling-update cluster ${NAME}`

* Apply the rolling-update `kops rolling-update cluster ${NAME} --yes`

