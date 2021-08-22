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
{{ kops_feature_table(kops_added_default='1.21', k8s_min='1.16') }}

The following cert-manager configuration allows provisioning cert-manager externally and allows all dependent plugins
to be deployed. Please note that addons might run into errors until cert-manager is deployed.

```yaml
spec:
  certManager:
    enabled: true
    managed: false
```


Read more about cert-manager in the [official documentation](https://cert-manager.io/docs/)

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
    enabled: true
    enableSQSTerminationDraining: true
    managedASGTag: "aws-node-termination-handler/managed"
```

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
        "events:PutTargets",
        "events:RemoveTargets",
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

**Warning: If you switch between the two operating modes on an existing cluster, the old resources have to be manually deleted. For IMDS to Queue Processor, this means deleting the k8s nth daemonset. For Queue Processor to IMDS, this means deleting the k8s nth deployment and the AWS resources: the SQS queue, EventBridge rules, and ASG Lifecycle hooks.**

## Static addons

The command `kops create cluster` does not support specifying addons to be added to the cluster when it is created. Instead they can be added after cluster creation using kubectl. Alternatively when creating a cluster from a yaml manifest, addons can be specified using `spec.addons`.
```yaml
spec:
  addons:
  - manifest: kubernetes-dashboard
  - manifest: s3://kops-addons/addon.yaml
```

This document describes how to install some common addons and how to create your own custom ones.

### Available addons

#### Ambassador

The [Ambassador API Gateway](https://getambassador.io/) provides all the functionality of a traditional ingress
controller (i.e., path-based routing) while exposing many additional capabilities such as authentication, URL rewriting,
CORS, rate limiting, and automatic metrics collection.

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ambassador/ambassador-operator.yaml
```

Detailed installation instructions in the [addon documentation](https://github.com/kubernetes/kops/blob/master/addons/ambassador/README.md).
See [Ambassador documentation](https://www.getambassador.io/docs/) on configuration and usage.

#### Dashboard

The [dashboard project](https://github.com/kubernetes/dashboard) provides a nice administrative UI:

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/kubernetes-dashboard/v1.10.1.yaml
```

And then follow the instructions in the [dashboard documentation](https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above) to access the dashboard.

The login credentials are:

* Username: `admin`
* Password: get by running `kops get secrets kube --type secret -oplaintext` or `kubectl config view --minify`

##### RBAC

It's necessary to add your own RBAC permission to the dashboard. Please read the [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) docs before applying permissions.

Below you see an example giving **cluster-admin access** to the dashboard.

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubernetes-dashboard
  labels:
    k8s-app: kubernetes-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: kubernetes-dashboard
  namespace: kube-system
```

### Monitoring with Heapster - Standalone

**This addons is deprecated. Please use metrics-server instead**

Monitoring supports the horizontal pod autoscaler.

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.11.0.yaml
```
Please note that [heapster is retired](https://github.com/kubernetes/heapster/blob/master/docs/deprecation.md). Consider using [metrics-server](https://github.com/kubernetes-incubator/metrics-server) and a third party metrics pipeline to gather Prometheus-format metrics instead.

### Monitoring with Prometheus Operator + kube-prometheus

The [Prometheus Operator](https://github.com/coreos/prometheus-operator/) makes the Prometheus configuration Kubernetes native and manages and operates Prometheus and Alertmanager clusters. It is a piece of the puzzle regarding full end-to-end monitoring.

[kube-prometheus](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus) combines the Prometheus Operator with a collection of manifests to help getting started with monitoring Kubernetes itself and applications running on top of it.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/prometheus-operator/v0.26.0.yaml
```

### Route53 Mapper

**This addon is deprecated. Please use [external-dns](https://github.com/kubernetes-sigs/external-dns) instead.**

Please note that kOps installs a Route53 DNS controller automatically (it is required for cluster discovery).
The functionality of the route53-mapper overlaps with the dns-controller, but some users will prefer to
use one or the other.
[README for the included dns-controller](https://github.com/kubernetes/kops/blob/master/dns-controller/README.md)

route53-mapper automates creation and updating of entries on Route53 with `A` records pointing
to ELB-backed `LoadBalancer` services created by Kubernetes. Install using:

The project is created by wearemolecule, and maintained at
[wearemolecule/route53-kubernetes](https://github.com/wearemolecule/route53-kubernetes).
[Usage instructions](https://github.com/kubernetes/kops/blob/master/addons/route53-mapper/README.md)

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/route53-mapper/v1.3.0.yml
```

### Custom addons

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

The yaml files in the foo/bar folders can be any kubernetes resource. Typically this file structure would be pushed to S3 or another of the supported backends and then referenced as above in `spec.addons`. In order for master nodes to be able to access the S3 bucket containing the addon manifests, one might have to add additional iam policies to the master nodes using `spec.additionalPolicies`, like so;
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
          "Resource": ["arn:aws:s3:::kops-addons/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetBucketLocation",
            "s3:ListBucket"
          ],
          "Resource": ["arn:aws:s3:::kops-addons"]
        }
      ]
```
The masters will poll for changes in the bucket and keep the addons up to date.
