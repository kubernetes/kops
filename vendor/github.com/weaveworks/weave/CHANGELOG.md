## Release 1.9.4

Bug fixes and minor improvements

* Support Kubernetes 1.6 by creating a new DaemonSet #2777,#2801
* Support Kubernetes 1.6 by allowing CNI callers to send a
  network-delete request for a container that is not running or has
  never been attached to the network #2850
* Leave non-weave ipsets alone in Network Policy Controller (e.g. when
  running Weave Net alonside keepalived-vip) #2751,#2846
* Fix various small issues revealed by 'staticcheck' tool #2843,#2857
* Avoid leaving 'defunct' processes when weave-kube container restarts #2836,#2845
* When using the CNI plugin with a non-standard network configuration
  file, the weave bridge could get the same IP as a container, if
  'weave expose' hadn't run at that point #2839,#2856

Build and Testing

* Check that no defunct processes remain after each test #2852
* Update build and test scripts to work with Kubernetes 1.6 beta #2851

[Full list of changes](https://github.com/weaveworks/weave/milestone/53?closed=1).

## Release 1.9.3

Bug fixes and minor improvements

* Fixed a race condition in Fast Datapath encrypted connections which could lead Weave Net to crash: #2824, #2825.

## Release 1.9.2

Bug fixes and minor improvements

* Fix a weave-kube bug when `br_netfilter` or `xt_set` module is compiled into
  kernel #2820/#2821
* Detect the absence of the required `xt_set` kernel module #2821

## Release 1.9.1

Bug fixes and minor improvements

* Fix a race condition when the Weave Net container is restarted
  which could allow a new container to be allocated the same IP
  address as an existing one #2784,#2787
* Handle the message type received when a pod has been deleted during
  Kubernetes api-server fail-over #2772,#2773
* Make weave-kube work with `dockerd --iptables=false` #2726
* Ensure we have the right kernel modules loaded for Network Policy in weave-kube #2819
* Reference-count addresses in Network Policy Controller, to avoid
  errors when updates come in an unexpected order #2792,#2795
* Allow the soft connection limit to be raised in weave-kube, so
  larger clusters can be created #2781
* WeaveDNS was incorrectly case-sensitive for reverse DNS lookups #2817,#2818

Build and Testing

* Scripts to create VMs to run automated tests were rewritten to use
  Terraform and Ansible, to make it much easier to test with different
  versions of components such as Docker and Kubernetes #2647,#2694,#2775,#2796
* Upgrade to latest Weaveworks common build-tools #2780
* Improve encryption tests #2793
* Update vishvananda/netlink library to bring in changes we had previously forked #2790
* Slight change to the build container to avoid permission errors and slow builds #2761,#2802

## Release 1.9.0

Highlights:

* Encryption is now available for Fast Datapath connections, which
  greatly improves the performance. #1644,#2687
* We now build images for Intel/AMD 64-bit, ARM and ARM 64-bit. #2713
* Weave Net Docker images are now labelled with description, vendor, etc. #2712
* `weave status connections` now shows the MTU #2389,#2663
* CNI plugin is now a stand-alone binary that does not depend on Docker #2594,#2662
* Embedded docker client updated to version 1.10.3 #2395

**NOTE:** The move to multi-architecture required that we update the
  embedded Docker client, and this has the effect that this release of
  Weave Net will not work with Docker installations older than
  1.10.

[Full list of changes](https://github.com/weaveworks/weave/issues?q=milestone%3A1.9.0).

## Release 1.8.2

Bug fixes and minor improvements

* Fixed a bug where looping flows were installed which caused high CPU
  usage #2650, #2674
* Fixed a bug where Kubernetes master could not contact pods #2673, #2683
* Fixed a bug where weave-kube was crashing in a loop due to invalid
  Weave bridge state #2657
* Fixed a bug where iptables NAT rules were not appended due to
  "temporary unavailable" iptables error #2679
* Added a detection of enabling the hairpin mode on the Weave bridge port
  which caused installation of looping flows #2674
* Added a detection of overlaps between Weave and the host IP address
  ranges when launching weave-kube #2669, #2672
* Added logging of connections blocked by weave-npc #2546, #2573

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.8.2).

## Release 1.8.1

Bug fixes and minor improvements

* Fixed weave-npc crash from Succeeded/Failed pods #2632,#2658
* Fixed occasional failure to create Weave bridge on node reboot
  #2617,#2637
* Fixed a bug where weave-kube would fail to install when run with
  unrelease snapshot builds #2642
* Improved conformance to CNI spec by not releasing IP addresses when
  a container dies #2643
* Improved troubleshooting of install failure by creating CNI config
  after Weave Net is up #2570
* "up to date" shown even when the version check was blocked by
  firewall #2537,#2565,#2645
* "Unable to claim" message on re-launching Weave after using CNI
  #2548,#2577
* Eliminated spurious IP reclaim operations when IPAM was disabled
  #2567,#2644
* Include `jq` tool in our build VM configuration #2656

## Release 1.8.0

Highlights:

* Exposed network policy controller Prometheus metrics
  weaveworks/weave-npc#23, #2595, #2549
* Exposed router Prometheus metrics #2535, #2547, #2523, #2579, #2578,
  #2568, #2560, #2561

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.8.0).

## Release 1.7.2

Bug fixes and minor improvements

* Fixed an error where the Docker plugin could fail to attach a
  container with a `bridge "weave" not present` error #2540/#2541
* Fixed `cannot connect to itself` panic on weave launch #2527/#2543
* Fixed inferred initial peer count when target peers includes self
  #2481/#2543
* Fixed compilation on Raspberry Pi #2506/#2538

## Release 1.7.1

Bug fixes and minor improvements

* Added utility to recover container addresses without asking Docker #2531
* Improved CI #2274

## Release 1.7.0

Highlights:

* weave-kube - Deploy Weave Net to Kubernetes with a single command
* weave-npc - Kubernetes network policy controller implementation

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.7.0).

## Release 1.6.2

Bug fixes and minor improvements

* Fixed hang after stopping and restarting on Docker 1.12 #2469/#2502
* Avoid an error on Google container images by checking for tx offload support #2504
* Fixed an issue where the supplied peer list could be ignored when restarting after failure #2503/#2509
* Check for empty peer name on launch #2495/#2501

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.6.2).

## Release 1.6.1

Bug fixes and minor improvements

* `weave ps` was occasionally failing to list allocated addresses of containers #2388/#2418/#2445
* `weave launch[-router]` on 4.2 kernel would appear to succeed even if the fast datapath VXLAN UDP port was in use by a different process #2375/#2474
* Launching the proxy would fail when the Docker daemon could not be detected #2457/#2424
* The CNI plugin did not work with Apache Mesos #2394/#2442
* Router stopped working after a restart in the AWSVPC mode #2381/#2409
* Router crashed when the Docker API endpoint parameter was explicitly set to empty #2421/#2467
* The CNI plugin did not work on recent versions of Docker for Mac #2434/#2442
* The CNI plugin assigns an IP to the bridge if necessary, which avoids failures if `weave expose` has not run yet #2471
* Distinguish peer name collisions from attempts to connect to self in logs #2460
* Improve host clock skew detection message #2174
* Improve the error message returned when executing `weave launch-plugin` without the router running #2293/#2416
* The `create-bridge` subcommand was not enabled in the fast datapath mode #2464/2466
* Allow users to omit `weave setup[-cni]` by initializing the CNI plugin on the `launch[-router]` subcommand #2435/#2442
* Reduce verbosity of fast datapath miss event logs #1852/#2417
* Include the `ipam` option in the help output of the `status` subcommand #2425/#2426
* Remove a harmless duplication of the `--no-dns` parameter #2430
* Internal refactoring
* Improvements to testing and building
* Improvements to the documentation

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.6.1).

## Release 1.6.0

## Highlights

* A new [AWS VPC](https://weave.works/docs/net/latest/using-weave/awsvpc/) mode
  that leverages Amazon Virtual Private Cloud for near-native network
  performance, as an alternative to the Sleeve and Fast Datapath
  overlays
* Docker 1.12 introduced some internal changes that made it
  incompatible with previous version of Weave Net - this version
  restores compatibility
* An [operational guide](https://weave.works/docs/net/latest/operational-guide)
  detailing best practices for deploying and operating Weave Net
* Changes to the target peer list are remembered across restarts,
  making it much easier to deploy resilient networks
* The version checkpoint now transmits network mode (e.g. 'awsvpc')
  and kernel/docker versions to us to inform and guide our development
  efforts. See the [installation documentation](https://weave.works/docs/net/latest/installing-weave/)
  for instructions on disabling the checkpoint feature.

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.6.0).

## Release 1.5.2

Bug fixes and minor improvements

* Weave Proxy did not flush the initial http header in the Docker event stream, which could cause Docker Swarm to show all nodes pending. #2306/#2311
* When using the CNI plugin, if a container was removed and quickly replaced by another using the same IP address, other containers might be unable to contact it. Send an address resolution protocol message to update them. #2313
* Avoid Docker hanging for 1 minute in `weave launch` if the plugin had not shut down cleanly #2286/#2292
* Print an error message when Weave bridge mode is changed without `weave reset` #2304
* Eliminate spurious warning message from IP allocator on plugin shutdown #2300/#2319
* Display error message when address requested in a subnet that is too small (/31 or /32) #2282/#2321
* Add short wait after `weave reset` to allow updates to reach peers #2280
* Weave was occasionally unable to claim existing IP address immediately after launch #2275/#2281
* Refactor some integration tests to run faster and more reliably #2291

## Release 1.5.1

Bug fixes and minor improvements

* Persisted data that was rendered invalid by changing peer name or
  allocation range is detected and removed automatically, preventing
  crashes and hangs #2246/#2209/#2249
* `weave rmpeer` persists the range takeover in case the peer on which
  it was executed dies subsequently #2238
* Launching a container with an explicit `WEAVE_CIDR` in the
  allocation range now waits instead of erroring if the allocator
  hasn't finished initialising #2232/#2265
* Weave DNS now responds to AAAA queries with an empty answer section,
  instead of NXDOMAIN which could be cached and block subsequent
  resolution of A records #2244/#2252
* Many improvements to the documentation.

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.5.1).

## Release 1.5.0

## Highlights

- A new [Container Network Interface](https://github.com/appc/cni#cni---the-container-network-interface) plugin.
- This release is much more robust against unscheduled restarts,
  because it persists key data to disk.
- New configuration options that are useful when you create or
  auto-scale larger networks.
- Weave now periodically checks for updates (can be disabled)

Plus many bug fixes and minor enhancements. More details below and in
the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.5.0).

## Release 1.4.6

Fixes some issues encountered by our users.

* Restarting a peer could leave stale entries in WeaveDNS
  #1867/#2023/#2081
* Weave proxy occasionally failing to attach any new containers
  #2016/#2049

Other fixes:

* Resolved a crash when a restarting peer re-connected to a peer that
  had not received the latest IPAM data #2083/#2092
* Make IP address space available immediately after a dead peer is
  removed #2068

## Release 1.4.5

Higher performance for multicast and broadcast traffic when using
Weave's Fast Datapath

* The flow rule to deliver broadcast and multicast packets in-kernel
  was not created correctly, hence every such packet caused a
  context-switch to the software router #2003/#2008

Other fixes:

* Remove DNS entries for containers that are being restarted by Docker
  but are not live yet #1977/#1989
* Don't let one failing allocation attempt prevent others from
  succeeding; they could be in different subnets which are more
  available #1996/#2001
* Don't complain on second router launch on kernels that lack support
  for Fast Datapath #1929/#1983
* Fix build broken by change in libnetwork IPAM API #1984/#1985

## Release 1.4.4

Fixing a rather serious issue that slipped through our preparations
for Docker 1.10:

* Restarting Docker or rebooting your machine while the Weave Net
  plugin is running causes Docker 1.10.0 to fail to start #1959/#1963

Also one other small fix:

* Avoid a hang when trying to use plugin and proxy at the same time
  #1956/#1961

## Release 1.4.3

Preparing for Docker 1.10, plus some bug-fixes.

* Avoid hang in Docker v1.10 on `docker volume ls` after `weave stop`,
  `weave stop-plugin` or `weave reset` #1934/#1936
* Fix "unexpected EOF" from Docker 1.10 on `docker exec` with Weave
  proxy `--rewrite-inspect` enabled #1911/#1917
* Avoid losing DNS entries and potentially double-allocating IP
  Addresses allocated via plugin, on router restart; also extend
  `weave ps` to show IP addresses allocated via plugin #1745/#1921
* Stop creating lots of copies of `weavewait` program in Docker
  volumes #1757/#1935
* Prevent container starting prematurely when proxy in
`--no-multicast-route` mode #1942/#1943
* Log error message from plugin rather than crashing when weave not
  running #1906/#1918
* Warn, don't error, if unable to remove plugin network in 'weave
  stop', to avoid breaking the usual upgrade or config change process
  #1900/#1919
* Don't crash if network conditions suggest only very small packets
  will get through #1905/#1926
* Cope with unexpected errors during route traversal when starting
  container via proxy #1909/#1910/#1932

## Release 1.4.2

Bug-fixes and minor improvements.

* A race condition in weavewait that would occasionally hang
  containers at startup #1882/#1884
* Having the plugin auto-restart prevents successful `weave launch` on
  reboot #1869
* Work round weave router failure on CoreOS 4.3 caused by kernel bug
  #1854
* `weave launch` would exit with error code on docker <1.9 #1851
* Running `eval $(weave env)` multiple times would break `eval $(weave
  env --restore)` #1824/#1825
* Don't complain in `weave stop` about "Plugin is not running" when
  plugin is not enabled #1840/#1841
* `weave --local launch` would fail if utility program
  `docker_tls_args` could not be found #1844
* Improved error reporting when TLS arg detection fails #1843
* Improve error reporting when docker isn't running #1845
* Add `--trusted-subnets` usage to `weave` script #1842
* weave run can hang under rare combinations of options #1858

## Release 1.4.1

This is a bug-fix release to cover a few issues that came up since the
release of 1.4.0.

* Weave would fail to launch when `$DOCKER_HOST` was set to a TCP socket secured with TLS #1820/#1822
* Weave would fail to stop when run against Docker pre-1.9 #1815/#1817
* Issue a warning instead of failing on `--with-dns` option to proxy, which was removed #1810/#1812
* Make `weave version` show the plugin version #1797/#1813
* Make `weave launch` show when the a container is restarting #1778/#1814
* Make `weave launch` fail if the plugin is running, for consistency with router and proxy. #1818/#1819

More details in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.4.1).

## Release 1.4.0

**Highlights**

* The Docker Network plugin can now operate without a cluster store, so it is now run by default.
* You can now use the fast datapath over trusted links and Weave encryption over untrusted links, in the same network.

Plus many bug fixes and minor enhancements. More details below and in
the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.4.0).

## Release 1.3.1

**Highlights**

* The minimum Docker version has been increased to 1.6 due to the
  [upcoming deprecation of Dockerhub access](https://blog.docker.com/2015/10/docker-hub-deprecation-1-5/)
  for old clients. From December 7th onwards previous versions of the
  `weave` script will fail to pull down images from the hub; if you
  are unable to upgrade to 1.3.1 immediately you can work around this
  by running `weave --local setup` in conjunction with a compatible
  Docker client installation
* Docker networking plugin now works with older kernels and allows you
  to configure the MTU

More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.3.1).

## Release 1.3.0

**Highlights**

This release includes a Docker Plugin, so you have the option to use
Weave Net that way. More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.3.0).

## Release 1.2.1

This release contains a number of bug fixes and minor
enhancements. More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.2.1).

The release is fully compatible with 1.2.0 versions, so existing
clusters can be upgraded incrementally.

## Release 1.2.0

**Highlights**

This release introduces the 
[Fast Data Path](http://docs.weave.works/weave/master/head/features.html#fast-data-path),
which allows Weave networks to operate at near wire-level speeds. This
new feature is enabled by default.

Other highlights:

- auto-configuration of TLS for the Weave Docker API proxy, making it
  easier to run Weave on Macs and in conjunction with tools like
  Docker Swarm
- support for restart policies on application containers and weave
  infrastructure containers
- better compatibility with recent and future Docker versions

More details in the 
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.2.0).

## Release 1.1.2

This release contains a small number of bug fixes. More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.1.2).

The release is fully compatible with 1.1.0 versions, so existing
clusters can be upgraded incrementally. When upgrading from 1.0.x,
note the compatibility information in the
[Installation & Upgrading instructions for Weave 1.1.0](https://github.com/weaveworks/weave/releases/tag/v1.1.0).

## Release 1.1.1

This release contains a number of bug fixes and minor
enhancements. More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.1.1).

The release is fully compatible with 1.1.0 versions, so existing
clusters can be upgraded incrementally. When upgrading from 1.0.x,
note the compatibility information in the
[Installation & Upgrading instructions for Weave 1.1.0](https://github.com/weaveworks/weave/releases/tag/v1.1.0).

## Release v1.1.0

**Highlights**:

- `weave launch` now launches all weave components, simplifying
  startup.
- `weave status` has been completely revamped, with a much improved
  presentation of the information, and the option to select and output
  data in JSON.
- weaveDNS has been rewritten and embedded in the router. The new
  implementation simplifies configuration, improves performance, and
  provides fault resilience for services.
- the weave Docker API proxy now provides an even more seamless user
  experience, and enables easier integration of weave with other
  systems such as kubernetes.
- many usability improvements
- a few minor bug fixes, including a couple of security
  vulnerabilities

More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.1.0).

## Release 1.0.3

This release contains a weaveDNS feature enhancement as well as minor fixes for
improved stability and robustness.

More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.0.3).

The release is fully compatible with other 1.0.x versions, so existing
clusters can be upgraded incrementally.

## Release 1.0.2

This release fixes a number of bugs, including some security
vulnerabilities in the Weave Docker API proxy, hangs and failures in
address allocation, and sporadic failures in name resolution.

More details in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.0.2).

The release is fully compatible with other 1.0.x versions, so existing
clusters can be upgraded incrementally.

## Release 1.0.1

This is a bug fix release, addressing the following issue:

- executing `docker run` from a Docker 1.7 client against the weave
  proxy would report a `no such image` error when the requested image
  wasn't present on the Docker host, instead of downloading it.

## Release 1.0.0

**Highlights**:

- It is now easier than ever to start containers and for them to
  communicate, across multiple hosts. Automatic IP address allocation,
  and name resolution via weaveDNS are now enabled by default, and the
  proxy has become more fully-featured. In short, once weave has been
  launched the following is possible:

      ````
      host1$ docker run --name=pingme -dti ubuntu
      host2$ docker run -ti ubuntu
      root@d11e9287f65b:/# ping pingme
      ````
- Containers can now be
  [load-balanced](http://docs.weave.works/weave/latest_release/weavedns.html#load-balancing)
  easily.
- IP address allocation is now available across [multiple
  subnets](http://docs.weave.works/weave/latest_release/ipam.html#range),
  and hence can be employed when running multiple, isolated
  applications.
- The proxy now supports [TLS
  connections](http://docs.weave.works/weave/latest_release/proxy.html#tls),
  enabling its deployment when the communication between docker
  clients and the server must be secured.

There are many other new features, plus the usual assortment of bug
fixes and improvements under the hood. More detail below and in the
[change log](https://github.com/weaveworks/weave/issues?q=milestone%3A1.0.0).

*NB: This release changes the weave protocol version. Therefore, when
 upgrading an existing installation, all hosts need to be upgraded in
 order to for them to be able to communicate and form a network.*

## Release 0.11.2

This is a bug fix release, addressing the following issues:

- `weave run` did not respect DOCKER_CLIENT_ARGS. #855
- Negative result cache did not expire if requeried within TTL. #845

More details in the [change
log](https://github.com/weaveworks/weave/issues?q=milestone%3A0.11.2).

*NB: This release does not change the weave protocol version.
Therefore, when upgrading an existing 0.11 installation incrementally,
connectivity between peers will be retained.*

## Release 0.11.1

This is a bug fix release, addressing the following issues:

- The IP Allocator could crash in some relatively rare
  circumstances. #782/#783.
- When the proxy failed to attach a container to the weave network,
  there was no failure indication and descriptive error anywhere, and
  the application process would still start. Now an error is reported
  to the client (i.e. typically the Docker CLI), recorded in the proxy
  logs, and the container is terminated. #788/#799.
- `weave launch-proxy --with-ipam` failed to configure the entrypoint
  and DNS unless a (possibly blank) `WEAVE_CIDR` was
  specified. Affected containers could start the application process
  w/o the weave network interface being available, and without
  functioning name resolution for the weaveDNS domain.
  #744/#747/#751/#752
- The `weave status` output for the IP Allocator was misleadingly
  conveying a sense of brokenness when no IP allocation requests had
  been made yet. #787/#801
- When invoking `weave launch-proxy` twice, the second invocation
  would output a blank line and terminate with a success exit
  status. Now it reports that the proxy is already running and exits
  with a non-zero status. #767/#780
- `weave launch-proxy` was not respecting WEAVEPROXY_DOCKER_ARGS, so
  any user-supplied custom configuration for the weaveproxy container
  was ignored. #755/#780
- The proxy was not intercepting API calls to the unversioned (1.0)
  Docker Remote API. Hence none of the weave functionality was
  available when Docker clients were using that version of the
  API. #770/#774/#777,#809
- The proxy would crash when certain elements of the
  `/containers/create` JSON were missing. We are not aware of any
  Docker clients doing this, but it's been fixed regardless. #778/#777

More details in the [change
log](https://github.com/weaveworks/weave/issues?q=milestone%3A0.11.1).

*NB: This release does not change the weave protocol version.
Therefore, when upgrading an existing 0.11 installation incrementally,
connectivity between peers will be retained.*

## Release 0.11.0

**Highlights**:

- **automatic IP Address Management (IPAM)**, which allows application
  containers to be started and attached to the weave network without
  needing to supply an IP address.
- **proxy** for automatically attaching containers started with
  ordinary `docker run`, or the Docker remote API, to the weave
  network.
- ability to **add/remove extra DNS records**.
- performance and scalability improvements
- fixes for a small number of bugs discovered during testing

More detail in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A0.11.0)

*NB: This release changes the weave protocol version. Therefore, when
upgrading an existing installation, all hosts need to be upgraded in
order to for them to be able to communicate and form a network.*

## Release 0.10.0

**Highlights**:

- **bug fixes**, in particular eradicating several deadlocks that could
  cause peer connection failures and hangs.
- **performance and scalability improvements** to the weave control plane,
  as a result of which it is now possible to construct much larger
  weave networks than previously.
- **improved installation and administration**, particularly the
  introduction of remote execution of the weave script in a container,
  permitting fully containerised deployment of weave.
- **improved diagnostics**, such as the reporting of connection failures
  in `weave status`.
- **new weaveDNS features**, most notably the caching of DNS records
  for application containers, which makes finding container IP
  addresses via weaveDNS much faster.

More detail in the [change log](https://github.com/weaveworks/weave/issues?q=milestone%3A0.10.0)

*NB: This release changes the Weave Docker image names. To upgrade
from an older version, 1) stop all application containers, 2) run
`weave reset` from the old version to remove all traces of weave, and
only then 3) install the new version.*

## Release 0.9.0

- Improve WeaveDNS to the point where it can act as the name server
  for containers in nearly all situations.

- Diagnose and report peer connectivity more comprehensively.

- Adapt to changes in topology - adding & removing of weave peers,
  disruption of connectivity - more rapidly.

- Cope with delays in downloading/running docker images/containers
  required for weave operation.

See the
[complete change log](https://github.com/weaveworks/weave/issues?q=milestone%3A0.9.0)
for more details.

## Release 0.8.0

- Align script and image version. When the `weave` script has a
  version number, i.e. it is part of an official release, it runs
  docker images matching that version. Thus the script and image
  versions are automatically aligned. Unversioned/unreleased `weave`
  scripts run the 'latest'-tagged image versions.

- Eliminate dependency on ethtool and conntrack. Instead of requiring
  these to be installed on the host, weave invokes them via a
  `weavetools` docker image that contains minimally packaged versions
  of these utilities.

- New `weave setup` command. This downloads all docker images used by
  weave. Invoking this is strictly optional; its main purpose is to
  facilitate automated installation of weave and preventing delays in
  subsequent weave command execution due to image downloading.

## Release 0.7.0

This is the first release assigned a version number.

When downloading weave you now have the following choices...

1. a specific version, e.g. https://github.com/weaveworks/weave/releases/download/v0.7.0/weave
2. latest released version: https://github.com/weaveworks/weave/releases/download/latest_release/weave
3. most recent 'master' commit: https://raw.githubusercontent.com/weaveworks/weave/master/weave

Previously the only documented download location was (3). We recommend
that any automated scripts using that be changed to either (1) or (2).
