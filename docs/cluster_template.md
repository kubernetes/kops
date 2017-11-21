# Cluster Templating

The command `kops replace` can replace a cluster desired configuration from the config in a yaml file (see [cli/kops_replace.md](cli/kops_replace.md)).

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

You can pass configuration such as an environment file by using the `--values PATH` command line option. Note `--values` is a slice so can be defined multiple times; the configuration is overridden by each configuration file *(so order is important assuming duplicating values)*; a use-case for this would be a default configuration which upstream clusters can override.

The file passed as `--values` must contain the variables referenced in the template. Example:
```yaml
# File values.yaml
clusterName: eu1
kubernetesVersion: 1.7.1
dnsZone: k8s.example.com
awsRegion: eu-west-1
```

Running `kops toolbox template` replaces the placeholders in the template by values and generates the file output.yaml, which can then be used to replace the desired cluster configuration with `kops replace -f cluster.yaml`.

Note: when creating a cluster desired configuration template, you can

- use `kops get k8s-cluster.example.com -o yaml > cluster-desired-config.yaml` to create the cluster desired configuration file (see [cli/kops_get.md](cli/kops_get.md)). The values in this file are defined in [cluster_spec.md](cluster_spec.md).
- replace values by placeholders in that file to create the template.

### Templates

The `--template` command line option can point to either a specific file or a directory with a collection of templates. An example usage would be;

```shell
$ kops toolbox template --values dev.yaml --template cluster.yaml --template instance_group_directory
```

The cluster.yaml *(your main cluster spec for example)* would be written first followed by any templates found in the instance_group_directory directory. Note the toolbox will automatically add YAML separators between the documents for you.

### Snippets

The toolbox template also supports the reuse or break up of code blocks into snippets directories. By passing a `--snippets PATH` to a directory holding templates;

```shell
$ kops toolbox template --values dev.yaml --template cluster.yaml --template instancegroups --snippets snippets
```

The example below assumes you have placed the appropriate files i.e. *(nodes.json, master.json etc)* in to the snippets directory. Note, the namespace of the snippets are flat and always the basename() of the file path; so `snippets/components/docker.options` is still referred to as 'docker.options'.

```YAML
apiVersion: kops/v1alpha2
kind: Cluster
metadata:
  name: {{ .environment }}.{{ .dns_zone }}
spec:
  docker:
    {{ include "docker" . | indent 4 }}
  additionalPolicies:
    master: |
      {{ include "masters.json" . | indent 6 }}
    node: |
      {{ include "nodes.json" . | indent 6 }}
```

### Template Functions

The entire set of https://github.com/Masterminds/sprig functions are available within the templates for you. Note if you want to use the 'defaults' functions switch off the verification check on the command line by `--fail-on-missing=false`;

```YAML
image: {{ default $image $node.image }}
machineType: {{ default $instance $node.machine_type }}
maxSize: {{ default "10" $node.max_size }}
minSize: {{ default "1" $node.min_size }}
```

### Formatting

Formatting in golang templates is a pain! At the start or at the end of a statement can be infuriating to get right, so a `--format-yaml=true` *(defaults to false)* command line option has been added. This will first unmarshal the generated content *(performing a syntax verification)* and then marshal back the content removing all those nasty formatting issues, newlines etc.
