// This contains the Overlay implementation for weave's own UDP
// encapsulation protocol ("sleeve" because a sleeve encapsulates
// something, it's often woven, it rhymes with "weave", make up your
// own cheesy reason).

package router

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/weaveworks/mesh"
)

// This diagram explains the various arithmetic and variables related
// to packet offsets and lengths below:
//
// +----+-----+--------+--------+----------+--------------------------+
// | IP | UDP | Sleeve | Sleeve | Overlay  | Overlay Layer 3 Payload  |
// |    |     | Packet | Frame  | Ethernet |                          |
// |    |     | Header | Header |          |                          |
// +----+-----+--------+--------+----------+--------------------------+
//
// <------------------------------------ msgTooBigError.underlayPMTU ->
//
//            <-------------------------- sleeveForwarder.maxPayload ->
//
// <---------->                                             UDPOverhead
//
//            <-------->                       Encryptor.PacketOverhead
//
//                     <-------->               Encryptor.FrameOverhead
//
//                              <---------->           EthernetOverhead
//
// <---------------------------------------> sleeveForwarder.overheadDF
//
// sleeveForwarder.mtu                     <-------------------------->

const (
	EthernetOverhead  = 14
	UDPOverhead       = 28 // 20 bytes for IPv4, 8 bytes for UDP
	DefaultMTU        = 65535
	FragTestSize      = 60001
	PMTUDiscoverySize = 60000
	FragTestInterval  = 5 * time.Minute
	MTUVerifyAttempts = 8
	MTUVerifyTimeout  = 10 * time.Millisecond // doubled with each attempt

	ProtocolConnectionEstablished = mesh.ProtocolReserved1
	ProtocolFragmentationReceived = mesh.ProtocolReserved2
	ProtocolPMTUVerified          = mesh.ProtocolReserved3
)

type SleeveOverlay struct {
	host      string
	localPort int

	// These fields are set in StartConsumingPackets, and not
	// subsequently modified
	localPeer    *mesh.Peer
	localPeerBin []byte
	consumer     OverlayConsumer
	peers        *mesh.Peers
	conn         *net.UDPConn

	lock       sync.Mutex
	forwarders map[mesh.PeerName]*sleeveForwarder
}

func NewSleeveOverlay(host string, localPort int) NetworkOverlay {
	return &SleeveOverlay{host: host, localPort: localPort}
}

func (sleeve *SleeveOverlay) StartConsumingPackets(localPeer *mesh.Peer, peers *mesh.Peers, consumer OverlayConsumer) error {
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprint(sleeve.host, ":", sleeve.localPort))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return err
	}

	f, err := conn.File()
	if err != nil {
		return err
	}

	defer f.Close()
	fd := int(f.Fd())

	// This makes sure all packets we send out do not have DF set
	// on them.
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_MTU_DISCOVER, syscall.IP_PMTUDISC_DONT)
	if err != nil {
		return err
	}

	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()

	if sleeve.localPeer != nil {
		conn.Close()
		return fmt.Errorf("StartConsumingPackets already called")
	}

	sleeve.localPeer = localPeer
	sleeve.localPeerBin = localPeer.NameByte
	sleeve.consumer = consumer
	sleeve.peers = peers
	sleeve.conn = conn
	sleeve.forwarders = make(map[mesh.PeerName]*sleeveForwarder)
	go sleeve.readUDP()
	return nil
}

func (*SleeveOverlay) InvalidateRoutes() {
	// no cached information, so nothing to do
}

func (*SleeveOverlay) InvalidateShortIDs() {
	// no cached information, so nothing to do
}

func (*SleeveOverlay) AddFeaturesTo(map[string]string) {
	// No features to be provided, to facilitate compatibility
}

func (*SleeveOverlay) Diagnostics() interface{} {
	return nil
}

func (*SleeveOverlay) Stop() {
	// do nothing
}

func (sleeve *SleeveOverlay) lookupForwarder(peer mesh.PeerName) *sleeveForwarder {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	return sleeve.forwarders[peer]
}

func (sleeve *SleeveOverlay) addForwarder(peer mesh.PeerName, fwd *sleeveForwarder) {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	sleeve.forwarders[peer] = fwd
}

func (sleeve *SleeveOverlay) removeForwarder(peer mesh.PeerName, fwd *sleeveForwarder) {
	sleeve.lock.Lock()
	defer sleeve.lock.Unlock()
	if sleeve.forwarders[peer] == fwd {
		delete(sleeve.forwarders, peer)
	}
}

func (sleeve *SleeveOverlay) readUDP() {
	defer sleeve.conn.Close()
	dec := NewEthernetDecoder()
	buf := make([]byte, MaxUDPPacketSize)

	for {
		n, sender, err := sleeve.conn.ReadFromUDP(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Print("ignoring UDP read error ", err)
			continue
		} else if n < NameSize {
			log.Print("ignoring too short UDP packet from ", sender)
			continue
		}

		fwdName := mesh.PeerNameFromBin(buf[:NameSize])
		fwd := sleeve.lookupForwarder(fwdName)
		if fwd == nil {
			continue
		}

		packet := make([]byte, n-NameSize)
		copy(packet, buf[NameSize:n])

		err = fwd.crypto.Dec.IterateFrames(packet,
			func(src []byte, dst []byte, frame []byte) {
				sleeve.handleFrame(sender, fwd, src, dst, frame, dec)
			})
		if err != nil {
			// Errors during UDP packet decoding /
			// processing are non-fatal. One common cause
			// is that we receive and attempt to decrypt a
			// "stray" packet. This can actually happen
			// quite easily if there is some connection
			// churn between two peers. After all, UDP
			// isn't a connection-oriented protocol, yet
			// we pretend it is.
			//
			// If anything really is seriously,
			// unrecoverably amiss with a connection, that
			// will typically result in missed heartbeats
			// and the connection getting shut down
			// because of that.
			log.Print(fwd.logPrefixFor(sender), err)
		}
	}
}

func (sleeve *SleeveOverlay) handleFrame(sender *net.UDPAddr, fwd *sleeveForwarder, src []byte, dst []byte, frame []byte, dec *EthernetDecoder) {
	dec.DecodeLayers(frame)
	decodedLen := len(dec.decoded)
	if decodedLen == 0 {
		return
	}

	srcPeer := sleeve.peers.Fetch(mesh.PeerNameFromBin(src))
	dstPeer := sleeve.peers.Fetch(mesh.PeerNameFromBin(dst))
	if srcPeer == nil || dstPeer == nil {
		return
	}

	// Handle special frames produced internally (rather than
	// captured/forwarded) by the remote router.
	//
	// We really shouldn't be decoding these above, since they are
	// not genuine Ethernet frames. However, it is actually more
	// efficient to do so, as we want to optimise for the common
	// (i.e. non-special) frames. These always need decoding, and
	// detecting special frames is cheaper post decoding than pre.
	if decodedLen == 1 && dec.IsSpecial() {
		if srcPeer == fwd.remotePeer && dstPeer == fwd.sleeve.localPeer {
			select {
			case fwd.specialChan <- specialFrame{sender, frame}:
			case <-fwd.finishedChan:
			}
		}

		return
	}

	sleeve.sendToConsumer(srcPeer, dstPeer, frame, dec)
}

func (sleeve *SleeveOverlay) sendToConsumer(srcPeer, dstPeer *mesh.Peer, frame []byte, dec *EthernetDecoder) {
	if sleeve.consumer == nil {
		return
	}

	fop := sleeve.consumer(ForwardPacketKey{
		SrcPeer:   srcPeer,
		DstPeer:   dstPeer,
		PacketKey: dec.PacketKey(),
	})
	if fop != nil {
		fop.Process(frame, dec, false)
	}
}

type udpSender interface {
	send([]byte, *net.UDPAddr) error
}

func (sleeve *SleeveOverlay) send(msg []byte, raddr *net.UDPAddr) error {
	sleeve.lock.Lock()
	conn := sleeve.conn
	sleeve.lock.Unlock()

	if conn == nil {
		// Consume wasn't called yet
		return nil
	}

	_, err := conn.WriteToUDP(msg, raddr)
	return err
}

type sleeveCrypto struct {
	Dec   Decryptor
	Enc   Encryptor
	EncDF Encryptor
}

func newSleeveCrypto(name []byte, sessionKey *[32]byte, outbound bool) sleeveCrypto {
	if sessionKey == nil {
		return sleeveCrypto{
			Dec:   NewNonDecryptor(),
			Enc:   NewNonEncryptor(name),
			EncDF: NewNonEncryptor(name),
		}
	}
	return sleeveCrypto{
		Dec:   NewNaClDecryptor(sessionKey, outbound),
		Enc:   NewNaClEncryptor(name, sessionKey, outbound, false),
		EncDF: NewNaClEncryptor(name, sessionKey, outbound, true),
	}
}

func (crypto sleeveCrypto) Overhead() int {
	return UDPOverhead + crypto.EncDF.PacketOverhead() + crypto.EncDF.FrameOverhead() + EthernetOverhead
}

type sleeveForwarder struct {
	// Immutable
	sleeve         *SleeveOverlay
	remotePeer     *mesh.Peer
	remotePeerBin  []byte
	sendControlMsg func(byte, []byte) error
	connUID        uint64

	// Channels to communicate with the aggregator goroutine
	aggregatorChan   chan<- aggregatorFrame
	aggregatorDFChan chan<- aggregatorFrame
	specialChan      chan<- specialFrame
	controlMsgChan   chan<- controlMessage
	confirmedChan    chan<- struct{}
	finishedChan     <-chan struct{}

	// listener channels
	establishedChan chan struct{}
	errorChan       chan error

	// Explicitly locked state
	lock       sync.RWMutex
	remoteAddr *net.UDPAddr

	// These fields are accessed and updated independently, so no
	// locking needed.
	mtu       int // the mtu for this link on the overlay network
	stackFrag bool

	// State only used within the forwarder goroutine
	crypto     sleeveCrypto
	senderDF   *udpSenderDF
	maxPayload int

	// How many bytes of overhead it takes to turn an IP packet on
	// the overlay network into an encapsulated packet on the underlay
	// network
	overheadDF int

	heartbeatInterval time.Duration
	heartbeatTimer    *time.Timer
	heartbeatTimeout  *time.Timer
	fragTestTicker    *time.Ticker
	ackedHeartbeat    bool

	mtuTestTimeout *time.Timer
	mtuTestsSent   uint
	mtuHighestGood int
	mtuLowestBad   int
	mtuCandidate   int
}

type aggregatorFrame struct {
	src   []byte
	dst   []byte
	frame []byte
}

// A "special" frame over UDP
type specialFrame struct {
	sender *net.UDPAddr
	frame  []byte
}

// A control message
type controlMessage struct {
	tag byte
	msg []byte
}

func (sleeve *SleeveOverlay) PrepareConnection(params mesh.OverlayConnectionParams) (mesh.OverlayConnection, error) {
	aggChan := make(chan aggregatorFrame, ChannelSize)
	aggDFChan := make(chan aggregatorFrame, ChannelSize)
	specialChan := make(chan specialFrame, 1)
	controlMsgChan := make(chan controlMessage, 1)
	confirmedChan := make(chan struct{})
	finishedChan := make(chan struct{})

	var remoteAddr *net.UDPAddr
	if params.Outbound {
		remoteAddr = makeUDPAddr(params.RemoteAddr)
	}

	crypto := newSleeveCrypto(sleeve.localPeer.NameByte, params.SessionKey, params.Outbound)

	fwd := &sleeveForwarder{
		sleeve:           sleeve,
		remotePeer:       params.RemotePeer,
		remotePeerBin:    params.RemotePeer.NameByte,
		sendControlMsg:   params.SendControlMessage,
		connUID:          params.ConnUID,
		aggregatorChan:   aggChan,
		aggregatorDFChan: aggDFChan,
		specialChan:      specialChan,
		controlMsgChan:   controlMsgChan,
		confirmedChan:    confirmedChan,
		finishedChan:     finishedChan,
		establishedChan:  make(chan struct{}),
		errorChan:        make(chan error, 1),
		remoteAddr:       remoteAddr,
		mtu:              DefaultMTU,
		crypto:           crypto,
		maxPayload:       DefaultMTU - UDPOverhead,
		overheadDF:       crypto.Overhead(),
		senderDF:         newUDPSenderDF(params.LocalAddr.IP, sleeve.localPort),
	}

	go fwd.run(aggChan, aggDFChan, specialChan, controlMsgChan, confirmedChan, finishedChan)
	return fwd, nil
}

func (fwd *sleeveForwarder) logPrefixFor(sender *net.UDPAddr) string {
	return fmt.Sprintf("sleeve ->[%s|%s]: ", sender, fwd.remotePeer)
}

func (fwd *sleeveForwarder) logPrefix() string {
	fwd.lock.RLock()
	remoteAddr := fwd.remoteAddr
	fwd.lock.RUnlock()
	return fwd.logPrefixFor(remoteAddr)
}

func (fwd *sleeveForwarder) Confirm() {
	log.Debug(fwd.logPrefix(), "Confirm")
	select {
	case fwd.confirmedChan <- struct{}{}:
	case <-fwd.finishedChan:
	}
}

func (fwd *sleeveForwarder) EstablishedChannel() <-chan struct{} {
	return fwd.establishedChan
}

func (fwd *sleeveForwarder) ErrorChannel() <-chan error {
	return fwd.errorChan
}

type curriedForward struct {
	NonDiscardingFlowOp
	fwd *sleeveForwarder
	key ForwardPacketKey
}

func (fwd *sleeveForwarder) Forward(key ForwardPacketKey) FlowOp {
	return curriedForward{fwd: fwd, key: key}
}

func (f curriedForward) Process(frame []byte, dec *EthernetDecoder, broadcast bool) {
	fwd := f.fwd
	fwd.lock.RLock()
	haveContact := (fwd.remoteAddr != nil)
	mtu := fwd.mtu
	stackFrag := fwd.stackFrag
	fwd.lock.RUnlock()

	if !haveContact {
		log.Print(fwd.logPrefix(), "Cannot forward frame yet - awaiting contact")
		return
	}

	srcName := f.key.SrcPeer.NameByte
	dstName := f.key.DstPeer.NameByte

	// We could use non-blocking channel sends here, i.e. drop frames
	// on the floor when the forwarder is busy. This would allow our
	// caller - the capturing loop in the router - to read frames more
	// quickly when under load, i.e. we'd drop fewer frames on the
	// floor during capture. And we could maximise CPU utilisation
	// since we aren't stalling a thread. However, a lot of work has
	// already been done by the time we get here. Since any packet we
	// drop will likely get re-transmitted we end up paying that cost
	// multiple times. So it's better to drop things at the beginning
	// of our pipeline.
	if dec.DF() {
		if !frameTooBig(frame, mtu) {
			fwd.aggregate(fwd.aggregatorDFChan, srcName, dstName, frame)
			return
		}

		// Why do we need an explicit broadcast hint here,
		// rather than just checking the frame for a broadcast
		// destination MAC address?  Because even
		// non-broadcast frames can be broadcast, if the
		// destination MAC was not in our MAC cache.
		if broadcast {
			log.Print(fwd.logPrefix(), "dropping too big DF broadcast frame (", dec.IP.SrcIP, " -> ", dec.IP.DstIP, "): MTU=", mtu)
			return
		}

		// Send an ICMP back to where the frame came from
		fragNeededPacket, err := dec.makeICMPFragNeeded(mtu)
		if err != nil {
			log.Print(fwd.logPrefix(), err)
			return
		}

		dec.DecodeLayers(fragNeededPacket)

		// The frag-needed packet does not have DF set, so the
		// potential recursion here is bounded.
		fwd.sleeve.sendToConsumer(f.key.DstPeer, f.key.SrcPeer, fragNeededPacket, dec)
		return
	}

	if stackFrag || len(dec.decoded) < 2 {
		fwd.aggregate(fwd.aggregatorChan, srcName, dstName, frame)
		return
	}

	// Don't have trustworthy stack, so we're going to have to
	// send it DF in any case.
	if !frameTooBig(frame, mtu) {
		fwd.aggregate(fwd.aggregatorDFChan, srcName, dstName, frame)
		return
	}

	// We can't trust the stack to fragment, we have IP, and we
	// have a frame that's too big for the MTU, so we have to
	// fragment it ourself.
	checkWarn(fragment(dec.Eth, dec.IP, mtu,
		func(segFrame []byte) {
			fwd.aggregate(fwd.aggregatorDFChan, srcName, dstName, segFrame)
		}))
}

func (fwd *sleeveForwarder) aggregate(ch chan<- aggregatorFrame, src []byte, dst []byte, frame []byte) {
	select {
	case ch <- aggregatorFrame{src, dst, frame}:
	case <-fwd.finishedChan:
	}
}

func fragment(eth layers.Ethernet, ip layers.IPv4, mtu int, forward func([]byte)) error {
	// We are not doing any sort of NAT, so we don't need to worry
	// about checksums of IP payload (eg UDP checksum).
	headerSize := int(ip.IHL) * 4
	// &^ is bit clear (AND NOT). So here we're clearing the lowest 3
	// bits.
	maxSegmentSize := (mtu - headerSize) &^ 7
	opts := gopacket.SerializeOptions{
		FixLengths:       false,
		ComputeChecksums: true}
	payloadSize := int(ip.Length) - headerSize
	payload := ip.BaseLayer.Payload[:payloadSize]
	offsetBase := int(ip.FragOffset) << 3
	origFlags := ip.Flags
	ip.Flags = ip.Flags | layers.IPv4MoreFragments
	ip.Length = uint16(headerSize + maxSegmentSize)
	if eth.EthernetType == layers.EthernetTypeLLC {
		// using LLC, so must set eth length correctly. eth length
		// is just the length of the payload
		eth.Length = ip.Length
	} else {
		eth.Length = 0
	}
	for offset := 0; offset < payloadSize; offset += maxSegmentSize {
		var segmentPayload []byte
		if len(payload) <= maxSegmentSize {
			// last one
			segmentPayload = payload
			ip.Length = uint16(len(payload) + headerSize)
			ip.Flags = origFlags
			if eth.EthernetType == layers.EthernetTypeLLC {
				eth.Length = ip.Length
			} else {
				eth.Length = 0
			}
		} else {
			segmentPayload = payload[:maxSegmentSize]
			payload = payload[maxSegmentSize:]
		}
		ip.FragOffset = uint16((offset + offsetBase) >> 3)
		buf := gopacket.NewSerializeBuffer()
		segPayload := gopacket.Payload(segmentPayload)
		err := gopacket.SerializeLayers(buf, opts, &eth, &ip, &segPayload)
		if err != nil {
			return err
		}

		forward(buf.Bytes())
	}
	return nil
}

func frameTooBig(frame []byte, mtu int) bool {
	// We capture/forward complete ethernet frames. Therefore the
	// frame length includes the ethernet header. However, MTUs
	// operate at the IP layer and thus do not include the ethernet
	// header. To put it another way, when a sender that was told an
	// MTU of M sends an IP packet of exactly that length, we will
	// capture/forward M + EthernetOverhead bytes of data.
	return len(frame) > mtu+EthernetOverhead
}

func (fwd *sleeveForwarder) ControlMessage(tag byte, msg []byte) {
	select {
	case fwd.controlMsgChan <- controlMessage{tag, msg}:
	case <-fwd.finishedChan:
	}
}

func (fwd *sleeveForwarder) Attrs() map[string]interface{} {
	return map[string]interface{}{"name": "sleeve", "mtu": fwd.mtu}
}

func (fwd *sleeveForwarder) Stop() {
	fwd.sleeve.removeForwarder(fwd.remotePeer.Name, fwd)

	// Tell the forwarder goroutine to finish.  We don't need to
	// wait for it.
	close(fwd.confirmedChan)
}

func (fwd *sleeveForwarder) run(aggChan <-chan aggregatorFrame,
	aggDFChan <-chan aggregatorFrame,
	specialChan <-chan specialFrame,
	controlMsgChan <-chan controlMessage,
	confirmedChan <-chan struct{},
	finishedChan chan<- struct{}) {
	defer close(finishedChan)

	var err error
loop:
	for err == nil {
		select {
		case frame := <-aggChan:
			err = fwd.aggregateAndSend(frame, aggChan, fwd.crypto.Enc, fwd.sleeve, MaxUDPPacketSize-UDPOverhead)

		case frame := <-aggDFChan:
			err = fwd.aggregateAndSend(frame, aggDFChan, fwd.crypto.EncDF, fwd.senderDF, fwd.maxPayload)

		case sf := <-specialChan:
			err = fwd.handleSpecialFrame(sf)

		case cm := <-controlMsgChan:
			err = fwd.handleControlMessage(cm)

		case _, ok := <-confirmedChan:
			if !ok {
				// confirmedChan is closed to indicate
				// the forwarder is being closed
				break loop
			}

			err = fwd.confirmed()

		case <-timerChan(fwd.heartbeatTimer):
			err = fwd.sendHeartbeat()

		case <-timerChan(fwd.heartbeatTimeout):
			err = fmt.Errorf("timed out waiting for UDP heartbeat")

		case <-tickerChan(fwd.fragTestTicker):
			err = fwd.sendFragTest()

		case <-timerChan(fwd.mtuTestTimeout):
			err = fwd.handleMTUTestFailure()
		}
	}

	if fwd.heartbeatTimer != nil {
		fwd.heartbeatTimer.Stop()
	}
	if fwd.heartbeatTimeout != nil {
		fwd.heartbeatTimeout.Stop()
	}
	if fwd.fragTestTicker != nil {
		fwd.fragTestTicker.Stop()
	}
	if fwd.mtuTestTimeout != nil {
		fwd.mtuTestTimeout.Stop()
	}

	checkWarn(fwd.senderDF.close())

	fwd.lock.RLock()
	defer fwd.lock.RUnlock()

	// this is the only place we send an error to errorChan
	fwd.errorChan <- err
}

func (fwd *sleeveForwarder) aggregateAndSend(frame aggregatorFrame, aggChan <-chan aggregatorFrame, enc Encryptor, sender udpSender, limit int) error {
	// Give up after processing N frames, to avoid starving the
	// other activities of the forwarder goroutine.
	i := 0

	for {
		// Adding the first frame to an empty buffer
		if !fits(frame, enc, limit) {
			log.Print(fwd.logPrefix(), "Dropping too big frame during forwarding: frame len ", len(frame.frame), ", limit ", limit)
			return nil
		}

		for {
			enc.AppendFrame(frame.src, frame.dst, frame.frame)
			i++

			gotOne := false
			if i < 100 {
				select {
				case frame = <-aggChan:
					gotOne = true
				default:
				}
			}

			if !gotOne {
				return fwd.flushEncryptor(enc, sender)
			}

			// Accumulate frames until doing so would
			// exceed the MTU.  Even in the non-DF case,
			// it doesn't seem worth adding a frame where
			// that would lead to fragmentation,
			// potentially delaying or risking other
			// frames.
			if !fits(frame, enc, fwd.maxPayload) {
				break
			}
		}

		if err := fwd.flushEncryptor(enc, sender); err != nil {
			return err
		}
	}
}

func fits(frame aggregatorFrame, enc Encryptor, limit int) bool {
	return enc.TotalLen()+enc.FrameOverhead()+len(frame.frame) <= limit
}

func (fwd *sleeveForwarder) flushEncryptor(enc Encryptor, sender udpSender) error {
	msg, err := enc.Bytes()
	if err != nil {
		return err
	}

	return fwd.processSendError(sender.send(msg, fwd.remoteAddr))
}

func (fwd *sleeveForwarder) sendSpecial(enc Encryptor, sender udpSender, data []byte) error {
	enc.AppendFrame(fwd.sleeve.localPeerBin, fwd.remotePeerBin, data)
	return fwd.flushEncryptor(enc, sender)
}

func (fwd *sleeveForwarder) handleSpecialFrame(special specialFrame) error {
	// The special frame types are distinguished by length
	switch len(special.frame) {
	case EthernetOverhead + 8:
		return fwd.handleHeartbeat(special)

	case FragTestSize:
		return fwd.handleFragTest(special.frame)

	default:
		return fwd.handleMTUTest(special.frame)
	}
}

func (fwd *sleeveForwarder) handleControlMessage(cm controlMessage) error {
	switch cm.tag {
	case ProtocolConnectionEstablished:
		return fwd.handleHeartbeatAck()

	case ProtocolFragmentationReceived:
		return fwd.handleFragTestAck()

	case ProtocolPMTUVerified:
		return fwd.handleMTUTestAck(cm.msg)

	default:
		log.Print(fwd.logPrefix(), "Ignoring unknown control message tag: ", cm.tag)
		return nil
	}
}

func (fwd *sleeveForwarder) confirmed() error {
	log.Debug(fwd.logPrefix(), "confirmed")

	if fwd.heartbeatInterval != 0 {
		// already confirmed
		return nil
	}

	// when the connection is confirmed, this should be the only
	// forwarder to the peer.
	fwd.sleeve.addForwarder(fwd.remotePeer.Name, fwd)

	// heartbeatInterval flags that we want to send heartbeats,
	// even if we don't do sendHeartbeat() yet due to lacking the
	// remote address.
	fwd.heartbeatInterval = FastHeartbeat
	if fwd.remoteAddr != nil {
		if err := fwd.sendHeartbeat(); err != nil {
			return err
		}
	}

	fwd.heartbeatTimeout = time.NewTimer(HeartbeatTimeout)
	return nil
}

func (fwd *sleeveForwarder) sendHeartbeat() error {
	log.Debug(fwd.logPrefix(), "sendHeartbeat")

	// Prime the timer for the next heartbeat.  We don't use a
	// ticker because the interval is not constant.
	fwd.heartbeatTimer = setTimer(fwd.heartbeatTimer, fwd.heartbeatInterval)

	buf := make([]byte, EthernetOverhead+8)
	binary.BigEndian.PutUint64(buf[EthernetOverhead:], fwd.connUID)
	return fwd.sendSpecial(fwd.crypto.EncDF, fwd.senderDF, buf)
}

func (fwd *sleeveForwarder) handleHeartbeat(special specialFrame) error {
	uid := binary.BigEndian.Uint64(special.frame[EthernetOverhead:])
	if uid != fwd.connUID {
		return nil
	}

	log.Debug(fwd.logPrefix(), "handleHeartbeat")

	if fwd.remoteAddr == nil {
		fwd.setRemoteAddr(special.sender)
		if fwd.heartbeatInterval != 0 {
			if err := fwd.sendHeartbeat(); err != nil {
				return err
			}
		}
	} else if !udpAddrsEqual(fwd.remoteAddr, special.sender) {
		log.Print(fwd.logPrefix(), "Peer UDP address changed to ", special.sender)
		fwd.setRemoteAddr(special.sender)
	}

	if !fwd.ackedHeartbeat {
		fwd.ackedHeartbeat = true
		if err := fwd.sendControlMsg(ProtocolConnectionEstablished, nil); err != nil {
			return err
		}
	}

	// we can receive a heartbeat before confirmed() has set up
	// heartbeatTimeout
	if fwd.heartbeatTimeout != nil {
		fwd.heartbeatTimeout.Reset(HeartbeatTimeout)
	}

	return nil
}

func (fwd *sleeveForwarder) setRemoteAddr(addr *net.UDPAddr) {
	// remoteAddr is only modified here, so we don't need to hold
	// the lock when reading it from the forwarder goroutine.  But
	// other threads may read it while holding the read lock, so
	// when we modify it, we need to hold the write lock.
	fwd.lock.Lock()
	fwd.remoteAddr = addr
	fwd.lock.Unlock()
}

func (fwd *sleeveForwarder) handleHeartbeatAck() error {
	log.Debug(fwd.logPrefix(), "handleHeartbeatAck")

	if fwd.heartbeatInterval != SlowHeartbeat {
		fwd.heartbeatInterval = SlowHeartbeat
		if fwd.heartbeatTimer != nil {
			fwd.heartbeatTimer.Reset(fwd.heartbeatInterval)
		}

		// The connection is now regarded as established
		close(fwd.establishedChan)
	}

	fwd.fragTestTicker = time.NewTicker(FragTestInterval)
	if err := fwd.sendFragTest(); err != nil {
		return err
	}

	// Send a large frame down the DF channel.  An EMSGSIZE will
	// result, which is handled in processSendError, prompting
	// PMTU discovery to start.
	return fwd.sendSpecial(fwd.crypto.EncDF, fwd.senderDF, make([]byte, PMTUDiscoverySize))
}

func (fwd *sleeveForwarder) sendFragTest() error {
	log.Debug(fwd.logPrefix(), "sendFragTest")
	fwd.stackFrag = false
	return fwd.sendSpecial(fwd.crypto.Enc, fwd.sleeve, make([]byte, FragTestSize))
}

func (fwd *sleeveForwarder) handleFragTest(frame []byte) error {
	if !allZeros(frame) {
		return nil
	}

	return fwd.sendControlMsg(ProtocolFragmentationReceived, nil)
}

func (fwd *sleeveForwarder) handleFragTestAck() error {
	log.Debug(fwd.logPrefix(), "handleFragTestAck")
	fwd.stackFrag = true
	return nil
}

func (fwd *sleeveForwarder) processSendError(err error) error {
	if mtbe, ok := err.(msgTooBigError); ok {
		mtu := mtbe.underlayPMTU - fwd.overheadDF
		if fwd.mtuCandidate != 0 && mtu >= fwd.mtuCandidate {
			return nil
		}

		fwd.mtuHighestGood = 552
		fwd.mtuLowestBad = mtu + 1
		fwd.mtuCandidate = mtu
		fwd.mtuTestsSent = 0
		fwd.maxPayload = mtbe.underlayPMTU - UDPOverhead
		fwd.mtu = mtu
		return fwd.sendMTUTest()
	}

	return err
}

func (fwd *sleeveForwarder) sendMTUTest() error {
	log.Debug(fwd.logPrefix(), "sendMTUTest: mtu candidate ", fwd.mtuCandidate)

	err := fwd.sendSpecial(fwd.crypto.EncDF, fwd.senderDF, make([]byte, fwd.mtuCandidate+EthernetOverhead))
	if err != nil {
		return err
	}

	fwd.mtuTestTimeout = setTimer(fwd.mtuTestTimeout, MTUVerifyTimeout<<fwd.mtuTestsSent)
	fwd.mtuTestsSent++
	return nil
}

func (fwd *sleeveForwarder) handleMTUTest(frame []byte) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(len(frame)-EthernetOverhead))
	return fwd.sendControlMsg(ProtocolPMTUVerified, buf)
}

func (fwd *sleeveForwarder) handleMTUTestAck(msg []byte) error {
	if len(msg) < 2 {
		log.Print(fwd.logPrefix(), "Received truncated MTUTestAck")
		return nil
	}

	mtu := int(binary.BigEndian.Uint16(msg))
	log.Debug(fwd.logPrefix(), "handleMTUTestAck: for mtu candidate ", mtu)
	if mtu != fwd.mtuCandidate {
		return nil
	}

	fwd.mtuHighestGood = mtu
	return fwd.searchMTU()
}

func (fwd *sleeveForwarder) handleMTUTestFailure() error {
	if fwd.mtuTestsSent < MTUVerifyAttempts {
		return fwd.sendMTUTest()
	}

	log.Debug(fwd.logPrefix(), "handleMTUTestFailure")
	fwd.mtuLowestBad = fwd.mtuCandidate
	return fwd.searchMTU()
}

func (fwd *sleeveForwarder) searchMTU() error {
	log.Debug(fwd.logPrefix(), "searchMTU: ", fwd.mtuHighestGood, fwd.mtuLowestBad)

	if fwd.mtuHighestGood+1 >= fwd.mtuLowestBad {
		mtu := fwd.mtuHighestGood
		log.Print(fwd.logPrefix(), "Effective MTU verified at ", mtu)

		if fwd.mtuTestTimeout != nil {
			fwd.mtuTestTimeout.Stop()
			fwd.mtuTestTimeout = nil
		}

		fwd.mtuCandidate = 0
		fwd.maxPayload = mtu + fwd.overheadDF - UDPOverhead
		fwd.mtu = mtu
		return nil
	}

	fwd.mtuCandidate = (fwd.mtuHighestGood + fwd.mtuLowestBad) / 2
	fwd.mtuTestsSent = 0
	return fwd.sendMTUTest()
}

type udpSenderDF struct {
	ipBuf     gopacket.SerializeBuffer
	opts      gopacket.SerializeOptions
	udpHeader *layers.UDP
	localIP   net.IP
	remoteIP  net.IP
	socket    *net.IPConn
}

func newUDPSenderDF(localIP net.IP, localPort int) *udpSenderDF {
	return &udpSenderDF{
		ipBuf: gopacket.NewSerializeBuffer(),
		opts: gopacket.SerializeOptions{
			FixLengths: true,
			// UDP header is calculated with a phantom IP
			// header. Yes, it's totally nuts. Thankfully,
			// for UDP over IPv4, the checksum is
			// optional. It's not optional for IPv6, but
			// we'll ignore that for now. TODO
			ComputeChecksums: false,
		},
		udpHeader: &layers.UDP{SrcPort: layers.UDPPort(localPort)},
		localIP:   localIP,
	}
}

func (sender *udpSenderDF) dial() error {
	if sender.socket != nil {
		if err := sender.socket.Close(); err != nil {
			return err
		}

		sender.socket = nil
	}

	laddr := &net.IPAddr{IP: sender.localIP}
	raddr := &net.IPAddr{IP: sender.remoteIP}
	s, err := net.DialIP("ip4:UDP", laddr, raddr)
	if err != nil {
		return err
	}

	f, err := s.File()
	if err != nil {
		return err
	}

	defer f.Close()

	// This makes sure all packets we send out have DF set on them.
	err = syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_IP, syscall.IP_MTU_DISCOVER, syscall.IP_PMTUDISC_DO)
	if err != nil {
		return err
	}

	sender.socket = s
	return nil
}

func (sender *udpSenderDF) send(msg []byte, raddr *net.UDPAddr) error {
	// Ensure we have a socket sending to the right IP address
	if sender.socket == nil || !sender.remoteIP.Equal(raddr.IP) {
		sender.remoteIP = raddr.IP
		if err := sender.dial(); err != nil {
			return err
		}
	}

	sender.udpHeader.DstPort = layers.UDPPort(raddr.Port)
	payload := gopacket.Payload(msg)
	err := gopacket.SerializeLayers(sender.ipBuf, sender.opts, sender.udpHeader, &payload)
	if err != nil {
		return err
	}

	packet := sender.ipBuf.Bytes()
	_, err = sender.socket.Write(packet)
	if err == nil || PosixError(err) != syscall.EMSGSIZE {
		return err
	}

	f, err := sender.socket.File()
	if err != nil {
		return err
	}
	defer f.Close()

	log.Print("EMSGSIZE on send, expecting PMTU update (IP packet was ", len(packet), " bytes, payload was ", len(msg), " bytes)")
	pmtu, err := syscall.GetsockoptInt(int(f.Fd()), syscall.IPPROTO_IP, syscall.IP_MTU)
	if err != nil {
		return err
	}

	return msgTooBigError{underlayPMTU: pmtu}
}

type msgTooBigError struct {
	underlayPMTU int // actual pmtu, i.e. what the kernel told us
}

func (mtbe msgTooBigError) Error() string {
	return fmt.Sprint("Msg too big error. PMTU is ", mtbe.underlayPMTU)
}

func (sender *udpSenderDF) close() error {
	if sender.socket == nil {
		return nil
	}

	return sender.socket.Close()
}

func udpAddrsEqual(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port && a.Zone == b.Zone
}

func allZeros(s []byte) bool {
	for _, b := range s {
		if b != byte(0) {
			return false
		}
	}

	return true
}

func setTimer(timer *time.Timer, d time.Duration) *time.Timer {
	if timer == nil {
		return time.NewTimer(d)
	}

	timer.Reset(d)
	return timer

}

func timerChan(timer *time.Timer) <-chan time.Time {
	if timer != nil {
		return timer.C
	}
	return nil
}

func tickerChan(ticker *time.Ticker) <-chan time.Time {
	if ticker != nil {
		return ticker.C
	}
	return nil
}

func makeUDPAddr(addr *net.TCPAddr) *net.UDPAddr {
	return &net.UDPAddr{IP: addr.IP, Port: addr.Port, Zone: addr.Zone}
}

// Look inside an error produced by the net package to get to the
// syscall.Errno at the root of the problem.
func PosixError(err error) error {
	if operr, ok := err.(*net.OpError); ok {
		err = operr.Err
	}

	// go1.5 wraps an Errno inside a SyscallError inside an OpError
	if scerr, ok := err.(*os.SyscallError); ok {
		err = scerr.Err
	}

	return err
}
