# Using a custom certificate authority

## Background Info

When deploying a `kops` based Kubernetes cluster, `kops` will generate a certificate authority keypair for signing
various certificates with. In some cases, you may want to provide your own CA keypair.

Another use case would be to use the CA keypair of another cluster if you are creating many
short lived clusters and don't want to create a unique CA for each one.

### Building a cluster with a custom CA

The following procedure will allow you to override the CA when creating a cluster. For the sake of this example, you have two files
`ca.crt` and `ca.key`. 

>`cluster-name.com` should be the cluster name you put in the `cluster.yaml`

```bash
kops create -f cluster.yaml
kops create secret keypair ca --cert ca.crt --key ca.key --name cluster-name.com
kops update cluster --yes
```

1. First we create the cluster folder structure in the statestore.
2. Second, we create a `Secret` of type `Keypair` with the name `ca` and provide our own values.
3. Lastly, we run `kops update cluster --yes`, which will generate all the certificates needed, referencing the `Secret` called `ca` we just defined (versus generating its own).

### Using a previous `kops` cluster CA

In some cases you will want to create a cluster and use the CA generated in a previous `kops` cluster.
To do so, you will need to copy the CA files from the state store, and then use them as values in the above procedure.

The files are located as follows:

`s3://state-store/<cluster-name>/pki/issued/ca/<id>.crt`

`s3://state-store/<cluster-name>/pki/private/ca/<id>.key`
