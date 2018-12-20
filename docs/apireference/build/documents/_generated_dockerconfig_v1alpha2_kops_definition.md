## DockerConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | DockerConfig



DockerConfig is the configuration for docker

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
authorizationPlugins <br /> *string array*    | AuthorizationPlugins is a list of authorization plugins
bridge <br /> *string*    | Bridge is the network interface containers should bind onto
bridgeIP <br /> *string*    | BridgeIP is a specific IP address and netmask for the docker0 bridge, using standard CIDR notation
defaultUlimit <br /> *string array*    | DefaultUlimit is the ulimits for containers
hosts <br /> *string array*    | Hosts enables you to configure the endpoints the docker daemon listens on i.e. tcp://0.0.0.0.2375 or unix:///var/run/docker.sock etc
insecureRegistry <br /> *string*    | InsecureRegistry enable insecure registry communication @question according to dockers this a list??
ipMasq <br /> *boolean*    | IPMasq enables ip masquerading for containers
ipTables <br /> *boolean*    | IPtables enables addition of iptables rules
liveRestore <br /> *boolean*    | LiveRestore enables live restore of docker when containers are still running
logDriver <br /> *string*    | LogDriver is the default driver for container logs (default "json-file")
logLevel <br /> *string*    | LogLevel is the logging level ("debug", "info", "warn", "error", "fatal") (default "info")
logOpt <br /> *string array*    | Logopt is a series of options given to the log driver options for containers
mtu <br /> *integer*    | MTU is the containers network MTU
registryMirrors <br /> *string array*    | RegistryMirrors is a referred list of docker registry mirror
storage <br /> *string*    | Storage is the docker storage driver to use
storageOpts <br /> *string array*    | StorageOpts is a series of options passed to the storage driver
version <br /> *string*    | Version is consumed by the nodeup and used to pick the docker version

