# kubectl

## Create kubecfg settings for kubectl

`update cluster` will do it automatically after cluster creation.
But we expect that if you're part of a team you might share the KOPS_STATE_STORE, and then you can do
this on different machines instead of having to share kubecfg files)

To create the kubecfg configuration settings for use with kubectl:

```
export KOPS_STATE_STORE=s3://<somes3bucket>
# NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops export kubecfg ${NAME}
```

You can now use kubernetes using the kubectl tool (after allowing a few minutes for the cluster to come up):

```kubectl get nodes```
