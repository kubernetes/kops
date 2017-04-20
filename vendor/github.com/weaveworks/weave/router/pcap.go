package router

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/gopacket/pcap"
)

type Pcap struct {
	NonDiscardingFlowOp

	iface *net.Interface
	bufSz int

	// The libpcap handle for writing packets. It's possible that a
	// single handle could be used for reading and writing, but
	// we'd have to examine the performance implications.
	writeHandle *pcap.Handle

	// pcap handles are single-threaded, so we need to lock around
	// uses of writeHandle.
	mutex sync.Mutex

	// The libpcap handle for reading packets
	readHandle *pcap.Handle
}

func NewPcap(iface *net.Interface, bufSz int) (Bridge, error) {
	wh, err := newPcapHandle(iface.Name, false, 0, 0)
	if err != nil {
		return nil, err
	}

	return &Pcap{iface: iface, bufSz: bufSz, writeHandle: wh}, nil
}

func (p *Pcap) StartConsumingPackets(consumer BridgeConsumer) error {
	rh, err := newPcapHandle(p.iface.Name, true, 65535, p.bufSz)
	if err != nil {
		return err
	}

	// Under Linux, libpcap implements the SetDirection filtering
	// in userspace.  So set a BPF filter to discard outbound
	// packets inside the kernel.  We do this here rather than in
	// newPcapIO because libpcap doesn't like this in combination
	// with a 0 snaplen.
	err = rh.SetBPFFilter("inbound")
	if err != nil {
		rh.Close()
		return err
	}

	// readHandle is just for the benefit of Stats.
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.readHandle != nil {
		panic("already consuming")
	}

	p.readHandle = rh
	go p.sniff(rh, consumer)
	return nil
}

func newPcapHandle(ifName string, promisc bool, snaplen int, bufSz int) (handle *pcap.Handle, err error) {
	inactive, err := pcap.NewInactiveHandle(ifName)
	if err != nil {
		return
	}
	defer inactive.CleanUp()
	if err = inactive.SetPromisc(promisc); err != nil {
		return
	}
	if err = inactive.SetSnapLen(snaplen); err != nil {
		return
	}
	if err = inactive.SetTimeout(MaxDuration); err != nil {
		return
	}
	if err = inactive.SetImmediateMode(true); err != nil {
		// If gopacket is compiled against an older pcap.h that
		// doesn't have pcap_set_immediate_mode, it supplies a dummy
		// definition that always returns PCAP_ERROR.  That becomes
		// "Generic error", which is not very helpful.  The real
		// pcap_set_immediate_mode never returns PCAP_ERROR, so this
		// turns it into a more informative message.
		if fmt.Sprint(err) == "Generic error" {
			err = fmt.Errorf("compiled against an old version of libpcap; please compile against libpcap-1.5.0 or later")
		}

		return
	}
	if err = inactive.SetBufferSize(bufSz); err != nil {
		return
	}
	handle, err = inactive.Activate()
	if err != nil {
		return
	}
	err = handle.SetDirection(pcap.DirectionIn)
	return
}

func (p *Pcap) Interface() *net.Interface {
	return p.iface
}

func (p *Pcap) String() string {
	return fmt.Sprint(p.iface.Name, " (via pcap)")
}

func (p *Pcap) InjectPacket(PacketKey) FlowOp {
	return p
}

func (p *Pcap) Process(frame []byte, dec *EthernetDecoder, broadcast bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	checkWarn(p.writeHandle.WritePacketData(frame))
}

func (p *Pcap) sniff(readHandle *pcap.Handle, consumer BridgeConsumer) {
	dec := NewEthernetDecoder()

	for {
		pkt, _, err := readHandle.ZeroCopyReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		}

		checkFatal(err)
		dec.DecodeLayers(pkt)
		if len(dec.decoded) == 0 {
			continue
		}

		if fop := consumer(dec.PacketKey()); !fop.Discards() {
			// We are handing over the frame to
			// forwarders, so we need to make a copy of it
			// in order to prevent the next capture from
			// overwriting the data
			pktLen := len(pkt)
			pktCopy := make([]byte, pktLen, pktLen)
			copy(pktCopy, pkt)

			fop.Process(pktCopy, dec, false)
		}
	}
}

func (p *Pcap) Stats() map[string]int {
	p.mutex.Lock()
	rh := p.readHandle
	p.mutex.Unlock()

	if rh != nil {
		stats, err := rh.Stats()
		if err == nil {
			return map[string]int{
				"PacketsReceived":  stats.PacketsReceived,
				"PacketsDropped":   stats.PacketsDropped,
				"PacketsIfDropped": stats.PacketsIfDropped,
			}
		}
	}

	return nil
}
