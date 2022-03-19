# Experimental features

Enable experimental features with:

`export KOPS_FEATURE_FLAGS=`

The following experimental features are currently available:

* `+EnableExternalDNS` - Enable external-dns with default settings (ingress sources only).
* `+VPCSkipEnableDNSSupport` - Enables creation of a VPC that does not need DNSSupport enabled.
* `+EnableSeparateConfigBase` - Allow a config-base that is different from the state store.
* `+SpecOverrideFlag` - Allow setting spec values on `kops create`.
* `+ExperimentalClusterDNS` - Turns off validation of the kubelet cluster dns flag.
* `+GoogleCloudBucketAcl` - Enables setting the ACL on the state store bucket when using GCS
* `+Spotinst` - Enables the use of the Spot integration
* `+SpotinstOcean` - Enables the use of the Spot Ocean integration
* `+SpotinstOceanTemplate` - Enables the use of Spot Ocean object as a template for Virtual Node Groups
* `+SpotinstHybrid` - Toggles between hybrid and full instance group implementations
* `-SpotinstController` - Toggles the installation of the Spot controller addon off
* `+SkipEtcdVersionCheck` - Bypasses the check that etcd-manager is using a supported etcd version
* `+VFSVaultSupport` - Enables setting Vault as secret/keystore
* `+APIServerNodes` - Enables support for dedicated API server nodes
