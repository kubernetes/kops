
# How to rotate all secrets / credentials

There are two types of credentials managed by kOps:

* "secrets" are symmetric credentials.

* "keypairs" are pairs of X.509 certificates and their corresponding private keys.
  The exceptions are "service-account" keypairs, which are stored as
  certificate and private key pairs, but do not use any part of the certificates
  other than the public keys.

  Keypairs are grouped into named "keysets", according to their use. For example,
  the "kubernetes-ca" keyset is used for the cluster's Kubernetes general CA.
  Each keyset has a single primary keypair, which is the one whose private key
  is used. The remaining, secondary keypairs are either trusted or distrusted.
  The trusted keypairs, including the primary keypair, have their certificates
  included in relevant trust stores.

## Rotating keypairs

{{ kops_feature_table(kops_added_default='1.22') }}

You may gracefully rotate keypairs of keysets that are either Certificate Authorities
or are "service-account" by performing the following procedure. Other keypairs will be
automatically reissued by a non-dryrun `kops update cluster` when their issuing
CA is rotated.

### Create and stage new keypair

Create a new keypair for each keyset that you are going to rotate.
Then update the cluster and perform a rolling update.
To stage all rotatable keysets, run:

```shell
kops create keypair all
kops update cluster --yes
kops rolling-update cluster --yes
```

#### Rollback procedure

A failure at this stage is unlikely. To roll back this change:

* Use `kops get keypairs` to get the IDs of the newly created keysets.
* Then use `kops distrust keypair` to distrust each of them by keyset and ID.
* Then use `kops update cluster --yes`
* Then use `kops rolling-update cluster --yes`

### Export and distribute new kubeconfig certificate-authority-data

If you are rotating the Kubernetes general CA ("kubernetes-ca" or "all") and
you are not using a load balancer for the Kubernetes API with its own separate
certificate, export a new kubeconfig with the new CA certificate
included in the `certificate-authority-data` field for the cluster:

```shell
kops export kubecfg
```

Distribute the new `certificate-authority-data` to all clients of that cluster's
Kubernetes API.

#### Rollback procedure

To roll back this change, distribute the previous kubeconfig `certificate-authority-data`.

### Promote the new keypairs

Promote the new keypairs to primary with:

```shell
kops promote keypair all
kops update cluster --yes
kops rolling-update cluster --yes
```

On cloud providers, such as AWS, that use kops-controller to bootstrap worker nodes, after
the `kops update cluster --yes` step there is a temporary impediment to node scale-up.
Instances using the new launch template will not be able to bootstrap off of old kops-controllers.
Similarly, instances using the old launch template and which have not yet bootstrapped will not
be able to bootstrap off of new kops-controllers. The subsequent rolling update will eventually
replace all instances using the old launch template.

#### Rollback procedure

The most likely failure at this stage would be a client of the Kubernetes API that
did not get the new `certificate-authority-data` and thus do not trust the
new TLS server certificate.

To roll back this change:

* Use `kops get keypairs` to get the IDs of the previous primary keysets,
  most likely by identifying the issue dates.
* Then use `kops promote keypair` to promote each of them by keyset and ID.
* Then use `kops update cluster --yes`
* Then use `kops rolling-update cluster --yes`

### Export and distribute new kubeconfig admin credentials

If you are rotating the Kubernetes general CA ("kubernetes-ca" or "all") and
have kubeconfigs with cluster admin credentials, export new kubeconfigs
with new admin credentials for the cluster:

```shell
kops export kubecfg --admin=DURATION
```

where `DURATION` is the desired lifetime of the admin credential.

Distribute the new credentials to all clients that require them.

#### Rollback procedure

To roll back this change, distribute the previous kubeconfig admin credentials.

### Distrust the previous keypairs

Remove trust in the previous keypairs with:

```shell
kops distrust keypair all
kops update cluster --yes
kops rolling-update cluster --yes
```

#### Rollback procedure

The most likely failure at this stage would be a client of the Kubernetes API that
is still using a credential issued by the previous keypair.

To roll back this change:

* Use `kops get keypairs --distrusted` to get the IDs of the previously trusted keysets,
  most likely by identifying the distrust dates.
* Then use `kops trust keypair` to trust each of them by keyset and ID.
* Then use `kops update cluster --yes`
* Then use `kops rolling-update cluster --force --yes`

### Export and distribute new kubeconfig certificate-authority-data

If you are rotating the Kubernetes general CA ("kubernetes-ca" or "all") and
you are not using a load balancer for the Kubernetes API with its own separate
certificate, export a new kubeconfig with the previous CA certificate
removed from the `certificate-authority-data` field for the cluster:

```shell
kops export kubecfg
```

Distribute the new `certificate-authority-data` to all clients of that cluster's
Kubernetes API.

#### Rollback procedure

To roll back this change, distribute the previous kubeconfig `certificate-authority-data`.

## Rotating the API Server encryptionconfig

See [the Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#rotating-a-decryption-key)
for information on how to gracefully rotate keys in the encryptionconfig.

Use `kops create secret encryptionconfig --force` to update the encryptionconfig secret.
Following that, use `kops update cluster --yes` and `kops rolling-update cluster --yes`.

## Rotating the Cilium IPSec keys

See the Cilium documentation for information on how to gracefully rotate the Cilium IPSec keys.

Use `kops create secret ciliumpassword --force` to update the cilium-ipsec-keys secret.
Following that, use `kops update cluster --yes` and `kops rolling-update cluster --yes`.

## Rotating the Docker secret

[TODO]

Use `kops create secret dockerconfig --force` to update the Docker secret.
Following that, use `kops update cluster --yes` and `kops rolling-update cluster --yes`.

## Rotating the Weave password

It is not possible to rotate the Weave password without a disruptive partition of the Weave network.
As of the writing of this document, this is a limitation of Weave itself.

Use `kops create secret weavepassword --force` to update the Docker secret.
Following that, use `kops update cluster --yes` and `kops rolling-update cluster --cloudonly --yes`.

## Legacy procedure

The following is the procedure to rotate secrets and keypairs in kOps versions
prior to 1.22.

**This is a disruptive procedure.**

### Delete all secrets

Delete all secrets & keypairs that kOps is holding:

```shell
kops get secrets  | grep '^Secret' | awk '{print $2}' | xargs -I {} kops delete secret secret {}

kops get secrets  | grep '^Keypair' | awk '{print $2}' | xargs -I {} kops delete secret keypair {}
```

### Recreate all secrets

Now run `kops update` to regenerate the secrets & keypairs.
```
kops update cluster
kops update cluster --yes
```

kOps may fail to recreate all the keys on first try. If you get errors about ca key for 'ca' not being found, run `kops update cluster --yes` once more.

### Force cluster to use new secrets

Now you will have to remove the etcd certificates from every master.

Find all the master IPs. One easy way of doing that is running

```
kops toolbox dump
```

Then SSH into each node and run

```
sudo find /mnt/ -name server.* | xargs -I {} sudo rm {}
sudo find /mnt/ -name me.* | xargs -I {} sudo rm {}
```

You need to reboot every node (using a rolling-update). You have to use `--cloudonly` because the keypair no longer matches.

```
kops rolling-update cluster --cloudonly --force --yes
```

Re-export kubecfg with new settings:

```
kops export kubecfg
```

### Recreate all service accounts

Now the service account tokens will need to be regenerated inside the cluster:

`kops toolbox dump` and find a master IP

Then `ssh admin@${IP}` and run this to delete all the service account tokens:

```shell
# Delete all service account tokens in all namespaces
NS=`kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}'`
for i in ${NS}; do kubectl get secrets --namespace=${i} --no-headers | grep "kubernetes.io/service-account-token" | awk '{print $1}' | xargs -I {} kubectl delete secret --namespace=$i {}; done

# Allow for new secrets to be created
sleep 60

# Bounce all pods to make use of the new service tokens
pkill -f kube-controller-manager
kubectl delete pods --all --all-namespaces
```

### Verify the cluster is back up

The last command from the previous section will take some time. Meanwhile you can check validation to see the cluster gradually coming back online.

```
kops validate cluster --wait 10m
```
