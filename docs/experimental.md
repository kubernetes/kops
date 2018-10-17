# Experimental features

Enable experimental features with:

`export KOPS_FEATURE_FLAGS=`

The following experimental features are currently available:

* `+VSphereCloudProvider` - Enable vSphere cloud provider.
* `+EnableExternalDNS` - Enable external-dns with default settings (ingress sources only).
* `+VPCSkipEnableDNSSupport` - Enables creation of a VPC that does not need DNSSupport enabled.
* `+SkipTerraformFormat` - Do not `terraform fmt` the generated terraform files.
* `+EnableExternalCloudController` - Enables the use of cloud-controller-manager introduced in v1.7.
* `+EnableSeparateConfigBase` - Allow a config-base that is different from the state store.
* `+SpecOverrideFlag` - Allow setting spec values on `kops create`.
* `+ExperimentalClusterDNS` - Turns off validation of the kubelet cluster dns flag.
* `+EnableNodeAuthorization` - Enable support of Node Authorization, see [node_authorization.md](node_authorization.md).
