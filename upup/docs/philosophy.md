There is a schema-ed object ClusterConfiguration

Users tweak values in the "specified" ClusterConfiguration

We compute the "complete" ClusterConfiguration by populating defaults and inferring values
  * We try to remove any logic from downstream pieces
  * This also means that there is one source of truth

Note this is a little different to how kubernetes specs normally work, k8s has a
separation between spec and status, but this is all spec.  k8s will auto-populate the spec
and not retain the "user-specified" spec, and this sometimes causes a few problems when it comes to
exports & updates (e.g. ClusterIP).  By storing the complete spec separately we ensure that the spec
has all the information - so dependent steps don't have inference logic - but we still only keep the
values that are specified.  As a concrete example, we only store the kubernetes version if the user specifies
it, if not we will follow k8s versions as they come out.  (TODO: Not the best example.  Maybe instance type?)

The way we store the ClusterConfiguration is an implementation detail, in terms of how it is broken
into files.  This might well change in future.  For example, we might put NodeSet configuration storage into the
kubernetes API.
