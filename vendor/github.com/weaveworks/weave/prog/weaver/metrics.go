package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/weaveworks/weave/ipam"
	"github.com/weaveworks/weave/nameserver"
	"github.com/weaveworks/weave/net/address"
	weave "github.com/weaveworks/weave/router"
)

func metricsHandler(router *weave.NetworkRouter, allocator *ipam.Allocator, ns *nameserver.Nameserver, dnsserver *nameserver.DNSServer) http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
	reg.MustRegister(newMetrics(router, allocator, ns, dnsserver))
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

type collector struct {
	router    *weave.NetworkRouter
	allocator *ipam.Allocator
	ns        *nameserver.Nameserver
	dnsserver *nameserver.DNSServer
}

type metric struct {
	*prometheus.Desc
	Collect func(WeaveStatus, *prometheus.Desc, chan<- prometheus.Metric)
}

func desc(fqName, help string, variableLabels ...string) *prometheus.Desc {
	return prometheus.NewDesc(fqName, help, variableLabels, prometheus.Labels{})
}

func intGauge(desc *prometheus.Desc, val int, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(val), labels...)
}
func uint64Counter(desc *prometheus.Desc, val uint64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(val), labels...)
}

var metrics = []metric{
	{desc("weave_connections", "Number of peer-to-peer connections.", "state"),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			counts := make(map[string]int)
			for _, conn := range s.Router.Connections {
				counts[conn.State]++
			}
			for _, state := range allConnectionStates {
				ch <- intGauge(desc, counts[state], state)
			}
		}},
	{desc("weave_connection_terminations_total", "Number of peer-to-peer connections terminated."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			ch <- uint64Counter(desc, uint64(s.Router.TerminationCount))
		}},
	{desc("weave_ips", "Number of IP addresses.", "state"),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if s.IPAM != nil {
				ch <- intGauge(desc, s.IPAM.ActiveIPs, "local-used")
			}
		}},
	{desc("weave_max_ips", "Size of IP address space used by allocator."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if s.IPAM != nil {
				ch <- intGauge(desc, s.IPAM.RangeNumIPs)
			}
		}},
	{desc("weave_dns_entries", "Number of DNS entries."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if s.DNS != nil {
				ch <- intGauge(desc, countDNSEntriesForPeer(s.Router.Name, s.DNS.Entries))
			}
		}},
	{desc("weave_flows", "Number of FastDP flows."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if metrics := fastDPMetrics(s); metrics != nil {
				ch <- intGauge(desc, metrics.Flows)
			}
		}},
	{desc("weave_ipam_pending_allocates", "Number of pending allocates."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if s.IPAM != nil {
				ch <- intGauge(desc, len(s.IPAM.PendingAllocates))
			}
		}},
	{desc("weave_ipam_pending_claims", "Number of pending claims."),
		func(s WeaveStatus, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
			if s.IPAM != nil {
				ch <- intGauge(desc, len(s.IPAM.PendingClaims))
			}
		}},
}

func fastDPMetrics(s WeaveStatus) *weave.FastDPMetrics {
	if diagMap, ok := s.Router.OverlayDiagnostics.(map[string]interface{}); ok {
		if diag, ok := diagMap["fastdp"]; ok {
			if fastDPStats, ok := diag.(weave.FastDPStatus); ok {
				return fastDPStats.Metrics().(*weave.FastDPMetrics)
			}
		}
	}
	return nil
}

func newMetrics(router *weave.NetworkRouter, allocator *ipam.Allocator, ns *nameserver.Nameserver, dnsserver *nameserver.DNSServer) *collector {
	return &collector{
		router:    router,
		allocator: allocator,
		ns:        ns,
		dnsserver: dnsserver,
	}
}

func (m *collector) Collect(ch chan<- prometheus.Metric) {

	status := WeaveStatus{"", nil,
		weave.NewNetworkRouterStatus(m.router),
		ipam.NewStatus(m.allocator, address.CIDR{}),
		nameserver.NewStatus(m.ns, m.dnsserver)}

	for _, metric := range metrics {
		metric.Collect(status, metric.Desc, ch)
	}
}

func (m *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metrics {
		ch <- metric.Desc
	}
}
