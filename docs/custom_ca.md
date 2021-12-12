# Using a custom certificate authority

## Background Info

When deploying a `kops` based Kubernetes cluster, `kops` will generate a Certificate Authority keypair for signing
various certificates. In some cases, you may want to provide your own CA keypair.

### Building a cluster with a custom CA

The following procedure will allow you to override the CA when creating a cluster. For the sake of this example, you have two files
`ca.crt` and `ca.key`. 

>`cluster-name.com` should be the cluster name you put in the `cluster.yaml`

```bash
kops create -f cluster.yaml
kops create keypair kubernetes-ca --primary --cert ca.crt --key ca.key --name cluster-name.com
kops update cluster --yes
```

1. First we create the cluster folder structure in the statestore.
2. Second, we create a keypair with the name `kubernetes-ca` and provide our own values.
3. Last, we run `kops update cluster --yes`, which will generate all the certificates needed, referencing the keypair called `kubernetes-ca` we just defined (instead of generating its own).
