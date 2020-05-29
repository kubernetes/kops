This is an overview of how we added a feature:

To add an option for Cilium to use the ENI IPAM mode.

## Adding a field to the API

We want to make this an option, so we need to add a field to CiliumNetworkingSpec:

```
	// Ipam specifies the IP address allocation mode to use.
	// Possible values are "crd" and "eni".
	// "eni" will use AWS native networking for pods. Eni requires masquerade to be set to false.
	// "crd" will use CRDs for controlling IP address management.
	// Empty value will use host-scope address management.
	Ipam string `json:"ipam,omitempty"`
```

A few things to note here:

* We could probably use a boolean for today's needs, but we want to leave some flexibility, so we use a string.

* We define a value `crd` for Cilium's current default mode,
so we leave the default "" value as meaning "default mode, whatever it may be in future".

So, we just need to check if `Ipam` is `eni` when determining which mode to configure.

## Validation

We should add some validation that the value entered is valid.  We only accept `eni`, `crd` or the empty string right now.

Validation is done in validation.go, and is fairly simple - we just add an error to a slice if something is not valid:

```
	if v.Ipam != "" {
		// "azure" not supported by kops
		allErrs = append(allErrs, IsValidValue(fldPath.Child("ipam"), &v.Ipam, []string{"crd", "eni"})...)

		if v.Ipam == kops.CiliumIpamEni {
			if c.CloudProvider != string(kops.CloudProviderAWS) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipam"), "Cilum ENI IPAM is supported only in AWS"))
			}
			if !v.DisableMasquerade {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("disableMasquerade"), "Masquerade must be disabled when ENI IPAM is used"))
			}
		}
	}
```

## Configuring Cilium

Cilium is deployed as a "bootstrap addon", a set of resource template files under upup/models/cloudup/resources/addons/networking.cilium.io,
one file per range of Kubernetes versions. These files are referenced by upup/pkg/fi/cloudup/bootstrapchannelbuilder.go

First we add to the `cilium-config` ConfigMap:

```
  {{ with .Ipam }}
  ipam: {{ . }}
  {{ if eq . "eni" }}
  enable-endpoint-routes: "true"
  auto-create-cilium-node-resource: "true"
  blacklist-conflicting-routes: "false"
  {{ end }}
  {{ end }}
```

Then we conditionally move cilium-operator to masters:

```
      {{ if eq .Ipam "eni" }}
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        operator: Exists
        tolerationSeconds: 300
      - effect: NoExecute
        key: node.kubernetes.io/unreachable
        operator: Exists
        tolerationSeconds: 300
      {{ end }}
```

## Configuring kubelet

When Cilium is in ENI mode `kubelet` needs to be configured with the local IP address, so that it can distinguish it
from the secondary IP address used by ENI. Kubelet is configured by nodeup, in nodeup/pkg/model/kubelet.go. That code
passes the local IP address to `kubelet` when the `UsesSecondaryIP()` receiver of the `NodeupModelContext` returns true.

So we modify `UsesSecondaryIP()` to also return `true` when Cilium is in ENI mode:

```
return (c.Cluster.Spec.Networking.CNI != nil && c.Cluster.Spec.Networking.CNI.UsesSecondaryIP) || c.Cluster.Spec.Networking.AmazonVPC != nil || c.Cluster.Spec.Networking.LyftVPC != nil ||
    (c.Cluster.Spec.Networking.Cilium != nil && c.Cluster.Spec.Networking.Cilium.Ipam == kops.CiliumIpamEni)
```

## Configuring IAM

When Cilium is in ENI mode, `cilium-operator` on the master nodes needs additional IAM permissions. The masters' IAM permissions
are built by `BuildAWSPolicyMaster()` in pkg/model/iam/iam_builder.go:

```
	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.Ipam == kops.CiliumIpamEni {
		addCiliumEniPermissions(p, resource, b.Cluster.Spec.IAM.Legacy)
	}
```

```
func addCiliumEniPermissions(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool) {
	if legacyIAM {
		// Legacy IAM provides ec2:*, so no additional permissions required
		return
	}

	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:DescribeSubnets",
				"ec2:AttachNetworkInterface",
				"ec2:AssignPrivateIpAddresses",
				"ec2:UnassignPrivateIpAddresses",
				"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DescribeVpcPeeringConnections",
				"ec2:DescribeSecurityGroups",
				"ec2:DetachNetworkInterface",
				"ec2:DeleteNetworkInterface",
				"ec2:ModifyNetworkInterfaceAttribute",
				"ec2:DescribeVpcs",
			}),
			Resource: resource,
		},
	)
}
```
## Tests

Prior to testing this for real, it can be handy to write a few unit tests.

We should test that validation works as we expect (in validation_test.go):

```
func Test_Validate_Cilium(t *testing.T) {
	grid := []struct {
		Cilium         kops.CiliumNetworkingSpec
		Spec           kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Cilium: kops.CiliumNetworkingSpec{},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Ipam: "crd",
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				DisableMasquerade: true,
				Ipam:              "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: "aws",
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				DisableMasquerade: true,
				Ipam:              "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: "aws",
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Ipam: "foo",
			},
			ExpectedErrors: []string{"Unsupported value::cilium.ipam"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Ipam: "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: "aws",
			},
			ExpectedErrors: []string{"Forbidden::cilium.disableMasquerade"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				DisableMasquerade: true,
				Ipam:              "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: "gce",
			},
			ExpectedErrors: []string{"Forbidden::cilium.ipam"},
		},
	}
	for _, g := range grid {
		g.Spec.Networking = &kops.NetworkingSpec{
			Cilium: &g.Cilium,
		}
		errs := validateNetworkingCilium(&g.Spec, g.Spec.Networking.Cilium, field.NewPath("cilium"))
		testErrors(t, g.Spec, errs, g.ExpectedErrors)
	}
}
```

## Documentation

If your feature touches important configuration options in `config` or `cluster.spec`, document them in [cluster_spec.md](../cluster_spec.md).

## Testing

You can `make` and run `kops` locally.  But `nodeup` is pulled from an S3 bucket.

To rapidly test a nodeup change, you can build it, scp it to a running machine, and
run it over SSH with the output viewable locally:

`make push-aws-run TARGET=admin@<publicip>`


For more complete testing though, you will likely want to do a private build of
nodeup and launch a cluster from scratch.

To do this, you can repoint the nodeup source url by setting the `NODEUP_URL` env var,
and then push nodeup using:


```
export S3_BUCKET_NAME=<yourbucketname>
make kops-install dev-upload UPLOAD_DEST=s3://${S3_BUCKET_NAME}

KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export KOPS_BASE_URL=https://${S3_BUCKET_NAME}.s3.amazonaws.com/kops/${KOPS_VERSION}/
kops create cluster <clustername> --zones us-east-1b
...
```

If you have changed the dns or kops controllers, you would want to test them as well. To do so, run the respective snippets below before creating the cluster.

For dns-controller:

```bash
KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export DOCKER_IMAGE_PREFIX=${USER}/
export DOCKER_REGISTRY=
make dns-controller-push
export DNSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}dns-controller:${KOPS_VERSION}
```

For kops-controller:

```bash
KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export DOCKER_IMAGE_PREFIX=${USER}/
export DOCKER_REGISTRY=
make kops-controller-push
export KOPSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_VERSION}
```



## Using the feature

Users would simply `kops edit cluster`, and add a value like:
```
  spec:
    networking:
      cilium:
        disableMasquerade: true
        ipam: eni
```

Then `kops update cluster --yes` would create the new NodeUpConfig, which is included in the instance startup script
and thus requires a new LaunchConfiguration, and thus a `kops rolling update`.  We're working on changing settings
without requiring a reboot, but likely for this particular setting it isn't the sort of thing you need to change
very often.

## Other steps

* We could also create a CLI flag on `create cluster`.  This doesn't seem worth it in this case; this is a relatively advanced option.
