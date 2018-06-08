## FileAssetSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | FileAssetSpec



FileAssetSpec defines the structure for a file asset

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
<li><a href="#instancegroupspec-v1alpha2-kops">InstanceGroupSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
content <br /> *string*    | Content is the contents of the file
isBase64 <br /> *boolean*    | IsBase64 indicates the contents is base64 encoded
name <br /> *string*    | Name is a shortened reference to the asset
path <br /> *string*    | Path is the location this file should reside
roles <br /> *string array*    | Roles is a list of roles the file asset should be applied, defaults to all

