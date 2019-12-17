## HookSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | HookSpec



HookSpec is a definition hook

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
<li><a href="#instancegroupspec-v1alpha2-kops">InstanceGroupSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
before <br /> *string array*    | Before is a series of systemd units which this hook must run before
disabled <br /> *boolean*    | Disabled indicates if you want the unit switched off
execContainer <br /> *[ExecContainerAction](#execcontaineraction-v1alpha2-kops)*    | ExecContainer is the image itself
manifest <br /> *string*    | Manifest is a raw systemd unit file
name <br /> *string*    | Name is an optional name for the hook, otherwise the name is kops-hook-<index>
requires <br /> *string array*    | Requires is a series of systemd units the action requires
roles <br /> *string array*    | Roles is an optional list of roles the hook should be rolled out to, defaults to all

