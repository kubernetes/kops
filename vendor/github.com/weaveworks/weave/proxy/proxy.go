package proxy

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	weaveapi "github.com/weaveworks/weave/api"
	weavedocker "github.com/weaveworks/weave/common/docker"
	weavenet "github.com/weaveworks/weave/net"
	"github.com/weaveworks/weave/net/address"
)

const (
	defaultCaFile   = "ca.pem"
	defaultKeyFile  = "key.pem"
	defaultCertFile = "cert.pem"

	weaveSock     = "/var/run/weave/weave.sock"
	weaveSockUnix = "unix://" + weaveSock

	initialInterval = 2 * time.Second
	maxInterval     = 1 * time.Minute
)

var (
	containerCreateRegexp  = dockerAPIEndpoint("containers/create")
	containerStartRegexp   = dockerAPIEndpoint("containers/[^/]*/(re)?start")
	containerInspectRegexp = dockerAPIEndpoint("containers/[^/]*/json")
	execCreateRegexp       = dockerAPIEndpoint("containers/[^/]*/exec")
	execInspectRegexp      = dockerAPIEndpoint("exec/[^/]*/json")

	ErrWeaveCIDRNone = errors.New("the container was created with the '-e WEAVE_CIDR=none' option")
	ErrNoDefaultIPAM = errors.New("the container was created without specifying an IP address with '-e WEAVE_CIDR=...' and the proxy was started with the '--no-default-ipalloc' option")
)

func dockerAPIEndpoint(endpoint string) *regexp.Regexp {
	return regexp.MustCompile("^(/v[0-9\\.]*)?/" + endpoint + "$")
}

type Config struct {
	HostnameFromLabel   string
	HostnameMatch       string
	HostnameReplacement string
	Image               string
	ListenAddrs         []string
	RewriteInspect      bool
	NoDefaultIPAM       bool
	NoRewriteHosts      bool
	TLSConfig           TLSConfig
	Version             string
	WithoutDNS          bool
	NoMulticastRoute    bool
	DockerBridge        string
	DockerHost          string
}

type wait struct {
	ident string
	ch    chan error
	done  bool
}

type Proxy struct {
	sync.Mutex
	Config
	client                 *docker.Client
	dockerBridgeIP         string
	hostnameMatchRegexp    *regexp.Regexp
	weaveWaitVolume        string
	weaveWaitNoopVolume    string
	weaveWaitNomcastVolume string
	normalisedAddrs        []string
	waiters                map[*http.Request]*wait
	attachJobs             map[string]*attachJob
	quit                   chan struct{}
}

type attachJob struct {
	id          string
	tryInterval time.Duration // retry delay on next failure
	timer       *time.Timer
}

func (proxy *Proxy) attachWithRetry(id string) {
	proxy.Lock()
	defer proxy.Unlock()
	if j, ok := proxy.attachJobs[id]; ok {
		j.timer.Reset(time.Duration(0))
		return
	}

	j := &attachJob{id: id, tryInterval: initialInterval}
	j.timer = time.AfterFunc(time.Duration(0), func() {
		if err := proxy.attach(id); err != nil {
			// The delay at the nth retry is a random value in the range
			// [i-i/2,i+i/2], where i = initialInterval * 1.5^(n-1).
			j.timer.Reset(j.tryInterval)
			j.tryInterval = j.tryInterval * 3 / 2
			if j.tryInterval > maxInterval {
				j.tryInterval = maxInterval
			}
			return
		}
		proxy.notifyWaiters(id, nil)
	})
	proxy.attachJobs[id] = j
}

func (j attachJob) Stop() {
	j.timer.Stop()
}

func NewProxy(c Config) (*Proxy, error) {
	p := &Proxy{
		Config:     c,
		waiters:    make(map[*http.Request]*wait),
		attachJobs: make(map[string]*attachJob),
		quit:       make(chan struct{}),
	}

	if err := p.TLSConfig.LoadCerts(); err != nil {
		Log.Fatalf("Could not configure tls for proxy: %s", err)
	}

	// We pin the protocol version to 1.18 (which corresponds to
	// Docker 1.6.x; the earliest version supported by weave) in order
	// to insulate ourselves from breaking changes to the API, as
	// happened in 1.20 (Docker 1.8.0) when the presentation of
	// volumes changed in `inspect`.
	client, err := weavedocker.NewVersionedClient(c.DockerHost, "1.18")
	if err != nil {
		return nil, err
	}
	Log.Info(client.Info())

	p.client = client.Client

	if !p.WithoutDNS {
		ip, err := weavenet.FindBridgeIP(c.DockerBridge, nil)
		if err != nil {
			return nil, err
		}
		p.dockerBridgeIP = ip.String()
		Log.Infof("Using docker bridge IP for DNS: %v", p.dockerBridgeIP)
	}

	p.hostnameMatchRegexp, err = regexp.Compile(c.HostnameMatch)
	if err != nil {
		err := fmt.Errorf("Incorrect hostname match '%s': %s", c.HostnameMatch, err.Error())
		return nil, err
	}

	if err = p.findWeaveWaitVolumes(); err != nil {
		return nil, err
	}

	client.AddObserver(p)

	return p, nil
}

func (proxy *Proxy) AttachExistingContainers() {
	containers, _ := proxy.client.ListContainers(docker.ListContainersOptions{})
	for _, c := range containers {
		proxy.attachWithRetry(c.ID)
	}
}

func (proxy *Proxy) Dial() (net.Conn, error) {
	proto := "tcp"
	addr := proxy.Config.DockerHost
	switch {
	case strings.HasPrefix(addr, "unix://"):
		proto = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
	case strings.HasPrefix(addr, "tcp://"):
		addr = strings.TrimPrefix(addr, "tcp://")
	}
	return net.Dial(proto, addr)
}

func (proxy *Proxy) findWeaveWaitVolumes() error {
	var err error
	if proxy.weaveWaitVolume, err = proxy.findVolume("/w"); err != nil {
		return err
	}
	if proxy.weaveWaitNoopVolume, err = proxy.findVolume("/w-noop"); err != nil {
		return err
	}
	proxy.weaveWaitNomcastVolume, err = proxy.findVolume("/w-nomcast")
	return err
}

func (proxy *Proxy) findVolume(v string) (string, error) {
	container, err := proxy.client.InspectContainer("weaveproxy")
	if err != nil {
		return "", fmt.Errorf("Could not find the weavewait volume: %s", err)
	}

	if container.Volumes == nil {
		return "", fmt.Errorf("Could not find the weavewait volume")
	}

	volume, ok := container.Volumes[v]
	if !ok {
		return "", fmt.Errorf("Could not find the weavewait volume")
	}

	return volume, nil
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Log.Infof("%s %s", r.Method, r.URL)
	path := r.URL.Path
	var i interceptor
	switch {
	case containerCreateRegexp.MatchString(path):
		i = &createContainerInterceptor{proxy}
	case containerStartRegexp.MatchString(path):
		i = &startContainerInterceptor{proxy}
	case containerInspectRegexp.MatchString(path):
		i = &inspectContainerInterceptor{proxy}
	case execCreateRegexp.MatchString(path):
		i = &createExecInterceptor{proxy}
	case execInspectRegexp.MatchString(path):
		i = &inspectExecInterceptor{proxy}
	default:
		i = &nullInterceptor{}
	}
	proxy.Intercept(i, w, r)
}

func (proxy *Proxy) Listen() []net.Listener {
	listeners := []net.Listener{}
	proxy.normalisedAddrs = []string{}
	unixAddrs := []string{}
	for _, addr := range proxy.ListenAddrs {
		if strings.HasPrefix(addr, "unix://") || strings.HasPrefix(addr, "/") {
			unixAddrs = append(unixAddrs, addr)
			continue
		}
		listener, normalisedAddr, err := proxy.listen(addr)
		if err != nil {
			Log.Fatalf("Cannot listen on %s: %s", addr, err)
		}
		listeners = append(listeners, listener)
		proxy.normalisedAddrs = append(proxy.normalisedAddrs, normalisedAddr)
	}

	if len(unixAddrs) > 0 {
		listener, _, err := proxy.listen(weaveSockUnix)
		if err != nil {
			Log.Fatalf("Cannot listen on %s: %s", weaveSockUnix, err)
		}
		listeners = append(listeners, listener)

		if err := proxy.symlink(unixAddrs); err != nil {
			Log.Fatalf("Cannot listen on unix sockets: %s", err)
		}

		proxy.normalisedAddrs = append(proxy.normalisedAddrs, weaveSockUnix)
	}

	for _, addr := range proxy.normalisedAddrs {
		Log.Infoln("proxy listening on", addr)
	}
	return listeners
}

func (proxy *Proxy) Serve(listeners []net.Listener) {
	errs := make(chan error)
	for _, listener := range listeners {
		go func(listener net.Listener) {
			errs <- (&http.Server{Handler: proxy}).Serve(listener)
		}(listener)
	}
	for range listeners {
		err := <-errs
		if err != nil {
			Log.Fatalf("Serve failed: %s", err)
		}
	}
}

func (proxy *Proxy) ListenAndServeStatus(socket string) {
	listener, err := weavenet.ListenUnixSocket(socket)
	if err != nil {
		Log.Fatalf("ListenAndServeStatus failed: %s", err)
	}
	handler := http.HandlerFunc(proxy.StatusHTTP)
	if err := (&http.Server{Handler: handler}).Serve(listener); err != nil {
		Log.Fatalf("ListenAndServeStatus failed: %s", err)
	}
}

func (proxy *Proxy) StatusHTTP(w http.ResponseWriter, r *http.Request) {
	for _, addr := range proxy.normalisedAddrs {
		fmt.Fprintln(w, addr)
	}
}

func copyOwnerAndPermissions(from, to string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return err
	}
	if err = os.Chmod(to, stat.Mode()); err != nil {
		return err
	}

	moreStat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}

	if err = os.Chown(to, int(moreStat.Uid), int(moreStat.Gid)); err != nil {
		return err
	}

	return nil
}

func (proxy *Proxy) listen(protoAndAddr string) (net.Listener, string, error) {
	var (
		listener    net.Listener
		err         error
		proto, addr string
	)

	if protoAddrParts := strings.SplitN(protoAndAddr, "://", 2); len(protoAddrParts) == 2 {
		proto, addr = protoAddrParts[0], protoAddrParts[1]
	} else if strings.HasPrefix(protoAndAddr, "/") {
		proto, addr = "unix", protoAndAddr
	} else {
		proto, addr = "tcp", protoAndAddr
	}

	switch proto {
	case "tcp":
		listener, err = net.Listen(proto, addr)
		if err != nil {
			return nil, "", err
		}
		if proxy.TLSConfig.IsEnabled() {
			listener = tls.NewListener(listener, proxy.TLSConfig.Config)
		}

	case "unix":
		// remove socket from last invocation
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return nil, "", err
		}
		listener, err = net.Listen(proto, addr)
		if err != nil {
			return nil, "", err
		}
		if strings.HasPrefix(proxy.Config.DockerHost, "unix://") {
			if err = copyOwnerAndPermissions(strings.TrimPrefix(proxy.Config.DockerHost, "unix://"), addr); err != nil {
				return nil, "", err
			}
		}

	default:
		Log.Fatalf("Invalid protocol format: %q", proto)
	}

	return &MalformedHostHeaderOverride{listener}, fmt.Sprintf("%s://%s", proto, addr), nil
}

// weavedocker.ContainerObserver interface
func (proxy *Proxy) ContainerStarted(ident string) {
	err := proxy.attach(ident)
	if err != nil {
		var e error
		// attach failed: if we have a request waiting on the start, kill the container,
		// otherwise assume it is a Docker-initated restart and kill the process inside.
		if proxy.waitChan(ident) != nil {
			e = proxy.client.KillContainer(docker.KillContainerOptions{ID: ident})
		} else {
			var c *docker.Container
			if c, e = proxy.client.InspectContainer(ident); e == nil {
				var process *os.Process
				if process, e = os.FindProcess(c.State.Pid); e == nil {
					e = process.Kill()
				}
			}
		}
		if e != nil {
			Log.Warningf("Error killing %s: %s", ident, e)
		}
	}
	proxy.notifyWaiters(ident, err)
}

func containerShouldAttach(container *docker.Container) bool {
	return len(container.Config.Entrypoint) > 0 && container.Config.Entrypoint[0] == weaveWaitEntrypoint[0]
}

func containerIsWeaveRouter(container *docker.Container) bool {
	return container.Name == weaveContainerName &&
		len(container.Config.Entrypoint) > 0 && container.Config.Entrypoint[0] == weaveEntrypoint
}

func (proxy *Proxy) createWait(r *http.Request, ident string) {
	proxy.Lock()
	proxy.waiters[r] = &wait{ident: ident, ch: make(chan error, 1)}
	proxy.Unlock()
}

func (proxy *Proxy) removeWait(r *http.Request) {
	proxy.Lock()
	delete(proxy.waiters, r)
	proxy.Unlock()
}

func (proxy *Proxy) notifyWaiters(ident string, err error) {
	proxy.Lock()
	if j, ok := proxy.attachJobs[ident]; ok {
		j.Stop()
		delete(proxy.attachJobs, ident)
	}
	for _, wait := range proxy.waiters {
		if ident == wait.ident && !wait.done {
			wait.ch <- err
			close(wait.ch)
			wait.done = true
		}
	}
	proxy.Unlock()
}

func (proxy *Proxy) waitForStart(r *http.Request) error {
	var ch chan error
	proxy.Lock()
	wait, found := proxy.waiters[r]
	if found {
		ch = wait.ch
	}
	proxy.Unlock()
	if ch != nil {
		Log.Debugf("Wait for start of container %s", wait.ident)
		return <-ch
	}
	return nil
}

func (proxy *Proxy) waitChan(ident string) chan error {
	proxy.Lock()
	defer proxy.Unlock()
	for _, wait := range proxy.waiters {
		if ident == wait.ident && !wait.done {
			return wait.ch
		}
	}
	return nil
}

// If some other operation is waiting for a container to start, join in the wait
func (proxy *Proxy) waitForStartByIdent(ident string) error {
	if ch := proxy.waitChan(ident); ch != nil {
		Log.Debugf("Wait for start of container %s", ident)
		return <-ch
	}
	return nil
}

func (proxy *Proxy) ContainerDied(ident string)      {}
func (proxy *Proxy) ContainerDestroyed(ident string) {}

// Check if this container needs to be attached, if so then attach it,
// and return nil on success or not needed.
func (proxy *Proxy) attach(containerID string) error {
	container, err := proxy.client.InspectContainer(containerID)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			Log.Warningf("unable to attach existing container %s since inspecting it failed: %v", containerID, err)
		}
		return nil
	}
	if containerIsWeaveRouter(container) {
		Log.Infof("Attaching weave router container: %s", container.ID)
		return callWeaveAttach(container, []string{"attach-router"})
	}
	if !containerShouldAttach(container) || !container.State.Running {
		return nil
	}

	cidrs, err := proxy.weaveCIDRs(container.HostConfig.NetworkMode, container.Config.Env)
	if err != nil {
		Log.Infof("Leaving container %s alone because %s", containerID, err)
		return nil
	}
	Log.Infof("Attaching container %s with WEAVE_CIDR \"%s\" to weave network", container.ID, strings.Join(cidrs, " "))
	if err := validateCIDRs(cidrs); err != nil {
		return err
	}

	args := []string{"attach"}
	args = append(args, cidrs...)
	if !proxy.NoRewriteHosts {
		args = append(args, "--rewrite-hosts")

		if container.HostConfig != nil {
			for _, eh := range container.HostConfig.ExtraHosts {
				args = append(args, fmt.Sprintf("--add-host=%s", eh))
			}
		}
	}
	if proxy.NoMulticastRoute {
		args = append(args, "--no-multicast-route")
	}
	args = append(args, container.ID)
	return callWeaveAttach(container, args)
}

func callWeaveAttach(container *docker.Container, args []string) error {
	if _, stderr, err := callWeave(args...); err != nil {
		Log.Warningf("Attaching container %s to weave network failed: %s", container.ID, string(stderr))
		return errors.New(string(stderr))
	} else if len(stderr) > 0 {
		Log.Warningf("Attaching container %s to weave network: %s", container.ID, string(stderr))
	}
	return nil
}

func validateCIDRs(cidrs []string) error {
	for _, cidr := range cidrs {
		if cidr == "net:default" {
			continue
		}
		for _, prefix := range []string{"ip:", "net:", ""} {
			if strings.HasPrefix(cidr, prefix) {
				if _, err := address.ParseCIDR(strings.TrimPrefix(cidr, prefix)); err == nil {
					break
				}
				return fmt.Errorf("invalid WEAVE_CIDR: %s", cidr)
			}
		}
	}
	return nil
}

func (proxy *Proxy) weaveCIDRs(networkMode string, env []string) ([]string, error) {
	if networkMode == "host" || strings.HasPrefix(networkMode, "container:") ||
		// Anything else, other than blank/none/default/bridge, is some sort of network plugin
		(networkMode != "" && networkMode != "none" && networkMode != "default" && networkMode != "bridge") {
		return nil, fmt.Errorf("the container has '--net=%s'", networkMode)
	}
	for _, e := range env {
		if strings.HasPrefix(e, "WEAVE_CIDR=") {
			if e[11:] == "none" {
				return nil, ErrWeaveCIDRNone
			}
			return strings.Fields(e[11:]), nil
		}
	}
	if proxy.NoDefaultIPAM {
		return nil, ErrNoDefaultIPAM
	}
	return nil, nil
}

func (proxy *Proxy) setWeaveDNS(hostConfig jsonObject, hostname, dnsDomain string) error {
	dns, err := hostConfig.StringArray("Dns")
	if err != nil {
		return err
	}
	hostConfig["Dns"] = append(dns, proxy.dockerBridgeIP)

	dnsSearch, err := hostConfig.StringArray("DnsSearch")
	if err != nil {
		return err
	}
	if len(dnsSearch) == 0 {
		if hostname == "" {
			hostConfig["DnsSearch"] = []string{dnsDomain}
		} else {
			hostConfig["DnsSearch"] = []string{"."}
		}
	}

	return nil
}

func (proxy *Proxy) getDNSDomain() string {
	if proxy.WithoutDNS {
		return ""
	}
	weave := weaveapi.NewClient(os.Getenv("WEAVE_HTTP_ADDR"), Log)
	domain, _ := weave.DNSDomain()
	return domain
}

func (proxy *Proxy) updateContainerNetworkSettings(container jsonObject) error {
	containerID, err := container.String("Id")
	if err != nil {
		return err
	}

	state, err := container.Object("State")
	if err != nil {
		return err
	}

	pid, err := state.Int("Pid")
	if err != nil {
		return err
	}

	if err := proxy.waitForStartByIdent(containerID); err != nil {
		return err
	}
	netDevs, err := weavenet.GetWeaveNetDevs(pid)
	if err != nil || len(netDevs) == 0 || len(netDevs[0].CIDRs) == 0 {
		return err
	}

	networkSettings, err := container.Object("NetworkSettings")
	if err != nil {
		return err
	}
	networkSettings["MacAddress"] = netDevs[0].MAC.String()
	networkSettings["IPAddress"] = netDevs[0].CIDRs[0].IP.String()
	networkSettings["IPPrefixLen"], _ = netDevs[0].CIDRs[0].Mask.Size()
	return nil
}

func (proxy *Proxy) symlink(unixAddrs []string) (err error) {
	var container *docker.Container
	binds := []string{"/var/run/weave:/var/run/weave"}
	froms := []string{}
	for _, addr := range unixAddrs {
		from := strings.TrimPrefix(addr, "unix://")
		if from == weaveSock {
			continue
		}
		dir := filepath.Dir(from)
		binds = append(binds, dir+":"+filepath.Join("/host", dir))
		froms = append(froms, filepath.Join("/host", from))
		proxy.normalisedAddrs = append(proxy.normalisedAddrs, addr)
	}
	if len(froms) == 0 {
		return
	}

	env := []string{
		"PATH=/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	if val := os.Getenv("WEAVE_DEBUG"); val != "" {
		env = append(env, fmt.Sprintf("%s=%s", "WEAVE_DEBUG", val))
	}

	container, err = proxy.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      proxy.Image,
			Entrypoint: []string{"/home/weave/symlink", weaveSock},
			Cmd:        froms,
			Env:        env,
		},
		HostConfig: &docker.HostConfig{Binds: binds},
	})
	if err != nil {
		return
	}

	defer func() {
		err2 := proxy.client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})
		if err == nil {
			err = err2
		}
	}()

	err = proxy.client.StartContainer(container.ID, nil)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = proxy.client.AttachToContainer(docker.AttachToContainerOptions{
		Container:   container.ID,
		ErrorStream: &buf,
		Logs:        true,
		Stderr:      true,
	})
	if err != nil {
		return
	}

	var rc int
	rc, err = proxy.client.WaitContainer(container.ID)
	if err != nil {
		return
	}
	if rc != 0 {
		err = errors.New(buf.String())
	}
	return
}

func (proxy *Proxy) Stop() {
	close(proxy.quit)
	proxy.Lock()
	defer proxy.Unlock()
	for _, j := range proxy.attachJobs {
		j.Stop()
	}
}
