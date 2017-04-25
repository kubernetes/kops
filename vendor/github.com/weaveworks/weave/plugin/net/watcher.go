package plugin

import (
	"fmt"

	weaveapi "github.com/weaveworks/weave/api"
	"github.com/weaveworks/weave/common/docker"
	weavenet "github.com/weaveworks/weave/net"
)

const (
	WeaveDomain = "weave.local"
)

type watcher struct {
	client *docker.Client
	weave  *weaveapi.Client
	driver *driver
}

type Watcher interface {
}

func NewWatcher(client *docker.Client, weave *weaveapi.Client, driver *driver) (Watcher, error) {
	w := &watcher{client: client, weave: weave, driver: driver}
	return w, client.AddObserver(w)
}

func (w *watcher) ContainerStarted(id string) {
	w.driver.debug("ContainerStarted", "%s", id)
	info, err := w.client.InspectContainer(id)
	if err != nil {
		w.driver.warn("ContainerStarted", "error inspecting container %s: %s", id, err)
		return
	}
	// check that it's on our network
	for _, net := range info.NetworkSettings.Networks {
		network, err := w.driver.findNetworkInfo(net.NetworkID)
		if err != nil {
			w.driver.warn("ContainerStarted", "unable to find network %s info: %s", net.NetworkID, err)
			continue
		}
		if network.isOurs {
			if w.driver.dns {
				fqdn := fmt.Sprintf("%s.%s", info.Config.Hostname, info.Config.Domainname)
				if err := w.weave.RegisterWithDNS(id, fqdn, net.IPAddress); err != nil {
					w.driver.warn("ContainerStarted", "unable to register %s with weaveDNS: %s", id, err)
				}
			}
			rootDir := "/"
			if w.driver.isPluginV2 {
				// We bind mount host's /proc to /host/proc for plugin-v2
				rootDir = "/host"
			}
			netNSPath := weavenet.NSPathByPidWithRoot(rootDir, info.State.Pid)
			if _, err := weavenet.WithNetNS(netNSPath, "configure-arp", weavenet.VethName); err != nil {
				w.driver.warn("ContainerStarted", "unable to configure interfaces: %s", err)
			}
		}
	}
}

func (w *watcher) ContainerDied(id string) {
	// don't need to do this as WeaveDNS removes names on container died anyway
	// (note by the time we get this event we can't see the EndpointID)
}

func (w *watcher) ContainerDestroyed(id string) {}
