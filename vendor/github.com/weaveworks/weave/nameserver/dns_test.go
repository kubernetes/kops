package nameserver

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"

	"github.com/weaveworks/mesh"
	"github.com/weaveworks/weave/net/address"
)

type mockUpstream struct {
	config *dns.ClientConfig
}

func (mu *mockUpstream) Config() (*dns.ClientConfig, error) {
	return mu.config, nil
}

func startServer(t *testing.T, upstream *dns.ClientConfig) (*DNSServer, *Nameserver, int, int) {
	peername, err := mesh.PeerNameFromString("00:00:00:02:00:00")
	require.Nil(t, err)
	nameserver := New(peername, "", func(mesh.PeerName) bool { return true })
	dnsserver, err := NewDNSServer(nameserver, "weave.local.", "0.0.0.0:0", &mockUpstream{upstream}, 30, 5*time.Second)
	require.Nil(t, err)
	udpPort := dnsserver.servers[0].PacketConn.LocalAddr().(*net.UDPAddr).Port
	tcpPort := dnsserver.servers[1].Listener.Addr().(*net.TCPAddr).Port
	go dnsserver.ActivateAndServe()
	return dnsserver, nameserver, udpPort, tcpPort
}

func TestTruncation(t *testing.T) {
	//common.SetLogLevel("debug")
	dnsserver, nameserver, udpPort, tcpPort := startServer(t, nil)
	defer dnsserver.Stop()

	// Add 100 mappings to nameserver
	addrs := []address.Address{}
	for i := address.Address(0); i < 100; i++ {
		addrs = append(addrs, i)
		nameserver.AddEntry("foo.weave.local.", "", mesh.UnknownPeerName, i)
	}

	doRequest := func(client *dns.Client, request *dns.Msg, port int, expectedErr error) *dns.Msg {
		request.SetQuestion("foo.weave.local.", dns.TypeA)
		response, _, err := client.Exchange(request, fmt.Sprintf("127.0.0.1:%d", port))
		require.Equal(t, expectedErr, err)
		return response
	}

	// do a udp query, ensure we get a truncated response
	{
		udpClient := dns.Client{Net: "udp", UDPSize: minUDPSize}
		response := doRequest(&udpClient, &dns.Msg{}, udpPort, dns.ErrTruncated)
		require.True(t, response.MsgHdr.Truncated)
		require.True(t, len(response.Answer) < 100)
	}

	// do a udp query with big size, ensure we don't get a truncated response
	{
		udpClient := dns.Client{Net: "udp", UDPSize: 65535}
		request := &dns.Msg{}
		request.SetEdns0(65535, false)
		response := doRequest(&udpClient, request, udpPort, nil)
		require.False(t, response.MsgHdr.Truncated)
		require.Equal(t, len(response.Answer), 100)
	}

	// do a tcp query, ensure we don't get a truncated response
	{
		tcpClient := dns.Client{Net: "tcp"}
		response := doRequest(&tcpClient, &dns.Msg{}, tcpPort, nil)
		require.False(t, response.MsgHdr.Truncated)
		require.Equal(t, len(response.Answer), 100)
	}
}

func TestTruncateResponse(t *testing.T) {

	header := dns.RR_Header{
		Name:   "host.domain.com",
		Rrtype: dns.TypePTR,
		Class:  dns.ClassINET,
		Ttl:    30,
	}

	for i := 0; i < 10000; i++ {
		// generate a random answer set
		numAnswers := 40 + rand.Intn(200)
		answers := make([]dns.RR, numAnswers)
		for j := 0; j < numAnswers; j++ {
			answers[j] = &dns.A{Hdr: header, A: address.Address(j).IP4()}
		}

		// pick a random max size, truncate response to that, check it
		maxSize := 512 + rand.Intn(2*512)
		h := handler{maxResponseSize: maxSize}
		response := h.makeResponse(&dns.Msg{}, answers)
		require.True(t, response.Len() <= maxSize)
	}
}

func TestRecursiveCompress(t *testing.T) {
	const (
		hostname = "foo.example."
		maxSize  = 512
	)

	// Construct a response that is >512 when uncompressed, <512 when compressed
	response := dns.Msg{}
	response.Authoritative = true
	response.Answer = []dns.RR{}
	header := dns.RR_Header{
		Name:   hostname,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    10,
	}
	for response.Len() <= maxSize {
		ip := address.Address(rand.Uint32()).IP4()
		response.Answer = append(response.Answer, &dns.A{Hdr: header, A: ip})
	}
	response.Compress = true
	require.True(t, response.Len() <= maxSize)

	// A dns server that returns the above response
	var gotRequest = make(chan struct{}, 1)
	handleRecursive := func(w dns.ResponseWriter, req *dns.Msg) {
		gotRequest <- struct{}{}
		require.Equal(t, req.Question[0].Name, hostname)
		response.SetReply(req)
		err := w.WriteMsg(&response)
		require.Nil(t, err)
	}
	mux := dns.NewServeMux()
	mux.HandleFunc(topDomain, handleRecursive)
	udpListener, err := net.ListenPacket("udp", "0.0.0.0:0")
	require.Nil(t, err)
	udpServer := &dns.Server{PacketConn: udpListener, Handler: mux}
	udpServerPort := udpListener.LocalAddr().(*net.UDPAddr).Port
	go udpServer.ActivateAndServe()
	defer udpServer.Shutdown()

	// The weavedns server, pointed at the above server
	dnsserver, _, udpPort, _ := startServer(t, &dns.ClientConfig{
		Servers:  []string{"127.0.0.1"},
		Port:     strconv.Itoa(udpServerPort),
		Ndots:    1,
		Timeout:  5,
		Attempts: 2,
	})
	defer dnsserver.Stop()

	// Now do lookup, check its what we expected.
	// NB this doesn't really test golang's resolver behaves correctly, as I can't see
	// a way to point golangs resolver at a specific hosts.
	req := new(dns.Msg)
	req.Id = dns.Id()
	req.RecursionDesired = true
	req.Question = make([]dns.Question, 1)
	req.Question[0] = dns.Question{
		Name:   hostname,
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}
	c := new(dns.Client)
	res, _, err := c.Exchange(req, fmt.Sprintf("127.0.0.1:%d", udpPort))
	require.Nil(t, err)
	require.True(t, len(gotRequest) > 0)
	require.True(t, res.Len() > maxSize)
}
