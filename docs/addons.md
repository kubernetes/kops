# kOps addons

kOps supports two types of addons:

 * Managed addons, which are configurable through the [cluster spec](cluster_spec.md)
 * Static addons, which are manifest files that are applied as-is

## Managed addons

The following addons are managed by kOps and will be upgraded following the kOps and kubernetes lifecycle, and configured based on your cluster spec. kOps will consider both the configuration of the addon itself as well as what other settings you may have configured where applicable.

### Available addons

#### AWS Load Balancer Controller
{{ kops_feature_table(kops_added_default='1.20') }}

AWS Load Balancer Controller offers additional functionality for provisioning ELBs.

```yaml
spec:
  awsLoadBalancerController:
    enabled: true
```

Though the AWS Load Balancer Controller can integrate the AWS WAF and
Shield services with your Application Load Balancers (ALBs), kOps
disables those capabilities by default.

{{ kops_feature_table(kops_added_default='1.24') }}

You can enable use of either or both of the WAF and WAF Classic
services by including the following fields in the cluster spec:

```yaml
spec:
  awsLoadBalancerController:
    enabled: true
    enableWAF: true
    enableWAFv2: true
```

Note that the controller will only succeed in associating one WAF with
a given ALB at a time, despite it accepting both the
"alb.ingress.kubernetes.io/waf-acl-id" and
"alb.ingress.kubernetes.io/wafv2-acl-arn" annotations on the same
_Ingress_ object.

You can enable use of Shield Advanced by including the following fields in the cluster spec:

```yaml
spec:
  awsLoadBalancerController:
    enabled: true
    enableShield: true
```

Support for the WAF and Shield services in kOps is currently **beta**, meaning
that the accepted configuration and the AWS resources involved may
change.

Read more in the [official documentation](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/).

#### Cluster autoscaler
{{ kops_feature_table(kops_added_default='1.19') }}

Cluster autoscaler can be enabled to automatically adjust the size of the kubernetes cluster.

```yaml
spec:
  clusterAutoscaler:
    enabled: true
    expander: least-waste
    balanceSimilarNodeGroups: false
    awsUseStaticInstanceList: false
    scaleDownUtilizationThreshold: 0.5
    skipNodesWithLocalStorage: true
    skipNodesWithSystemPods: true
    newPodScaleUpDelay: 0s
    scaleDownDelayAfterAdd: 10m0s
    image: <the latest supported image for the specified kubernetes version>
    cpuRequest: "100m"
    memoryRequest: "300Mi"
```

Read more about cluster autoscaler in the [official documentation](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler).

##### Expander strategies
Cluster autoscaler supports several different [expander strategies](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-are-expanders).

Note that the `priority` expander requires additional configuration through a ConfigMap as described in [its documentation](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/expander/priority/readme.md) - you will need to create this ConfigMap in your cluster before selecting this expander.

##### Disabling cluster autoscaler for a given instance group
{{ kops_feature_table(kops_added_default='1.20') }}

You can disable the autoscaler for a given instance group by adding the following to the instance group spec.

```yaml
spec:
  autoscale: false
```

#### Cert-manager
{{ kops_feature_table(kops_added_default='1.20', k8s_min='1.16') }}

Cert-manager handles x509 certificates for your cluster.

```yaml
spec:
  certManager:
    enabled: true
    defaultIssuer: yourDefaultIssuer
```

**Warning: cert-manager only supports one installation per cluster. If you are already running cert-manager, you need to
either remove this installation prior to enabling this addon, or mark cert-manger as not being managed by kOps (see below).
As long as you are using v1 versions of the cert-manager resources, it is safe to remove existing installs and replace it with this addon**

##### Self-provisioned cert-manager
{{ kops_feature_table(kops_added_default='1.20.2', k8s_min='1.16') }}

The following cert-manager configuration allows provisioning cert-manager externally and allows all dependent plugins
to be deployed. Please note that addons might run into errors until cert-manager is deployed.

```yaml
spec:
  certManager:
    enabled: true
    managed: false
```

##### DNS nameserver configuration for cert-manager pod
{{ kops_feature_table(kops_added_default='1.23.3', k8s_min='1.16') }}

Optional list of DNS nameserver IP addresses for the cert-manager pod to use.
This is useful if you have a public and private DNS zone for the same domain to ensure that cert-manager can access ingress, or DNS01 challenge TXT records at all times.

You can set pod DNS nameserver configuration for cert-manager like so:
```yaml
spec:
  certManager:
    enabled: true
    nameservers:
      - 1.1.1.1
      - 8.8.8.8
```

##### Enabling dns-01 challenges

{{ kops_feature_table(kops_added_default='1.25.0') }}

Cert Manager may be granted the necessary IAM privileges to solve dns-01 challenges by adding a list of hostedzone IDs.
This requires [external permissions for service accounts](/cluster_spec/#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) to be enabled.

```yaml
spec:
  certManager:
    enabled: true
    hostedZoneIDs:
    - ZONEID
  iam:
    useServiceAccountExternalPermissions: true
```

Read more about cert-manager in the [official documentation](https://cert-manager.io/docs/)

#### Karpenter
{{ kops_feature_table(kops_added_default='1.24') }}

The Karpenter addon enables Karpenter-managed InstanceGroups.

```yaml
spec:
  karpenter:
    enabled: true
```

See more details on how to configure Karpenter in the [kOps Karpenter docs](/operations/karpenter) and the [official documentation](https://karpenter.sh)

#### Metrics server
{{ kops_feature_table(kops_added_default='1.19') }}

Metrics Server is a scalable, efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines.

```yaml
spec:
  metricsServer:
    enabled: true
```

Read more about Metrics Server in the [official documentation](https://github.com/kubernetes-sigs/metrics-server).

##### Secure TLS

{{ kops_feature_table(kops_added_default='1.20') }}

By default, API server will not verify the metrics server TLS certificate. To enable TLS verification, set the following in the cluster spec:

```yaml
spec:
  certManager:
    enabled: true
  metricsServer:
    enabled: true
    insecure: false
```

This requires that cert-manager is installed in the cluster.



#### Node local DNS cache
{{ kops_feature_table(kops_added_default='1.18', k8s_min='1.15') }}

NodeLocal DNSCache can be enabled if you are using CoreDNS. It is used to improve the Cluster DNS performance by running a dns caching agent on cluster nodes as a DaemonSet.

`memoryRequest` and `cpuRequest` for the `node-local-dns` pods can also be configured. If not set, they will be configured by default to `5Mi` and `25m` respectively.

If `forwardToKubeDNS` is enabled, kubedns will be used as a default upstream

```yaml
spec:
  kubeDNS:
    provider: CoreDNS
    nodeLocalDNS:
      enabled: true
      memoryRequest: 5Mi
      cpuRequest: 25m
```

#### Node termination handler

{{ kops_feature_table(kops_added_default='1.19') }}

[Node Termination Handler](https://github.com/aws/aws-node-termination-handler) ensures that the Kubernetes control plane responds appropriately to events that can cause your EC2 instance to become unavailable, such as EC2 maintenance events, EC2 Spot interruptions, and EC2 instance rebalance recommendations. If not handled, your application code may not stop gracefully, take longer to recover full availability, or accidentally schedule work to nodes that are going down.

```yaml
spec:
  nodeTerminationHandler:
    cpuRequest: 200m
    enabled: true
    enableRebalanceMonitoring: true
    enableSQSTerminationDraining: true
    managedASGTag: "aws-node-termination-handler/managed"
    prometheusEnable: true
```

##### Queue Processor Mode

{{ kops_feature_table(kops_added_default='1.21') }}

If `enableSQSTerminationDraining` is true Node Termination Handler will operate in Queue Processor mode. In addition to the events mentioned above, Queue Processor mode allows Node Termination Handler to take care of ASG Scale-In, AZ-Rebalance, Unhealthy Instances, EC2 Instance Termination via the API or Console, and more. kOps will provision the necessary infrastructure: an SQS queue, EventBridge rules, and ASG Lifecycle hooks. `managedASGTag` can be configured with Queue Processor mode to distinguish resource ownership between multiple clusters.

The kOps CLI requires additional IAM permissions to manage the requisite EventBridge rules and SQS queue:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "events:DeleteRule",
        "events:ListRules",
        "events:ListTargetsByRule",
        "events:ListTagsForResource",
        "events:PutEvents",
        "events:PutRule",
        "events:PutTargets",
        "events:RemoveTargets",
        "events:TagResource",
        "sqs:CreateQueue",
        "sqs:DeleteQueue",
        "sqs:GetQueueAttributes",
        "sqs:ListQueues",
        "sqs:ListQueueTags"
      ],
      "Resource": "*"
    }
  ]
}
```

**Warning:** If you switch between the two operating modes on an existing cluster, the old resources have to be manually deleted. For IMDS to Queue Processor, this means deleting the k8s nth daemonset. For Queue Processor to IMDS, this means deleting the Kubernetes NTH deployment and the AWS resources: the SQS queue, EventBridge rules, and ASG Lifecycle hooks.

#### Node Problem Detector

{{ kops_feature_table(kops_added_default='1.22') }}

[Node Problem Detector](https://github.com/kubernetes/node-problem-detector) aims to make various node problems visible to the upstream layers in the cluster management stack. It is a daemon that runs on each node, detects node problems and reports them to apiserver.

```yaml
spec:
  nodeProblemDetector:
    enabled: true
    memoryRequest: 32Mi
    cpuRequest: 10m
```

#### Pod Identity Webhook

{{ kops_feature_table(kops_added_default='1.23') }}

When using [IAM roles for Service Accounts](/cluster_spec/#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) (IRSA), Pods require an additinal token to authenticate with the AWS API. In addition, the SDK requires specific environment variables set to make use of these tokens. This addon will mutate Pods configured to use IRSA so that users do not need to do this themselves.

All ServiceAccounts configured with AWS privileges in the Cluster spec will automatically be mutated to assume the configured role.


```yaml
spec:
  certManager:
    enabled: true
  podIdentityWebhook:
    enabled: true
```

The EKS annotations on ServiceAccounts are typically not necessary as kOps will configure the webhook with all ServiceAccount to role mapping configured in the Cluster spec. But if you need specific configuration, you may annotate the ServiceAccount, overriding the kOps configuration.

Read more about Pod Identity Webhook in the [official documentation](https://github.com/aws/amazon-eks-pod-identity-webhook).

#### Snapshot controller

{{ kops_feature_table(kops_added_default='1.21', k8s_min='1.20') }}

Snapshot controller implements the [volume snapshot features](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) of the Container Storage Interface (CSI).

You can enable the snapshot controller by adding the following to the cluster spec:

```yaml
spec:
  snapshotController:
    enabled: true
```

Note that the in-tree volume drivers do not support this feature. If you are running a cluster on AWS, you can enable the EBS CSI driver by adding the following:

```yaml
spec:
  cloudConfig:
    awsEBSCSIDriver:
      enabled: true
```

##### Self-managed aws-ebs-csi-driver

{{ kops_feature_table(kops_added_default='1.25') }}

The following configuration allows for a self-managed aws-ebs-csi-driver. Please note that if you’re using Amazon EBS volumes, you must install the Amazon EBS CSI driver. If the Amazon EBS CSI plugin is not installed, then volume operations will fail. 

If IRSA is not enabled, the control plane will have the permissions to provision nodes, and the self-managed controllers should run on the control plane. If IRSA is enabled, kOps will create the respective AWS IAM Role, assign the policy, and establish a trust relationship allowing the ServiceAccount to assume the IAM Role. To configure Pods to assume the given IAM roles, enable the [Pod Identity Webhook](https://kops.sigs.k8s.io/addons/#pod-identity-webhook). Without this webhook, you need to modify your Pod specs yourself for your Pod to assume the defined roles.

```yaml
spec:
  cloudConfig:
    awsEBSCSIDriver:
      enabled: true
      managed: false
```

## Custom addons

The command `kops create cluster` does not support specifying addons to be added to the cluster when it is created. Instead they can be added after cluster creation using kubectl. Alternatively when creating a cluster from a yaml manifest, addons can be specified using `spec.addons`.

```yaml
spec:
  addons:
  - manifest: s3://my-kops-addons/addon.yaml
```

The docs about the [addon management](contributing/addons.md#addon-management) describe in more detail how to define a addon resource with regards to versioning.
Here is a minimal example of an addon manifest that would install two different addons.

```yaml
kind: Addons
metadata:
  name: example
spec:
  addons:
  - name: foo.addons.org.io
    version: 0.0.1
    selector:
      k8s-addon: foo.addons.org.io
    manifest: foo.addons.org.io/v0.0.1.yaml
  - name: bar.addons.org.io
    version: 0.0.1
    selector:
      k8s-addon: bar.addons.org.io
    manifest: bar.addons.org.io/v0.0.1.yaml
```

In this example the folder structure should look like this;

```
addon.yaml
  foo.addons.org.io
    v0.0.1.yaml
  bar.addons.org.io
    v0.0.1.yaml
```

The yaml files in the foo/bar folders can be any kubernetes resource. Typically this file structure would be pushed to S3 or another of the supported backends and then referenced as above in `spec.addons`. In order for master nodes to be able to access the S3 bucket containing the addon manifests, one might have to add additional iam policies to the master nodes using `spec.additionalPolicies`, like so:

```yaml
spec:
  additionalPolicies:
    master: |
      [
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetObject"
          ],
          "Resource": ["arn:aws:s3:::my-kops-addons/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetBucketLocation",
            "s3:ListBucket"
          ],
          "Resource": ["arn:aws:s3:::my-kops-addons"]
        }
      ]
```
The masters will poll for changes in the bucket and keep the addons up to date.
