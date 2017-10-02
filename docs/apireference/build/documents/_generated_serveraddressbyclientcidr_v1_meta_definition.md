## ServerAddressByClientCIDR v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | ServerAddressByClientCIDR



ServerAddressByClientCIDR helps the client to determine the server address that they should use, depending on the clientCIDR that they match.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#apigroup-v1-meta">APIGroup meta/v1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
clientCIDR <br /> *string*    | The CIDR with which clients can match their IP to figure out the server address that they should use.
serverAddress <br /> *string*    | Address of this server, suitable for a client that matches the above CIDR. This can be a hostname, hostname:port, IP or IP:port.

