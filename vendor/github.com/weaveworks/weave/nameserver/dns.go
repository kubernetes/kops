package nameserver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/weaveworks/weave/net/address"
)

const (
	topDomain        = "."
	reverseDNSdomain = "in-addr.arpa."
	udpBuffSize      = uint16(4096)
	minUDPSize       = 512

	DefaultListenAddress = "0.0.0.0:53"
	DefaultTTL           = 1
	DefaultClientTimeout = 5 * time.Second
)

type Upstream interface {
	Config() (*dns.ClientConfig, error)
}

func NewUpstream(resolvConf, filterAddress string) Upstream {
	return &upstream{
		resolvConf:    resolvConf,
		filterAddress: filterAddress,
		cachedConfig:  &dns.ClientConfig{}}
}

type upstream struct {
	sync.Mutex

	resolvConf    string
	filterAddress string
	cachedConfig  *dns.ClientConfig
	lastStat      time.Time
	lastModified  time.Time
}

func (u *upstream) Config() (*dns.ClientConfig, error) {
	u.Lock()
	defer u.Unlock()

	// If we checked less than five seconds ago return what we already have
	now := time.Now()
	if now.Sub(u.lastStat) < 5*time.Second {
		return u.cachedConfig, nil
	}
	u.lastStat = now

	// Reread file if it has changed
	if fi, err := os.Stat(u.resolvConf); err != nil {
		return u.cachedConfig, err
	} else if fi.ModTime() != u.lastModified {
		if config, err := dns.ClientConfigFromFile(u.resolvConf); err == nil {
			config.Servers = filter(config.Servers, u.filterAddress)
			u.lastModified = fi.ModTime()
			u.cachedConfig = config
		} else {
			return u.cachedConfig, err
		}
	}

	return u.cachedConfig, nil
}

func filter(ss []string, s string) []string {
	for i := 0; i < len(ss); {
		if ss[i] == s {
			ss = append(ss[:i], ss[i+1:]...)
			continue
		}
		i++
	}
	return ss
}

type DNSServer struct {
	ns      *Nameserver
	domain  string
	ttl     uint32
	address string

	servers   []*dns.Server
	upstream  Upstream
	tcpClient *dns.Client
	udpClient *dns.Client
}

func NewDNSServer(ns *Nameserver, domain, address string, upstream Upstream, ttl uint32, clientTimeout time.Duration) (*DNSServer, error) {
	s := &DNSServer{
		ns:        ns,
		domain:    dns.Fqdn(domain),
		ttl:       ttl,
		address:   address,
		upstream:  upstream,
		tcpClient: &dns.Client{Net: "tcp", ReadTimeout: clientTimeout},
		udpClient: &dns.Client{Net: "udp", ReadTimeout: clientTimeout, UDPSize: udpBuffSize},
	}

	err := s.listen(address)
	return s, err
}

func (d *DNSServer) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "WeaveDNS (%s)\n", d.ns.ourName)
	fmt.Fprintf(&buf, "  listening on %s, for domain %s\n", d.address, d.domain)
	fmt.Fprintf(&buf, "  response ttl %d\n", d.ttl)
	return buf.String()
}

func (d *DNSServer) listen(address string) error {
	udpListener, err := net.ListenPacket("udp", address)
	if err != nil {
		return err
	}
	udpServer := &dns.Server{PacketConn: udpListener, Handler: d.createMux(d.udpClient, minUDPSize)}

	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		udpServer.Shutdown()
		return err
	}
	tcpServer := &dns.Server{Listener: tcpListener, Handler: d.createMux(d.tcpClient, -1)}

	d.servers = []*dns.Server{udpServer, tcpServer}
	return nil
}

func (d *DNSServer) ActivateAndServe() {
	for _, server := range d.servers {
		go func(server *dns.Server) {
			server.ActivateAndServe()
		}(server)
	}
}

func (d *DNSServer) Stop() error {
	for _, server := range d.servers {
		if err := server.Shutdown(); err != nil {
			return err
		}
	}
	return nil
}

type handler struct {
	*DNSServer
	maxResponseSize int
	client          *dns.Client
}

func (d *DNSServer) createMux(client *dns.Client, defaultMaxResponseSize int) *dns.ServeMux {
	m := dns.NewServeMux()
	h := &handler{
		DNSServer:       d,
		maxResponseSize: defaultMaxResponseSize,
		client:          client,
	}
	m.HandleFunc(d.domain, h.handleLocal)
	m.HandleFunc(reverseDNSdomain, h.handleReverse)
	m.HandleFunc(topDomain, h.handleRecursive)
	return m
}

func (h *handler) handleLocal(w dns.ResponseWriter, req *dns.Msg) {
	h.ns.debugf("local request: %+v", *req)
	if len(req.Question) != 1 {
		h.nameError(w, req)
		return
	}

	hostname := dns.Fqdn(req.Question[0].Name)
	if strings.Count(hostname, ".") == 1 {
		hostname = hostname + h.domain
	}

	addrs := h.ns.Lookup(hostname)
	if len(addrs) == 0 {
		h.nameError(w, req)
		return
	}
	// Per RFC4074, if we have an A but another type was requested,
	// return 'no error' with empty answer section
	if req.Question[0].Qtype != dns.TypeA {
		h.respond(w, h.makeResponse(req, nil))
		return
	}

	header := dns.RR_Header{
		Name:   req.Question[0].Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    h.ttl,
	}
	answers := make([]dns.RR, len(addrs))
	for i, addr := range addrs {
		ip := addr.IP4()
		answers[i] = &dns.A{Hdr: header, A: ip}
	}
	shuffleAnswers(&answers)

	h.respond(w, h.makeResponse(req, answers))
}

func (h *handler) handleReverse(w dns.ResponseWriter, req *dns.Msg) {
	h.ns.debugf("reverse request: %+v", *req)
	if len(req.Question) != 1 || req.Question[0].Qtype != dns.TypePTR {
		h.nameError(w, req)
		return
	}

	ipStr := strings.TrimSuffix(strings.ToLower(req.Question[0].Name), "."+reverseDNSdomain)
	ip, err := address.ParseIP(ipStr)
	if err != nil {
		h.nameError(w, req)
		return
	}

	hostname, err := h.ns.ReverseLookup(ip.Reverse())
	if err != nil {
		h.handleRecursive(w, req)
		return
	}

	header := dns.RR_Header{
		Name:   req.Question[0].Name,
		Rrtype: dns.TypePTR,
		Class:  dns.ClassINET,
		Ttl:    h.ttl,
	}
	answers := []dns.RR{&dns.PTR{
		Hdr: header,
		Ptr: hostname,
	}}

	h.respond(w, h.makeResponse(req, answers))
}

func (h *handler) handleRecursive(w dns.ResponseWriter, req *dns.Msg) {
	h.ns.debugf("recursive request: %+v", *req)

	// Resolve unqualified names locally
	if len(req.Question) == 1 {
		hostname := dns.Fqdn(req.Question[0].Name)
		if strings.Count(hostname, ".") == 1 {
			h.handleLocal(w, req)
			return
		}
	}

	upstreamConfig, err := h.upstream.Config()
	if err != nil {
		h.ns.errorf("unable to read upstream config: %s", err)
	}
	for _, server := range upstreamConfig.Servers {
		reqCopy := req.Copy()
		reqCopy.Id = dns.Id()
		response, _, err := h.client.Exchange(reqCopy, fmt.Sprintf("%s:%s", server, upstreamConfig.Port))
		if (err != nil && err != dns.ErrTruncated) || response == nil {
			h.ns.debugf("error trying %s: %v", server, err)
			continue
		}
		response.Id = req.Id
		if h.responseTooBig(req, response) {
			response.Compress = true
		}
		h.respond(w, response)
		return
	}

	h.respond(w, h.makeErrorResponse(req, dns.RcodeServerFailure))
}

func (h *handler) makeResponse(req *dns.Msg, answers []dns.RR) *dns.Msg {
	response := &dns.Msg{}
	response.SetReply(req)
	response.RecursionAvailable = true
	response.Authoritative = true
	response.Answer = answers
	if !h.responseTooBig(req, response) {
		return response
	}

	// search for smallest i that is too big
	maxSize := h.getMaxResponseSize(req)
	i := sort.Search(len(answers), func(i int) bool {
		// return true if too big
		response.Answer = answers[:i+1]
		return response.Len() > maxSize
	})

	response.Answer = answers[:i]
	if i < len(answers) {
		response.Truncated = true
	}
	return response
}

func (h *handler) makeErrorResponse(req *dns.Msg, code int) *dns.Msg {
	response := &dns.Msg{}
	response.SetReply(req)
	response.RecursionAvailable = true
	response.Rcode = code
	return response
}

func (h *handler) responseTooBig(req, response *dns.Msg) bool {
	return len(response.Answer) > 1 && h.maxResponseSize > 0 && response.Len() > h.getMaxResponseSize(req)
}

func (h *handler) respond(w dns.ResponseWriter, response *dns.Msg) {
	h.ns.debugf("response: %+v", response)
	if err := w.WriteMsg(response); err != nil {
		h.ns.infof("error responding: %v", err)
	}
}

func (h *handler) nameError(w dns.ResponseWriter, req *dns.Msg) {
	h.respond(w, h.makeErrorResponse(req, dns.RcodeNameError))
}

func (h *handler) getMaxResponseSize(req *dns.Msg) int {
	if opt := req.IsEdns0(); opt != nil {
		return int(opt.UDPSize())
	}
	return h.maxResponseSize
}

func shuffleAnswers(answers *[]dns.RR) {
	if len(*answers) <= 1 {
		return
	}

	for i := range *answers {
		j := rand.Intn(i + 1)
		(*answers)[i], (*answers)[j] = (*answers)[j], (*answers)[i]
	}
}
