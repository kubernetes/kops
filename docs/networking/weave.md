### Weave

&#9888; The Weave CNI is not supported for Kubernetes 1.23 or later.

#### Installation

To use the Weave, specify the following in the cluster spec.

```yaml
  networking:
    weave: {}
```

The following command sets up a cluster using Weave.

```sh
export ZONES=mylistofzone
kops create cluster \
  --zones $ZONES \
  --networking weave \
  --yes \
  --name myclustername.mydns.io
```

### Configuring Weave MTU

The Weave MTU is configurable by editing the cluster and setting `mtu` option in the weave configuration.
AWS VPCs support jumbo frames, so on cluster creation kOps sets the weave MTU to 8912 bytes (9001 minus overhead).

```yaml
spec:
  networking:
    weave:
      mtu: 8912
```

### Configuring Weave Net EXTRA_ARGS

Weave allows you to pass command line arguments to weave by adding those arguments to the EXTRA_ARGS environmental variable.
This can be used for debugging or for customizing the logging level of weave net.

```yaml
spec:
  networking:
    weave:
      netExtraArgs: "--log-level=info"
```

Note that it is possible to break the cluster networking if flags are improperly used and as such this option should be used with caution.

### Configuring Weave NPC EXTRA_ARGS

Weave-npc (the Weave network policy controller) allows you to customize arguments of the running binary by setting the EXTRA_ARGS environmental variable.
This can be used for debugging or for customizing the logging level of weave npc.

```yaml
spec:
  networking:
    weave:
      npcExtraArgs: "--log-level=info"
```

Note that it is possible to break the cluster networking if flags are improperly used and as such this option should be used with caution.

### Configuring Weave network encryption

The Weave network encryption is configurable by creating a weave network secret password.
Weaveworks recommends choosing a secret with [at least 50 bits of entropy](https://www.weave.works/docs/net/latest/tasks/manage/security-untrusted-networks/).
If no password is supplied, kOps will generate one at random.

```sh
cat /dev/urandom | tr -dc A-Za-z0-9 | head -c9 > password
kops create secret weavepassword -f password
kops update cluster
```

Since unencrypted nodes will not be able to connect to nodes configured with encryption enabled, this configuration cannot be changed easily without downtime.

### Override Weave image tag
{{ kops_feature_table(kops_added_default='1.19', k8s_min='1.12') }}

Weave networking comes with default specs and version which are the recommended ones, already configured by kOps .
In case users want to override Weave image tag, thus default version, specs should be customized as follows:
```yaml
spec:
  networking:
    weave:
      version: "2.7.0"
```

### Override default CPU/Memory resources

Weave networking comes with default specs related to CPU/Memory requests and limits, already configured by kOps.
In case users want to override default values, specs should be customized as follows:

```yaml
spec:
  networking:
    weave:
      memoryRequest: 300Mi
      cpuRequest: 100m
      memoryLimit: 300Mi
      cpuLimit: 100m
      npcMemoryRequest: 300Mi
      npcCPURequest: 100m
      npcMemoryLimit: 300Mi
      npcCPULimit: 100m
```

> **NOTE**: These are just example values and not necessarily the recommended values. You should override the default values according to your needs.
