package docker

import (
	"errors"
	"fmt"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/weave/common"
)

const (
	InitialInterval = 1 * time.Second
	MaxInterval     = 20 * time.Second
)

// An observer for container events
type ContainerObserver interface {
	ContainerStarted(ident string)
	ContainerDied(ident string)
	ContainerDestroyed(ident string)
}

type Client struct {
	*docker.Client
}

type syncPair struct {
	stop chan struct{}
	done chan struct{}
}

type pendingStarts map[string]*syncPair

// NewClient creates a new Docker client and checks we can talk to Docker
func NewClient(apiPath string) (*Client, error) {
	if apiPath != "" && !strings.Contains(apiPath, "://") {
		apiPath = "tcp://" + apiPath
	}
	dc, err := docker.NewClient(apiPath)
	if err != nil {
		return nil, err
	}
	client := &Client{dc}

	return client, client.checkWorking()
}

func NewVersionedClient(apiPath string, apiVersionString string) (*Client, error) {
	if !strings.Contains(apiPath, "://") {
		apiPath = "tcp://" + apiPath
	}
	dc, err := docker.NewVersionedClient(apiPath, apiVersionString)
	if err != nil {
		return nil, err
	}
	client := &Client{dc}

	return client, client.checkWorking()
}

func NewVersionedClientFromEnv(apiVersionString string) (*Client, error) {
	dc, err := docker.NewVersionedClientFromEnv(apiVersionString)
	if err != nil {
		return nil, err
	}
	client := &Client{dc}

	return client, client.checkWorking()
}

func (c *Client) checkWorking() error {
	_, err := c.Version()
	return err
}

func (c *Client) Info() string {
	env, err := c.Version()
	if err != nil {
		return fmt.Sprintf("Docker API error: %s", err)
	}
	return fmt.Sprintf("Docker API on %s: %v", c.Endpoint(), env)
}

func (c *Client) DockerVersion() string {
	if env, err := c.Version(); err == nil {
		if v, found := env.Map()["Version"]; found {
			return v
		}
	}
	return "unknown"
}

// AddObserver adds an observer for docker events
func (c *Client) AddObserver(ob ContainerObserver) error {
	go func() {
		pending := make(pendingStarts)
		retryInterval := InitialInterval
		for {
			events := make(chan *docker.APIEvents)
			if err := c.AddEventListener(events); err != nil {
				c.errorf("Unable to add listener to Docker API: %s - retrying in %ds", err, retryInterval/time.Second)
			} else {
				start := time.Now()
				for event := range events {
					switch event.Status {
					case "start":
						pending.finish(event.ID)
						pending.start(event.ID, c, ob)
					case "die":
						pending.finish(event.ID)
						ob.ContainerDied(event.ID)
					case "destroy":
						pending.finish(event.ID)
						ob.ContainerDestroyed(event.ID)
					}
				}
				if time.Since(start) > retryInterval {
					retryInterval = InitialInterval
				}
				c.errorf("Event listener channel closed - retrying subscription in %ds", retryInterval/time.Second)
			}
			time.Sleep(retryInterval)
			retryInterval = retryInterval * 3 / 2
			if retryInterval > MaxInterval {
				retryInterval = MaxInterval
			}
		}
	}()
	return nil
}

// Docker sends a 'start' event before it has attempted to start the
// container.  Delay notifying the observer until the container has a
// pid, or we are told to stop when a 'die' event arrives.
//
// Note we always deliver the event, even if the container seems to
// have gone away.
func (pending pendingStarts) start(id string, c *Client, ob ContainerObserver) {
	sync := syncPair{make(chan struct{}), make(chan struct{})}
	pending[id] = &sync
	go func() {
		defer close(sync.done)
		defer ob.ContainerStarted(id)
		for {
			if container, err := c.InspectContainer(id); err != nil || container.State.Pid != 0 {
				return
			}
			select {
			case <-sync.stop:
				return
			case <-time.After(200 * time.Millisecond):
			}
		}
	}()
}

func (pending pendingStarts) finish(id string) {
	if sync, found := pending[id]; found {
		close(sync.stop)
		<-sync.done
		delete(pending, id)
	}
}

// AllContainerIDs returns all the IDs of Docker containers,
// whether they are running or not.
func (c *Client) AllContainerIDs() ([]string, error) {
	return c.containerIDs(true)
}

// RunningContainerIDs returns all the IDs of the running
// Docker containers.
func (c *Client) RunningContainerIDs() ([]string, error) {
	return c.containerIDs(false)
}

func (c *Client) containerIDs(all bool) ([]string, error) {
	containers, err := c.ListContainers(docker.ListContainersOptions{All: all})
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, c := range containers {
		ids = append(ids, c.ID)
	}
	return ids, nil
}

// IsContainerNotRunning returns true if we have checked with Docker that the ID is not running
func (c *Client) IsContainerNotRunning(idStr string) bool {
	container, err := c.InspectContainer(idStr)
	if err == nil {
		return !container.State.Running || container.State.Restarting
	}
	if _, notThere := err.(*docker.NoSuchContainer); notThere {
		return true
	}
	c.errorf("Could not check container status: %s", err)
	return false
}

// This is intended to find an IP address that we can reach the container on;
// if it is on the Docker bridge network then that address; if on the host network
// then localhost
func (c *Client) GetContainerIP(nameOrID string) (string, error) {
	info, err := c.InspectContainer(nameOrID)
	if err != nil {
		return "", err
	}
	if info.NetworkSettings.Networks != nil {
		if bridgeNetwork, ok := info.NetworkSettings.Networks["bridge"]; ok {
			return bridgeNetwork.IPAddress, nil
		} else if _, ok := info.NetworkSettings.Networks["host"]; ok {
			return "127.0.0.1", nil
		}
	} else if info.HostConfig.NetworkMode == "host" {
		return "127.0.0.1", nil
	}
	if info.NetworkSettings.IPAddress == "" {
		return "", errors.New("No IP address found for container " + nameOrID)
	}
	return info.NetworkSettings.IPAddress, nil
}

// logging

func (c *Client) errorf(fmt string, args ...interface{}) {
	common.Log.Errorf("[docker] "+fmt, args...)
}
