package metrics

import (
	"net/http"
	"os"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/weaveworks/weave/common"
)

var (
	blockedConnections = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weavenpc_blocked_connections_total",
			Help: "Connection attempts blocked by policy controller.",
		},
		[]string{"protocol", "dport"},
	)
)

func gatherMetrics() {
	pipe, err := os.Open("/var/log/ulogd.pcap")
	if err != nil {
		common.Log.Fatalf("Failed to open pcap: %v", err)
	}

	reader, err := pcapgo.NewReader(pipe)
	if err != nil {
		common.Log.Fatalf("Failed to read pcap header: %v", err)
	}

	for {
		data, _, err := reader.ReadPacketData()
		if err != nil {
			common.Log.Fatalf("Failed to read pcap packet: %v", err)
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp.SYN && !tcp.ACK { // Only plain SYN constitutes a NEW TCP connection
				blockedConnections.With(prometheus.Labels{"protocol": "tcp", "dport": strconv.Itoa(int(tcp.DstPort))}).Inc()
				common.Log.Warnf("TCP connection from %v:%d to %v:%d blocked by Weave NPC.", srcIP(packet), tcp.SrcPort, dstIP(packet), tcp.DstPort)
				continue
			}
		}

		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			blockedConnections.With(prometheus.Labels{"protocol": "udp", "dport": strconv.Itoa(int(udp.DstPort))}).Inc()
			common.Log.Warnf("UDP connection from %v:%d to %v:%d blocked by Weave NPC.", srcIP(packet), udp.SrcPort, dstIP(packet), udp.DstPort)
			continue
		}
	}
}

const unknownIP string = "<unknown IP>"

func srcIP(packet gopacket.Packet) string {
	if layer := packet.Layer(layers.LayerTypeIPv4); layer != nil {
		if ip, ok := layer.(*layers.IPv4); ok {
			return ip.SrcIP.String()
		}
	}
	if layer := packet.Layer(layers.LayerTypeIPv6); layer != nil {
		if ip, ok := layer.(*layers.IPv6); ok {
			return ip.SrcIP.String()
		}
	}
	return unknownIP
}

func dstIP(packet gopacket.Packet) string {
	if layer := packet.Layer(layers.LayerTypeIPv4); layer != nil {
		if ip, ok := layer.(*layers.IPv4); ok {
			return ip.DstIP.String()
		}
	}
	if layer := packet.Layer(layers.LayerTypeIPv6); layer != nil {
		if ip, ok := layer.(*layers.IPv6); ok {
			return ip.DstIP.String()
		}
	}
	return unknownIP
}

func Start(addr string) error {
	if err := prometheus.Register(blockedConnections); err != nil {
		return err
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		common.Log.Infof("Serving /metrics on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			common.Log.Fatalf("Failed to bind metrics server: %v", err)
		}
	}()

	go gatherMetrics()

	return nil
}
