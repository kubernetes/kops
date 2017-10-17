# Cluster template

The command `kops replace` can replace a cluster desired configuration from the config in a yaml file (see [/cli/kops_replace.md](/cli/kops_replace.md)).

It is possible to generate that yaml file from a template, using the command `kops toolbox template` (see [cli/kops_toolbox_template.md](cli/kops_toolbox_template.md)).

This document details the template language used.

The file passed as `--template` must be a [go template](https://golang.org/pkg/text/template/). Example:
```yaml
# File cluster.tmpl.yaml
apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
  kops.k8s.io/cluster: {{.clusterName}}.{{.dnsZone}}
  name: nodes
spec:
  image: coreos.com/CoreOS-stable-1409.6.0-hvm
  kubernetesVersion: {{.kubernetesVersion}
  machineType: m4.large
  maxPrice: "0.5"
  maxSize: 2
  minSize: 15
  role: Node
  rootVolumeSize: 100
  subnets:
  - {{.awsRegion}}a
  - {{.awsRegion}}b
  - {{.awsRegion}}c
```

The file passed as `--values` must contain the variables referenced in the template. Example:
```yaml
# File values.yaml
clusterName: eu1
kubernetesVersion: 1.7.1
dnsZone: k8s.example.com
awsRegion: eu-west-1
```

Running `kops toolbox template` replaces the placeholders in the template by values and generates the file output.yaml, which can then be used to replace the desired cluster configuration with `kops replace -f cluster.yaml`.


Note:
When creating a cluster desired configuration template, you can 
- use `kops get k8s-cluster.example.com -o yaml > cluster-desired-config.yaml` to create the cluster desired configuration file (see [cli/kops_get.md](cli/kops_get.md)). The values in this file are defined in [cli/cluster_spec.md](cli/cluster_spec.md).
- replace values by placeholders in that file to create the template.
