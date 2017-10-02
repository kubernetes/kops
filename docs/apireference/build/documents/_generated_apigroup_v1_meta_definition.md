## APIGroup v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | APIGroup



APIGroup contains the name, the supported versions, and the preferred version of a group.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#apigrouplist-v1-meta">APIGroupList meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
name <br /> *string*    | name is the name of the group.
preferredVersion <br /> *[GroupVersionForDiscovery](#groupversionfordiscovery-v1-meta)*    | preferredVersion is the version preferred by the API server, which probably is the storage version.
serverAddressByClientCIDRs <br /> *[ServerAddressByClientCIDR](#serveraddressbyclientcidr-v1-meta) array*    | a map of client CIDR to server address that is serving this group. This is to help clients reach servers in the most network-efficient way possible. Clients can use the appropriate server address as per the CIDR that they match. In case of multiple matches, clients should use the longest matching CIDR. The server returns only those CIDRs that it thinks that the client can match. For example: the master will return an internal IP CIDR only, if the client reaches the server using an internal IP. Server looks at X-Forwarded-For header or X-Real-Ip header or request.RemoteAddr (in that order) to get the client IP.
versions <br /> *[GroupVersionForDiscovery](#groupversionfordiscovery-v1-meta) array*    | versions are the versions supported in this group.

