# kubectl cluster admin configuration

When you run `kops update cluster` during cluster creation, you automatically get a kubectl configuration for accessing the cluster. This configuration gives you full admin access to the cluster.
If you want to create this configuration on other machine, you can run the following as long as you have access to the kops state store.

To create the kubecfg configuration settings for use with kubectl:

```
export KOPS_STATE_STORE=<location of the kops state store>
NAME=<kubernetes.mydomain.com>
kops export kubecfg ${NAME}
```

Warning: Note that the exported configuration gives you full admin privileges using TLS certificates that are not easy to rotate. For regular kubectl usage, you should consider using another method for authenticating to the cluster.