# Experimental features

Enable experimental features with:

`export KOPS_FEATURE_FLAGS=`

The following experimental features are currently available:

* `+VSphereCloudProvider` - Enable vSphere cloud provider.
* `+DrainAndValidateRollingUpdate` - Enable drain and validate for rolling updates.
* `+EnableExternalDNS` - Enable external-dns with default settings (ingress sources only).
* `+RollingUpdateStrategies` - Enable strategies for replacing nodes during rolling updates.
* `+VPCSkipEnableDNSSupport` - if set will mean that VPC does not need DNSSupport enabled.
* `+SkipTerraformFormat` - if set will mean that we will not `tf fmt` the generated terraform.


