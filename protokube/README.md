# Protokube

(Note that some/most of this functionality is actually currently in the NodeUp models.  But I think they should
probably be in protokube instead, and will be moving them.)

Protokube acts as a proving ground for code that we will likely want in kubelet, for easy bring-up of a cluster.  As
we are still figuring out some of these things, it would be inefficient and premature to propose them into kubelet
immediately.  However, long-term the hope is that we can move some or all of protokube into kubelet itself.

Protokube has three main roles:

* Mount and discover master volumes
* Configures DNS for simple discovery
* Applies component configuration

## Mount and discover master volumes

Protokube will discovers "master" volumes (containing etcd data) by looking at their tags.

When it finds suitable volumes (same zone) with suitable tags, it tries to mount them, and if successful
creates a manifest so that the component (etcd) will be run by kubelet.  The details of the configuration
are currently encoded into the volume tags; we could also reference the cluster configuration but we haven't
had to do so yet.

## Configures DNS for simple discovery

DNS records are set up:

* For the master's internal IPs, allowing nodes to discover the master
* For the etcd's members internal IPs, allowing for etcd to discover peers without reconfiguration, even as the node
  members move around the IP space (mounting the shared volume)
* For the master's external IPs, allowing kubectl to reach the master (we could also use a load balancer)

Using DNS for etcd seems the easiest way to maintain a [no-ops etcd cluster](https://github.com/coreos/etcd/issues/5418)

## Applies component configuration

The k8s schema defines (in componentconfig) a strongly-typed schema for the configuration of the various components.
Currently this is used only for the components to expose their current configuration.

Protokube extends this to allow the componentconfig schema to be used to write the configuration also.  The only
mechanism available to us right now is the flags mechanism, so protokube will build manifests containing the flags.

(This is currently done by nodeup, but should be moved to protokube)

# Long-term evolution

* Volume discovery, mounting & spawning manifests could be done by kubelet.  It might even be possible to do so
  today by simply creating a manifest that includes a volume mount, although
  kubelet would likely consider a volume that cannot be mounted as a failure state, whereas this is not unexpected
  in available clusters where you might have multiple masters ready to mount the same etcd volume.
  
* DNS configuration should be done by kubelet.

* The individual components should be able to read their configuration from the componentconfig schema.  Thus
  we would simply point e.g. kubelet at a VFS path.
