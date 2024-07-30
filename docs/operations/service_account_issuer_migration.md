# Service Account Issuer (SAI) migration

In the past changing the Service Account Issuer has been a disruptive process. However since Kubernetes v1.22 you can specify multiple Service Account Issuers in the Kubernetes API Server ([Docs here](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection)).

As noted in the Kubernetes Docs when the `--service-account-issuer` flag is specified multiple times, the first is used to generate tokens and all are used to determine which issuers are accepted.

So with this feature we can migrate to a new Service Account Issuer without disruption to cluster operations.

**Note**: These procedures can be adapted to enable IAM Roles for Service Accounts (IRSA) without disruption to cluster operations.

## Migrate using kOps `additionalServiceAccountIssuers`

In [kubernetes/kops#16497](https://github.com/kubernetes/kops/pull/16497) kOps added support for specifying an additional Service Account Issuer. This was released in kOps v1.28.5 so it requires at least this version to follow this procedure. If you are using an earlier version you can follow the [Migrate using Instancegroup Hooks](#migrate-using-instancegroup-hooks) procedure.

**Warning**: This procedure is manual. We recommend testing this on a staging cluster before proceeding on a production cluster.

In this example we are switching from `master.[cluster-name].[domain]` to `api.internal.[cluster-name].[domain]`.

1. Add new SAI as additional (existing SAI as primary) to the Cluster resource
```yaml
  kubeAPIServer:
    serviceAccountIssuer: https://master.[cluster-name].[domain]
    additionalServiceAccountIssuers:
      - https://api.internal.[cluster-name].[domain]
```
2. Apply the changes to the cluster
3. Roll the control-plane nodes
4. Switch the primary/secondary SAI on the Cluster resource
```yaml
  kubeAPIServer:
    serviceAccountIssuer: https://api.internal.[cluster-name].[domain]
    additionalServiceAccountIssuers:
      - https://master.[cluster-name].[domain]
```
5. Apply the changes to the cluster
6. Roll the control-plane nodes
7. Wait 24 hours until the dynamic SA tokens have refreshed
8. Remove the `additionalServiceAccountIssuers` from the Cluster resource
9. Apply the changes to the cluster
10. Roll the control-plane nodes

## Migrate using Instancegroup Hooks

**Warning**: This procedure is manual and involves some tricky modification of manifest files. We recommend testing this on a staging cluster before proceeding on a production cluster.

In this example we are switching from `master.[cluster-name].[domain]` to `api.internal.[cluster-name].[domain]`.

1. Add the `modify-kube-api-manifest` (existing SAI as primary) hook to the control-plane InstanceGroup resources
```yaml
  hooks:
  - name: modify-kube-api-manifest
    before:
      - kubelet.service
    manifest: |
      User=root
      Type=oneshot
      ExecStart=/bin/bash -c "until [ -f /etc/kubernetes/manifests/kube-apiserver.manifest ];do sleep 5;done;sed -i '/- --service-account-issuer=https:\/\/api.internal.[cluster-name].[domain]/i\ \ \ \ - --service-account-issuer=https:\/\/master.[cluster-name].[domain]' /etc/kubernetes/manifests/kube-apiserver.manifest"
```
2. Apply the changes to the cluster
3. Roll the control-plane nodes
4. Update the `modify-kube-api-manifest` (switch the primary/secondary SAI) hook on the control-plane InstanceGroup resources
```yaml
  hooks:
  - name: modify-kube-api-manifest
    before:
      - kubelet.service
    manifest: |
      User=root
      Type=oneshot
      ExecStart=/bin/bash -c "until [ -f /etc/kubernetes/manifests/kube-apiserver.manifest ];do sleep 5;done;sed -i '/- --service-account-issuer=https:\/\/api.internal.[cluster-name].[domain]/a\ \ \ \ - --service-account-issuer=https:\/\/master.[cluster-name].[domain]' /etc/kubernetes/manifests/kube-apiserver.manifest"
```
5. Apply the changes to the cluster
6. Roll the control-plane nodes
7. Wait 24 hours until the dynamic SA tokens have refreshed
8. Remove the `modify-kube-api-manifest` hook on the control-plane InstanceGroup resources
9. Apply the changes to the cluster
10. Roll the control-plane nodes

This procedure was originally posted in a GitHub issue [here](https://github.com/kubernetes/kops/issues/16488#issuecomment-2084325891) with inspiration from [this comment](https://github.com/kubernetes/kops/issues/14201#issuecomment-1732035655).
