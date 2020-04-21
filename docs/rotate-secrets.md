# How to rotate all secrets / credentials

**This is a disruptive procedure.**

## Delete all secrets

Delete all secrets & keypairs that kops is holding:

```
kops get secrets  | grep ^Secret | awk '{print $2}' | xargs -I {} kops delete secret secret {}

kops get secrets  | grep ^Keypair | awk '{print $2}' | xargs -I {} kops delete secret keypair {}
```

## Recreate all secrets

Now run `kops update` to regenerate the secrets & keypairs.
```
kops update cluster
kops update cluster --yes
```

Kops may fail to recreate all the keys on first try. If you get errors about ca key for 'ca' not being found, run `kops update cluster --yes` once more.

## Force cluster to use new secrets

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

## Recreate all service accounts

Now the service account tokens will need to be regenerated inside the cluster:

`kops toolbox dump` and find a master IP

Then `ssh admin@${IP}` and run this to delete all the service account tokens:

```
# Delete all service account tokens in all namespaces
NS=`kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}'`
for i in ${NS}; do kubectl get secrets --namespace=${i} --no-headers | grep "kubernetes.io/service-account-token" | awk '{print $1}' | xargs -I {} kubectl delete secret --namespace=$i {}; done

# Allow for new secrets to be created
sleep 60

# Bounce all pods to make use of the new service tokens
pkill -f kube-controller-manager
kubectl delete pods --all --all-namespaces
```

## Verify the cluster is back up

The last command from the previous section will take some time. Meanwhile you can check validation to see the cluster gradually coming back online.

```
kops validate cluster
```
