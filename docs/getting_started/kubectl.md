# kubectl cluster admin configuration

When you run `kops create cluster --yes`, you automatically get a kubectl configuration for accessing the cluster. This configuration gives you full admin access to the cluster for 18 hours.

If you want to create this configuration on other machine, you can run the following as long as you have access to the kOps state store:

```
export KOPS_STATE_STORE=<location of the kops state store>
NAME=<kubernetes.mydomain.com>
kops export kubeconfig ${NAME} --admin
```

Warning: Note that the exported configuration gives you full admin privileges for 18 hours. For regular kubectl usage, you should consider using another method for authenticating to the cluster.

To create the kubeconfig configuration settings without the admin credential:

```
export KOPS_STATE_STORE=<location of the kops state store>
NAME=<kubernetes.mydomain.com>
kops export kubeconfig ${NAME}
```
