## Changing a cluster configuration

(This procedure is currently unnecessarily convoluted.  Expect it to get streamlined!)

Edit the cluster spec: `kops edit cluster --name ${NAME}`

View the changes you are going to apply `kops create cluster --name ${DRYRUN} --dryrun`

Apply the changes for real `kops create cluster --name ${DRYRUN}`

See which nodes need to be restarted `kops rolling-update cluster --name ${NAME} --region ${REGION}`

Apply the rolling-update `kops rolling-update cluster --name ${NAME} --region ${REGION} --yes`

