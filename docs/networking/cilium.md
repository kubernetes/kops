# Cilium
* [Cilium](http://docs.cilium.io)

The Cilium CNI uses a Linux kernel technology called BPF, which enables the dynamic insertion of powerful security visibility and control logic within the Linux kernel.

## Installing Cilium on a new Cluster

To use the Cilium, specify the following in the cluster spec.

```yaml
  networking:
    cilium: {}
```

The following command sets up a cluster using Cilium.

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking cilium\
  --yes \
  --name cilium.example.com
```

## Configuring Cilium

### Using etcd for agent state sync

This feature is in beta state as of kOps 1.18.

By default, Cilium will use CRDs for synchronizing agent state. This can cause performance problems on larger clusters. As of kOps 1.18, kOps can manage an etcd cluster using etcd-manager dedicated for cilium agent state sync. The [Cilium docs](https://docs.cilium.io/en/stable/gettingstarted/k8s-install-external-etcd/) contains recommendations for when this must be enabled.

For new clusters you can use the `cilium-etcd` networking provider:

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking cilium-etcd \
  --yes \
  --name cilium.example.com
```

For existing clusters, add the following to `spec.etcdClusters`:
Make sure `instanceGroup` match the other etcd clusters.
You should also enable auto compaction.

```yaml
  - etcdMembers:
    - instanceGroup: master-az-1a
      name: a
    - instanceGroup: master-az-1b
      name: b
    - instanceGroup: master-az-1c
      name: c
    manager:
      env:
      - name: ETCD_AUTO_COMPACTION_MODE
        value: revision
      - name: ETCD_AUTO_COMPACTION_RETENTION
        value: "2500"
    name: cilium
```

It is important that you perform a rolling update on the entire cluster so that all the nodes can connect to the new etcd cluster.

```sh
kops update cluster
kops update cluster --yes
kops rolling-update cluster --force --yes

```

Then enable etcd as kvstore:

```yaml
  networking:
    cilium:
      etcdManaged: true
```

### Enabling BPF NodePort

As of kOps 1.19, BPF NodePort is enabled by default for new clusters if the kubernetes version is 1.12 or newer. It can be safely enabled as of kOps 1.18.

In this mode, the cluster is fully functional without kube-proxy, with Cilium replacing kube-proxy's NodePort implementation using BPF.
Read more about this in the [Cilium docs - kubeproxy free](https://docs.cilium.io/en/stable/gettingstarted/kubeproxy-free/) and [Cilium docs - NodePort](https://docs.cilium.io/en/stable/gettingstarted/kubeproxy-free/#nodeport-devices)

Be aware that you need to use an AMI with at least Linux 4.19.57 for this feature to work.

Also be aware that while enabling this on an existing cluster is safe, disabling this is disruptive and requires you to run `kops rolling-upgrade cluster --cloudonly`.

```yaml
  kubeProxy:
    enabled: false
  networking:
    cilium:
      enableNodePort: true
```

If you are migrating an existing cluster, you need to manually roll the cilium DaemonSet before rolling the cluster:

```
kops update cluster
kops update cluster --yes
kubectl rollout restart ds/cilium -n kube-system
kops rolling-update cluster --yes
```

### Enabling Cilium ENI IPAM

{{ kops_feature_table(kops_added_default='1.18') }}

This feature is in beta state.

You can have Cilium provision AWS managed addresses and attach them directly to Pods much like AWS VPC. See [the Cilium docs for more information](https://docs.cilium.io/en/v1.6/concepts/ipam/eni/)

```yaml
  networking:
    cilium:
      ipam: eni
```

In kOps versions before 1.22, when using ENI IPAM you need to explicitly disable masquerading in Cilium as well.

```yaml
  networking:
    cilium:
      disableMasquerade: true
      ipam: eni
```

Note that since Cilium Operator is the entity that interacts with the EC2 API to provision and attaching ENIs, we force it to run on the master nodes when this IPAM is used.

Also note that this feature has only been tested on the default kOps AMIs.

#### Enabling Encryption in Cilium

##### ipsec
{{ kops_feature_table(kops_added_default='1.19', k8s_min='1.17') }}

As of kOps 1.19, it is possible to enable encryption for Cilium agent.
In order to enable encryption, you must first generate the pre-shared key using this command:
```bash
cat <<EOF | kops create secret ciliumpassword -f -
keys: $(echo "3 rfc4106(gcm(aes)) $(echo $(dd if=/dev/urandom count=20 bs=1 2> /dev/null| xxd -p -c 64)) 128")
EOF
```
The above command will create a dedicated secret for cilium and store it in the kOps secret store.
Once the secret has been created, encryption can be enabled by setting `enableEncryption` option in `spec.networking.cilium` to `true`:
```yaml
  networking:
    cilium:
      enableEncryption: true
```

##### wireguard
{{ kops_feature_table(kops_added_default='1.22', k8s_min='1.17') }}

Cilium can make use of the [wireguard protocol for transparent encryption](https://docs.cilium.io/en/v1.10/gettingstarted/encryption-wireguard/). Take care to familiarise yourself with the [limitations](https://docs.cilium.io/en/v1.10/gettingstarted/encryption-wireguard/#limitations).

```yaml
  networking:
    cilium:
      enableEncryption: true
      enableL7Proxy: false
      encryptionType: wireguard
```


#### Resources in Cilium
{{ kops_feature_table(kops_added_default='1.21', k8s_min='1.20') }}

As of kOps 1.20, it is possible to choose your own values for Cilium Agents + Operator. Example:
```yaml
  networking:
    cilium:
      cpuRequest: "25m"
      memoryRequest: "128Mi"
```

## Hubble
{{ kops_feature_table(kops_added_default='1.20.1', k8s_min='1.20') }}

Hubble is the observability layer of Cilium and can be used to obtain cluster-wide visibility into the network and security layer of your Kubernetes cluster. See the [Hubble documentation](https://docs.cilium.io/en/v1.10/gettingstarted/hubble_setup/) for more information.

Hubble can be enabled by adding the following to the spec:
```yaml
  networking:
    cilium:
      hubble:
        enabled: true
```

This will enable Hubble in the Cilium agent as well as install hubble-relay. kOps will also configure mTLS between the Cilium agent and relay. Note that since the Hubble UI does not support TLS, the relay is not configured to listen on a secure port.

The Hubble UI has to be installed separatly.

## Hubble UI

Hubble UI brings a dashboard on top of Hubble observability layer. It allows viewing service map and TCP flows directly inside a browser.

When Cilium is intalled and managed by kOps, Cilium cli should not be used as the configuration it produces conflicts with the configuration managed by kOps (certificates are not managed the same way). For this reason, deploying Hubble UI can be tricky.

Fortunately, recent versions of the Cilium Helm chart allow standalone install of Hubble UI. See `Helm (Standalone install)` tab in [Hubble UI documentation](https://docs.cilium.io/en/stable/gettingstarted/hubble/).

Basically, it requires to disable all components in the chart (they are already managed by kOps) except Hubble UI, and setting `hubble.ui.standalone.enabled` to `true`.

A minimal command line install should look like this:

```
helm upgrade --install --namespace kube-system --repo https://helm.cilium.io cilium cilium --version 1.11.1 --values - <<EOF
agent: false

operator:
  enabled: false

cni:
  install: false

hubble:
  enabled: false

  relay:
    enabled: false

  ui:
    # enable Hubble UI
    enabled: true

    standalone:
      # enable Hubble UI standalone deployment
      enabled: true

    # ...
EOF
```

Note that you can create an ingress resource for Hubble UI by configuring the `hubble.ui.ingress` stanza. See [Cilium Helm chart documentation](https://artifacthub.io/packages/helm/cilium/cilium/1.11.1) for more information.

## Getting help

For problems with deploying Cilium please post an issue to Github:

- [Cilium Issues](https://github.com/cilium/cilium/issues)

For support with Cilium Network Policies you can reach out on Slack or Github:

- [Cilium Github](https://github.com/cilium/cilium)
- [Cilium Slack](https://cilium.io/slack)
