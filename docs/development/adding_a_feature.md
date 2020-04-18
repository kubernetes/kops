This is an overview of how we added a feature:

To make auto-upgrading on the nodes an option.

Auto-upgrades are configured by nodeup.  nodeup is driven by the nodeup model (which is at [upup/models/nodeup/](https://github.com/kubernetes/kops/tree/master/upup/models/nodeup) )

Inside the nodeup model there are folders which serve three roles:

1) A folder with a well-known name means that items under that folder are treated as items of that type:

* files
* packages
* services

2) A folder starting with an underscore is a tag: nodeup will only descend into that folder if a tag with
the same name is configured.

3) Remaining folders are just structural, for organization.

So auto-upgrades are currently always enabled, so the folder `auto-upgrades` configures them.

To make auto-upgrades option, we will rename it to a "tag" folder (`_automatic_upgrades`), and then plumb through
the tag.  The rename is a simple file rename.

## Passing the `_automatic_upgrades` tag to nodeup

Tags reach nodeup from the `NodeUpConfig`.  And this is in turn populated by the `RenderNodeUpConfig` template function,
in `apply_cluster.go`.

(`RenderNodeUpConfig` is called inline from the instance startup script on AWS, in a heredoc.  On GCE,
it is rendered into its own resource, because GCE supports multiple resources for an instance)

If you look at the code for RenderNodeUpConfig, you can see that it in turn gets the tags by calling `buildNodeupTags`.

We want to make this optional, and it doesn't really make sense to have automatic upgrades at the instance group level:
either you trust upgrades or you don't.  At least that's a working theory; if we need to go the other way later we can
easily use the cluster value as the default.

So we need to add a field to ClusterSpec:

```
	// UpdatePolicy determines the policy for applying upgrades automatically.
	// Valid values:
	//   'external' do not apply upgrades
	//   missing: default policy (currently OS security upgrades that do not require a reboot)
	UpdatePolicy *string `json:"updatePolicy,omitempty"`
```

A few things to note here:

* We could probably use a boolean for today's needs, but we want to leave some flexibility, so we use a string.

* We use a `*string` instead of a `string` so we can know if the value is actually set.  This is less important
for strings than it is for booleans, where false can be very different from unset.

* We only define the value we care about for no - `external` to disable upgrades.  We could probably define an
actual value for enabled upgrades, but it isn't yet clear what that policy should be or what it should be called,
so we leave the nil value as meaning "default policy, whatever it may be in future".


So, we just need to check if `UpdatePolicy` is not nil and == `external`; we add the tag `_automatic_upgrades`,
which enabled automatic upgrades, only if that is _not_ the case!

## Validation

We should add some validation that the value entered is valid.  We only accept nil or `external` right now.

Validation is done in validation.go, and is fairly simple - we just return an error if something is not valid:

```
	// UpdatePolicy
	if c.Spec.UpdatePolicy != nil {
		switch (*c.Spec.UpdatePolicy) {
		case UpdatePolicyExternal:
			// Valid
		default:
			return fmt.Errorf("unrecognized value for UpdatePolicy: %v", *c.Spec.UpdatePolicy)
		}
	}
```

## Tests

Prior to testing this for real, it can be handy to write a few unit tests.

We should test that validation works as we expect (in validation_test.go):

```
func TestValidateFull_UpdatePolicy_Valid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.UpdatePolicy = fi.String(api.UpdatePolicyExternal)
	expectNoErrorFromValidate(t, c)
}

func TestValidateFull_UpdatePolicy_Invalid(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.UpdatePolicy = fi.String("not-a-real-value")
	expectErrorFromValidate(t, c, "UpdatePolicy")
}
```


And we should test the nodeup tag building:

```
func TestBuildTags_UpdatePolicy_Nil(t *testing.T) {
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			UpdatePolicy: nil,
		},
	}

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if !stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatalf("nodeUpTag _automatic_upgrades not found")
	}
}

func TestBuildTags_UpdatePolicy_External(t *testing.T) {
	c := &api.Cluster{
		Spec: api.ClusterSpec{
			UpdatePolicy: fi.String("external"),
		},
	}

	tags, err := buildCloudupTags(c)
	if err != nil {
		t.Fatalf("buildCloudupTags error: %v", err)
	}

	nodeUpTags, err := buildNodeupTags(api.InstanceGroupRoleNode, c, tags)
	if err != nil {
		t.Fatalf("buildNodeupTags error: %v", err)
	}

	if stringSliceContains(nodeUpTags, "_automatic_upgrades") {
		t.Fatalf("nodeUpTag _automatic_upgrades found unexpectedly")
	}
}
```

## Documentation

Add some documentation on your new field:

```
## UpdatePolicy

Cluster.Spec.UpdatePolicy

Values:

* `external` do not enable automatic software updates

* unset means to use the default policy, which is currently to apply OS security updates unless they require a reboot
```

Additionally, consider adding documentation of your new feature to the docs in [/docs](/). If your feature touches configuration options in `config` or `cluster.spec`, document them in [cluster_spec.md](../cluster_spec.md).

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
    updatePolicy: external
```

Then `kops update cluster --yes` would create the new NodeUpConfig, which is included in the instance startup script
and thus requires a new LaunchConfiguration, and thus a `kops rolling update`.  We're working on changing settings
without requiring a reboot, but likely for this particular setting it isn't the sort of thing you need to change
very often.

## Other steps

* We could also create a CLI flag on `create cluster`.  This doesn't seem worth it in this case; this is a relatively advanced option
for people that already have an external software update mechanism.  All the flag would do is save the default.
