# How to rotate all secrets / credentials

This is a disruptive procedure.

Delete all secrets & keypairs that kops is holding:

```
kops get secrets  | grep ^Secret | awk '{print $2}' | xargs -I {} kops delete secret secret {}

kops get secrets  | grep ^Keypair | awk '{print $2}' | xargs -I {} kops delete secret keypair {}
```

Now run `kops update cluster` and `kops update cluster --yes` to regenerate the secrets & keypairs.

We need to reboot every node (using a rolling-update).  We have to use `--cloudonly` because our keypair no longer matches.
We set the interval small because nodes will stop trusting each other during the process, so there is no point in going slowly.

`kops rolling-update cluster --cloudonly --master-interval=10s --node-interval=10s --force --yes`

Re-export kubecfg with new settings:

`kops export kubecfg`

Now the service account tokens will need to be regenerated inside the cluster:

`kops toolbox dump` and find a master IP

Then `ssh admin@${IP}` and run this to delete all the service account tokens:

```
# Delete all service account tokens in all namespaces
NS=`kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}'`
for i in ${NS}; do kubectl get secrets --namespace=${i} --no-headers | grep "kubernetes.io/service-account-token" | awk '{print $1}' | xargs -I {} kubectl delete secret --namespace=$i {}; done

# Allow for new secrets to be created
sleep 60

# Bounce pods that we know use service account tokens - you will likely have to bounce more
kubectl delete pods -lk8s-app=dns-controller --namespace=kube-system
kubectl delete pods -lk8s-app=kube-dns --namespace=kube-system
kubectl delete pods -lk8s-app=kube-dns-autoscaler --namespace=kube-system
pkill -f kube-controller-manager
```
