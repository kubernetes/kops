### Weave

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
AWS VPCs support jumbo frames, so on cluster creation kops sets the weave MTU to 8912 bytes (9001 minus overhead).

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
If no password is supplied, kops will generate one at random.

```sh
cat /dev/urandom | tr -dc A-Za-z0-9 | head -c9 > password
kops create secret weavepassword -f password
kops update cluster
```

Since unencrypted nodes will not be able to connect to nodes configured with encryption enabled, this configuration cannot be changed easily without downtime.

