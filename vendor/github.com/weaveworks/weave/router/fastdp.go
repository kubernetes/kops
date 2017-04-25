package router

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/weaveworks/go-odp/odp"
	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/net/ipsec"
)

// The virtual bridge accepts packets from ODP vports and the router
// port (i.e. InjectPacket).  We need a map key to index those
// possibilities:
type bridgePortID struct {
	vport  odp.VportID
	router bool
}

// A bridgeSender sends out a packet from the virtual bridge
type bridgeSender func(key PacketKey, lock *fastDatapathLock) FlowOp

// A missHandler handles an ODP miss
type missHandler func(fks odp.FlowKeys, lock *fastDatapathLock) FlowOp

type FastDatapath struct {
	lock             sync.Mutex // guards state and synchronises use of dpif
	iface            *net.Interface
	dpif             *odp.Dpif
	dp               odp.DatapathHandle
	deleteFlowsCount uint64
	missCount        uint64
	missHandlers     map[odp.VportID]missHandler
	localPeer        *mesh.Peer
	peers            *mesh.Peers
	overlayConsumer  OverlayConsumer
	ipsec            *ipsec.IPSec

	// Bridge state: How to send to the given bridge port
	sendToPort map[bridgePortID]bridgeSender

	// How to send to a given destination MAC
	sendToMAC map[MAC]bridgeSender

	// MACs seen on the bridge recently
	seenMACs map[MAC]struct{}

	// vxlan vports associated with the given UDP ports
	vxlanUDPPorts    map[int]odp.VportID
	vxlanVportIDs    map[odp.VportID]struct{}
	mainVxlanVportID odp.VportID
	mainVxlanUDPPort int

	// A singleton pool for the occasions when we need to decode
	// the packet.
	dec *EthernetDecoder

	// forwarders by remote peer
	forwarders map[mesh.PeerName]*fastDatapathForwarder
}

func NewFastDatapath(iface *net.Interface, port int, encryptionEnabled bool) (*FastDatapath, error) {
	var ipSec *ipsec.IPSec

	dpif, err := odp.NewDpif()
	if err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			dpif.Close()
		}
	}()

	dp, err := dpif.LookupDatapath(iface.Name)
	if err != nil {
		return nil, err
	}

	if encryptionEnabled {
		var err error
		if ipSec, err = ipsec.New(log); err != nil {
			return nil, errors.Wrap(err, "ipsec new")
		}
		if err := ipSec.Flush(false); err != nil {
			return nil, errors.Wrap(err, "ipsec flush")
		}
	}

	fastdp := &FastDatapath{
		iface:         iface,
		dpif:          dpif,
		dp:            dp,
		missHandlers:  make(map[odp.VportID]missHandler),
		ipsec:         ipSec,
		sendToPort:    nil,
		sendToMAC:     make(map[MAC]bridgeSender),
		seenMACs:      make(map[MAC]struct{}),
		vxlanUDPPorts: make(map[int]odp.VportID),
		vxlanVportIDs: make(map[odp.VportID]struct{}),
		forwarders:    make(map[mesh.PeerName]*fastDatapathForwarder),
	}

	// This delete happens asynchronously in the kernel, meaning that
	// we can sometimes fail to recreate the vxlan vport with EADDRINUSE -
	// consequently we retry a small number of times in
	// getVxlanVportIDHarder() to compensate.
	if err := fastdp.deleteVxlanVports(); err != nil {
		return nil, err
	}

	if err := fastdp.deleteFlows(); err != nil {
		return nil, err
	}

	// We use the weave port number plus 1 for vxlan.  Hard-coding
	// this relationship may seem dubious, but there is no moral
	// difference between this and requiring that the sleeve UDP
	// port number is the same as the TCP port number.  The hard
	// part would be not adding weaver flags to allow the port
	// numbers to be independent, but working out how to specify
	// them on the connecting side.  So we can wait to find out if
	// anyone wants that.
	fastdp.mainVxlanUDPPort = port + 1
	fastdp.mainVxlanVportID, err = fastdp.getVxlanVportIDHarder(fastdp.mainVxlanUDPPort, 5, time.Millisecond*10)
	if err != nil {
		return nil, err
	}

	// need to lock before we might receive events
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	if _, err := dp.ConsumeMisses(fastdp); err != nil {
		return nil, err
	}

	if _, err := dp.ConsumeVportEvents(fastdp); err != nil {
		return nil, err
	}

	vports, err := dp.EnumerateVports()
	if err != nil {
		return nil, err
	}

	for _, vport := range vports {
		fastdp.makeBridgeVport(vport)
	}

	success = true
	go fastdp.run()
	return fastdp, nil
}

func (fastdp *FastDatapath) Close() error {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()
	err := fastdp.dpif.Close()
	fastdp.dpif = nil
	return err
}

// While processing a packet, we can potentially acquire and drop the
// FastDatapath lock many times (acquiring it to acceess FastDatapath
// state, and invoke ODP operations; dropping it to invoke callbacks
// that may re-enter the FastDatapath).  A fastDatapathLock
// coordinates this process.
type fastDatapathLock struct {
	fastdp *FastDatapath
	locked bool

	// While the lock is dropped, deleteFlows could be called.  We
	// need to detect when this happens and avoid creating flows,
	// because they may be based on stale information.
	deleteFlowsCount uint64
}

func (fastdp *FastDatapath) startLock() fastDatapathLock {
	fastdp.lock.Lock()
	return fastDatapathLock{
		fastdp:           fastdp,
		locked:           true,
		deleteFlowsCount: fastdp.deleteFlowsCount,
	}
}

func (lock *fastDatapathLock) unlock() {
	if lock.locked {
		lock.fastdp.lock.Unlock()
		lock.locked = false
	}
}

func (lock *fastDatapathLock) relock() {
	if !lock.locked {
		lock.fastdp.lock.Lock()
		lock.locked = true
	}
}

// Bridge bits

type fastDatapathBridge struct {
	*FastDatapath
}

func (fastdp *FastDatapath) Bridge() Bridge {
	return fastDatapathBridge{fastdp}
}

func (fastdp fastDatapathBridge) Interface() *net.Interface {
	return fastdp.iface
}

func (fastdp fastDatapathBridge) String() string {
	return fmt.Sprint(fastdp.iface.Name, " (via ODP)")
}

func (fastdp fastDatapathBridge) Stats() map[string]int {
	lock := fastdp.startLock()
	defer lock.unlock()

	return map[string]int{
		"FlowMisses": int(fastdp.missCount),
	}
}

var routerBridgePortID = bridgePortID{router: true}

func (fastdp fastDatapathBridge) StartConsumingPackets(consumer BridgeConsumer) error {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	if fastdp.sendToPort[routerBridgePortID] != nil {
		return fmt.Errorf("FastDatapath already has a BridgeConsumer")
	}

	// set up delivery to the weave router port on the bridge
	fastdp.addSendToPort(routerBridgePortID,
		func(key PacketKey, lock *fastDatapathLock) FlowOp {
			// drop the FastDatapath lock in order to call
			// the consumer
			lock.unlock()
			return consumer(key)
		})
	return nil
}

func (fastdp fastDatapathBridge) InjectPacket(key PacketKey) FlowOp {
	lock := fastdp.startLock()
	defer lock.unlock()
	return fastdp.bridge(routerBridgePortID, key, &lock)
}

// Ethernet bridge implementation

func (fastdp *FastDatapath) bridge(ingress bridgePortID, key PacketKey, lock *fastDatapathLock) FlowOp {
	lock.relock()
	if fastdp.sendToMAC[key.SrcMAC] == nil {
		// Learn the source MAC
		fastdp.sendToMAC[key.SrcMAC] = fastdp.sendToPort[ingress]
		fastdp.seenMACs[key.SrcMAC] = struct{}{}
	}

	// If we know about the destination MAC, deliver it to the
	// associated port.
	if sender := fastdp.sendToMAC[key.DstMAC]; sender != nil {
		return NewMultiFlowOp(false, odpEthernetFlowKey(key), sender(key, lock))
	}

	// Otherwise, it might be a real broadcast, or it might
	// be for a MAC we don't know about yet.  Either way, we'll
	// broadcast it.
	mfop := NewMultiFlowOp(false)

	if (key.DstMAC[0] & 1) == 0 {
		// Not a real broadcast, so don't create a flow rule.
		// If we did, we'd need to delete the flows every time
		// we learned a new MAC address, or have a more
		// complicated selective invalidation scheme.
		log.Debug("fastdp: unknown dst", ingress, key)
		mfop.Add(vetoFlowCreationFlowOp{})
	} else {
		// A real broadcast
		log.Debug("fastdp: broadcast", ingress, key)
		mfop.Add(odpEthernetFlowKey(key))
	}

	// Send to all ports except the one it came in on. The
	// sendToPort map is immutable, so it is safe to iterate over
	// it even though the sender functions can drop the
	// fastDatapathLock
	for id, sender := range fastdp.sendToPort {
		if id != ingress {
			mfop.Add(sender(key, lock))
		}
	}

	return mfop
}

// Overlay bits

type fastDatapathOverlay struct {
	*FastDatapath
}

func (fastdp *FastDatapath) Overlay() NetworkOverlay {
	return fastDatapathOverlay{fastdp}
}

func (fastdp fastDatapathOverlay) InvalidateRoutes() {
	log.Debug("InvalidateRoutes")
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()
	checkWarn(fastdp.deleteFlows())
}

func (fastdp fastDatapathOverlay) InvalidateShortIDs() {
	log.Debug("InvalidateShortIDs")
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()
	checkWarn(fastdp.deleteFlows())
}

func (fastdp fastDatapathOverlay) Stop() {
	if fastdp.ipsec != nil {
		if err := fastdp.ipsec.Flush(true); err != nil {
			log.Errorf("ipsec flush failed: %s", err)
		}
	}
}

func (fastDatapathOverlay) AddFeaturesTo(features map[string]string) {
	// Nothing needed.  Fast datapath support is indicated through
	// OverlaySwitch.
}

type FastDPStatus struct {
	Vports []VportStatus
	Flows  []FlowStatus
}

type FlowStatus odp.FlowInfo

func (flowStatus *FlowStatus) MarshalJSON() ([]byte, error) {
	type jsonFlowStatus struct {
		FlowKeys []string
		Actions  []string
		Packets  uint64
		Bytes    uint64
		Used     uint64
	}

	flowKeys := make([]string, 0, len(flowStatus.FlowKeys))
	for _, flowKey := range flowStatus.FlowKeys {
		if !flowKey.Ignored() {
			flowKeys = append(flowKeys, fmt.Sprint(flowKey))
		}
	}

	actions := make([]string, 0, len(flowStatus.Actions))
	for _, action := range flowStatus.Actions {
		actions = append(actions, fmt.Sprint(action))
	}

	return json.Marshal(&jsonFlowStatus{flowKeys, actions, flowStatus.Packets, flowStatus.Bytes, flowStatus.Used})
}

type VportStatus odp.Vport

func (vport *VportStatus) MarshalJSON() ([]byte, error) {
	type jsonVportStatus struct {
		ID       odp.VportID
		Name     string
		TypeName string
	}

	return json.Marshal(&jsonVportStatus{vport.ID, vport.Spec.Name(), vport.Spec.TypeName()})
}

func (fastdp fastDatapathOverlay) Diagnostics() interface{} {
	lock := fastdp.startLock()
	defer lock.unlock()

	vports, err := fastdp.dp.EnumerateVports()
	checkWarn(err)
	vportStatuses := make([]VportStatus, 0, len(vports))
	for _, vport := range vports {
		vportStatuses = append(vportStatuses, VportStatus(vport))
	}

	flows, err := fastdp.dp.EnumerateFlows()
	checkWarn(err)
	flowStatuses := make([]FlowStatus, 0, len(flows))
	for _, flow := range flows {
		flowStatuses = append(flowStatuses, FlowStatus(flow))
	}

	return FastDPStatus{
		vportStatuses,
		flowStatuses,
	}
}

type FastDPMetrics struct {
	Flows        int
	TotalPackets uint64
	TotalBytes   uint64
}

func (s FastDPStatus) Metrics() interface{} {
	var m FastDPMetrics
	m.Flows = len(s.Flows)
	for _, flow := range s.Flows {
		m.TotalPackets += flow.Packets
		m.TotalBytes += flow.Bytes
	}
	return &m
}

func (fastdp fastDatapathOverlay) StartConsumingPackets(localPeer *mesh.Peer, peers *mesh.Peers, consumer OverlayConsumer) error {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	if fastdp.overlayConsumer != nil {
		return fmt.Errorf("FastDatapath already has an OverlayConsumer")
	}

	fastdp.localPeer = localPeer
	fastdp.peers = peers
	fastdp.overlayConsumer = consumer
	return nil
}

func (fastdp *FastDatapath) getVxlanVportIDHarder(udpPort int, retries int, duration time.Duration) (odp.VportID, error) {
	var vxlanVportID odp.VportID
	var err error
	for try := 0; try < retries; try++ {
		vxlanVportID, err = fastdp.getVxlanVportID(udpPort)
		if err == nil || err != odp.NetlinkError(syscall.EADDRINUSE) {
			return vxlanVportID, err
		}
		log.Warning("Address already in use creating vxlan vport ", udpPort, " - retrying")
		time.Sleep(duration)
	}
	return 0, err
}

func (fastdp *FastDatapath) getVxlanVportID(udpPort int) (odp.VportID, error) {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	if vxlanVportID, present := fastdp.vxlanUDPPorts[udpPort]; present {
		return vxlanVportID, nil
	}

	name := fmt.Sprintf("vxlan-%d", udpPort)
	vxlanVportID, err := fastdp.dp.CreateVport(
		odp.NewVxlanVportSpec(name, uint16(udpPort)))
	if err != nil {
		return 0, err
	}

	// If a netdev for the vxlan vport exists, we need to do an extra check
	// to bypass the kernel bug which makes the vxlan creation to complete
	// successfully regardless whether there were any errors when binding
	// to the given UDP port.
	if link, err := netlink.LinkByName(name); err == nil {
		if link.Attrs().Flags&net.FlagUp == 0 {
			// The netdev interface is down, so most likely bringing it up
			// has failed due to the UDP port being in use.
			if err := fastdp.dp.DeleteVport(vxlanVportID); err != nil {
				log.Warning("Unable to remove vxlan vport %d: %s", vxlanVportID, err)
			}
			return 0, odp.NetlinkError(syscall.EADDRINUSE)
		}
	}

	fastdp.vxlanUDPPorts[udpPort] = vxlanVportID
	fastdp.vxlanVportIDs[vxlanVportID] = struct{}{}
	fastdp.missHandlers[vxlanVportID] = func(fks odp.FlowKeys, lock *fastDatapathLock) FlowOp {
		log.Debug("ODP miss: ", fks, " on port ", vxlanVportID)
		tunnel := fks[odp.OVS_KEY_ATTR_TUNNEL].(odp.TunnelFlowKey)
		tunKey := tunnel.Key()

		lock.relock()
		consumer := fastdp.overlayConsumer
		if consumer == nil {
			return vetoFlowCreationFlowOp{}
		}

		srcPeer, dstPeer := fastdp.extractPeers(tunKey.TunnelId)
		if srcPeer == nil || dstPeer == nil {
			return vetoFlowCreationFlowOp{}
		}

		lock.unlock()
		pk := flowKeysToPacketKey(fks)
		var zeroMAC MAC
		if pk.SrcMAC == zeroMAC && pk.DstMAC == zeroMAC {
			return vxlanSpecialPacketFlowOp{
				fastdp:  fastdp,
				srcPeer: srcPeer,
				sender: &net.UDPAddr{
					IP:   net.IP(tunKey.Ipv4Src[:]),
					Port: udpPort,
				},
			}
		}

		key := ForwardPacketKey{
			SrcPeer:   srcPeer,
			DstPeer:   dstPeer,
			PacketKey: pk,
		}

		var tunnelFlowKey odp.TunnelFlowKey
		tunnelFlowKey.SetTunnelId(tunKey.TunnelId)
		tunnelFlowKey.SetIpv4Src(tunKey.Ipv4Src)
		tunnelFlowKey.SetIpv4Dst(tunKey.Ipv4Dst)

		return NewMultiFlowOp(false, odpFlowKey(tunnelFlowKey), consumer(key))
	}

	return vxlanVportID, nil
}

func (fastdp *FastDatapath) extractPeers(tunnelID [8]byte) (*mesh.Peer, *mesh.Peer) {
	vni := binary.BigEndian.Uint64(tunnelID[:])
	srcPeer := fastdp.peers.FetchByShortID(mesh.PeerShortID(vni & 0xfff))
	dstPeer := fastdp.peers.FetchByShortID(mesh.PeerShortID((vni >> 12) & 0xfff))
	return srcPeer, dstPeer
}

type vxlanSpecialPacketFlowOp struct {
	NonDiscardingFlowOp
	fastdp  *FastDatapath
	srcPeer *mesh.Peer
	sender  *net.UDPAddr
}

func (op vxlanSpecialPacketFlowOp) Process(frame []byte, dec *EthernetDecoder, broadcast bool) {
	op.fastdp.lock.Lock()
	fwd := op.fastdp.forwarders[op.srcPeer.Name]
	op.fastdp.lock.Unlock()

	if fwd != nil && dec.IsSpecial() {
		fwd.handleVxlanSpecialPacket(frame, op.sender)
	}
}

type fastDatapathForwarder struct {
	fastdp         *FastDatapath
	remotePeer     *mesh.Peer
	localIP        [4]byte
	sendControlMsg func(byte, []byte) error
	connUID        uint64
	vxlanVportID   odp.VportID

	sessionKey                 *[32]byte
	isEncrypted                bool
	isOutboundIPSecEstablished bool

	lock              sync.RWMutex
	confirmed         bool
	remoteAddr        *net.UDPAddr
	heartbeatInterval time.Duration
	heartbeatTimer    *time.Timer // for sending
	heartbeatTimeout  *time.Timer // for receiving
	ackedHeartbeat    bool
	stopChan          chan struct{}
	stopped           bool

	establishedChan chan struct{}
	errorChan       chan error
}

func (fastdp fastDatapathOverlay) PrepareConnection(params mesh.OverlayConnectionParams) (mesh.OverlayConnection, error) {
	vxlanVportID := fastdp.mainVxlanVportID
	vxlanUDPPort := fastdp.mainVxlanUDPPort

	remoteAddr := makeUDPAddr(params.RemoteAddr)
	if params.Outbound {
		var err error
		// The provided address contains the main weave port
		// number to connect to.  We need to derive the vxlan
		// port number from that.
		vxlanRemoteAddr := *remoteAddr
		vxlanRemoteAddr.Port++
		remoteAddr = &vxlanRemoteAddr
		vxlanUDPPort = remoteAddr.Port
		vxlanVportID, err = fastdp.getVxlanVportID(vxlanUDPPort)
		if err != nil {
			return nil, err
		}
	} else {
		remoteAddr.Port = vxlanUDPPort
	}

	localIP, err := ipv4Bytes(params.LocalAddr.IP)
	if err != nil {
		return nil, err
	}

	fwd := &fastDatapathForwarder{
		fastdp:         fastdp.FastDatapath,
		remotePeer:     params.RemotePeer,
		localIP:        localIP,
		sendControlMsg: params.SendControlMessage,
		connUID:        params.ConnUID,
		vxlanVportID:   vxlanVportID,
		sessionKey:     params.SessionKey,

		remoteAddr:        remoteAddr,
		heartbeatInterval: FastHeartbeat,
		stopChan:          make(chan struct{}),

		establishedChan: make(chan struct{}),
		errorChan:       make(chan error, 1),
	}

	return fwd, nil
}

func ipv4Bytes(ip net.IP) (res [4]byte, err error) {
	ipv4 := ip.To4()
	if ipv4 != nil {
		copy(res[:], ipv4)
	} else {
		err = fmt.Errorf("IP address %s is not IPv4", ip)
	}
	return
}

func (fwd *fastDatapathForwarder) logPrefix() string {
	return fmt.Sprintf("fastdp ->[%s|%s]: ", fwd.remoteAddr, fwd.remotePeer)
}

func (fwd *fastDatapathForwarder) Confirm() {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()

	if fwd.confirmed {
		log.Fatal(fwd.logPrefix(), "already confirmed")
	}

	if fwd.fastdp.ipsec != nil && fwd.sessionKey != nil {
		fwd.isEncrypted = true
		log.Info("Setting up IPsec between ", fwd.fastdp.localPeer, " and ", fwd.remotePeer)
		err := fwd.fastdp.ipsec.InitSALocal(
			fwd.fastdp.localPeer.Name, fwd.remotePeer.Name, fwd.connUID,
			net.IP(fwd.localIP[:]), fwd.remoteAddr.IP,
			fwd.remoteAddr.Port,
			fwd.sessionKey,
			func(msg []byte) error {
				return fwd.sendControlMsg(FastDatapathCryptoInitSARemote, msg)
			},
		)
		if err != nil {
			log.Error(fwd.logPrefix(), "ipsec init SA local failed: ", err)
			fwd.handleError(err)
			return
		}
	}

	log.Debug(fwd.logPrefix(), "confirmed")
	fwd.fastdp.addForwarder(fwd.remotePeer.Name, fwd)
	fwd.confirmed = true

	if fwd.remoteAddr != nil && (!fwd.isEncrypted || fwd.isOutboundIPSecEstablished) {
		// have the goroutine send a heartbeat straight away
		fwd.heartbeatTimer = time.NewTimer(0)
	} else {
		// we'll reset the timer when we learn the remote ip
		fwd.heartbeatTimer = time.NewTimer(MaxDuration)
	}

	fwd.heartbeatTimeout = time.NewTimer(HeartbeatTimeout)
	go fwd.doHeartbeats()
}

func (fwd *fastDatapathForwarder) EstablishedChannel() <-chan struct{} {
	return fwd.establishedChan
}

func (fwd *fastDatapathForwarder) ErrorChannel() <-chan error {
	return fwd.errorChan
}

func (fwd *fastDatapathForwarder) doHeartbeats() {
	var err error

	for err == nil {
		select {
		case <-fwd.heartbeatTimer.C:
			if fwd.confirmed {
				fwd.sendHeartbeat()
			}
			fwd.heartbeatTimer.Reset(fwd.heartbeatInterval)

		case <-fwd.heartbeatTimeout.C:
			err = fmt.Errorf("timed out waiting for vxlan heartbeat")

		case <-fwd.stopChan:
			return
		}
	}

	fwd.lock.Lock()
	defer fwd.lock.Unlock()
	fwd.handleError(err)
}

// Handle an error which leads to notifying the listener and
// termination of the forwarder
func (fwd *fastDatapathForwarder) handleError(err error) {
	if err == nil {
		return
	}

	select {
	case fwd.errorChan <- err:
	default:
	}

	// stop the heartbeat goroutine
	if !fwd.stopped {
		fwd.stopped = true
		close(fwd.stopChan)
	}
}

func (fwd *fastDatapathForwarder) sendHeartbeat() {
	fwd.lock.RLock()
	log.Debug(fwd.logPrefix(), "sendHeartbeat")

	// the heartbeat payload consists of the 64-bit connection uid
	// followed by the 16-bit packet size.
	buf := make([]byte, EthernetOverhead+fwd.fastdp.iface.MTU)
	binary.BigEndian.PutUint64(buf[EthernetOverhead:], fwd.connUID)
	binary.BigEndian.PutUint16(buf[EthernetOverhead+8:], uint16(len(buf)))

	dec := NewEthernetDecoder()
	dec.DecodeLayers(buf)
	pk := ForwardPacketKey{
		PacketKey: dec.PacketKey(),
		SrcPeer:   fwd.fastdp.localPeer,
		DstPeer:   fwd.remotePeer,
	}
	fwd.lock.RUnlock()

	if fop := fwd.Forward(pk); fop != nil {
		fop.Process(buf, dec, false)
	}
}

const (
	FastDatapathHeartbeatAck = iota
	FastDatapathCryptoInitSARemote
)

func (fwd *fastDatapathForwarder) handleVxlanSpecialPacket(frame []byte, sender *net.UDPAddr) {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()

	log.Debug(fwd.logPrefix(), "handleVxlanSpecialPacket")

	// the only special packet type is a heartbeat
	if len(frame) < EthernetOverhead+10 {
		log.Warning(fwd.logPrefix(), "short vxlan special packet: ", len(frame), " bytes")
		return
	}

	if binary.BigEndian.Uint64(frame[EthernetOverhead:]) != fwd.connUID ||
		uint16(len(frame)) != binary.BigEndian.Uint16(frame[EthernetOverhead+8:]) {
		return
	}

	if fwd.remoteAddr == nil {
		fwd.remoteAddr = sender

		if fwd.confirmed {
			fwd.heartbeatTimer.Reset(0)
		}
	} else if !udpAddrsEqual(fwd.remoteAddr, sender) {
		log.Info(fwd.logPrefix(), "Peer IP address changed to ", sender)
		fwd.remoteAddr = sender
	}

	if !fwd.ackedHeartbeat {
		fwd.ackedHeartbeat = true
		fwd.handleError(fwd.sendControlMsg(FastDatapathHeartbeatAck, nil))
	}

	// we can receive a heartbeat before Confirm() has set up
	// heartbeatTimeout
	if fwd.heartbeatTimeout != nil {
		fwd.heartbeatTimeout.Reset(HeartbeatTimeout)
	}
}

func (fwd *fastDatapathForwarder) ControlMessage(tag byte, msg []byte) {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()

	switch tag {
	case FastDatapathHeartbeatAck:
		fwd.handleHeartbeatAck()
	case FastDatapathCryptoInitSARemote:
		fwd.handleCryptoInitSARemote(msg)

	default:
		log.Info(fwd.logPrefix(), "Ignoring unknown control message: ", tag)
	}
}

func (fwd *fastDatapathForwarder) Attrs() map[string]interface{} {
	return map[string]interface{}{"name": "fastdp", "mtu": fwd.fastdp.iface.MTU}
}

func (fwd *fastDatapathForwarder) handleHeartbeatAck() {
	log.Debug(fwd.logPrefix(), "handleHeartbeatAck")

	if fwd.heartbeatInterval != SlowHeartbeat {
		close(fwd.establishedChan)
		fwd.heartbeatInterval = SlowHeartbeat
		if fwd.heartbeatTimer != nil {
			fwd.heartbeatTimer.Reset(fwd.heartbeatInterval)
		}
	}
}

func (fwd *fastDatapathForwarder) handleCryptoInitSARemote(msg []byte) {
	log.Info(fwd.logPrefix(), "IPSec init SA remote")
	err := fwd.fastdp.ipsec.InitSARemote(
		msg,
		fwd.fastdp.localPeer.Name, fwd.remotePeer.Name, fwd.connUID,
		net.IP(fwd.localIP[:]), fwd.remoteAddr.IP, fwd.remoteAddr.Port,
		fwd.sessionKey,
	)
	if err != nil {
		log.Warning(fwd.logPrefix(), "IPSec init SA remote failed: ", err)
		fwd.handleError(err)
		return
	}

	// FastDatapathCryptoInitSARemote can be received before Confirm'ing
	// connection, thus before InitSALocal.
	if fwd.confirmed && !fwd.isOutboundIPSecEstablished {
		fwd.isOutboundIPSecEstablished = true
		fwd.heartbeatTimer.Reset(0)
	}
}

func (fwd *fastDatapathForwarder) Forward(key ForwardPacketKey) FlowOp {
	if !key.SrcPeer.HasShortID || !key.DstPeer.HasShortID {
		return nil
	}

	fwd.lock.RLock()
	defer fwd.lock.RUnlock()

	if fwd.remoteAddr == nil {
		// Returning a DiscardingFlowOp would discard the
		// packet, but also result in a flow rule, which we
		// would have to invalidate when we learn the remote
		// IP.  So for now, just prevent flows.
		return vetoFlowCreationFlowOp{}
	}

	remoteIP, err := ipv4Bytes(fwd.remoteAddr.IP)
	if err != nil {
		log.Error(err)
		return DiscardingFlowOp{}
	}

	var sta odp.SetTunnelAction
	sta.SetTunnelId(tunnelIDFor(key))
	sta.SetIpv4Src(fwd.localIP)
	sta.SetIpv4Dst(remoteIP)
	sta.SetTos(0)
	sta.SetTtl(64)
	sta.SetDf(true)
	sta.SetCsum(false)
	return fwd.fastdp.odpActions(sta, odp.NewOutputAction(fwd.vxlanVportID))
}

func tunnelIDFor(key ForwardPacketKey) (tunnelID [8]byte) {
	src := uint64(key.SrcPeer.ShortID)
	dst := uint64(key.DstPeer.ShortID)
	binary.BigEndian.PutUint64(tunnelID[:], src|dst<<12)
	return
}

func (fwd *fastDatapathForwarder) Stop() {
	// Might be nice to delete all the relevant flows here, but we
	// can just let them expire.
	fwd.fastdp.removeForwarder(fwd.remotePeer.Name, fwd)

	fwd.lock.Lock()
	defer fwd.lock.Unlock()
	fwd.sendControlMsg = func(byte, []byte) error { return nil }

	if fwd.isEncrypted {
		localIP := net.IP(fwd.localIP[:])
		log.Info("Destroying IPsec between ", fwd.fastdp.localPeer, " and ", fwd.remotePeer)
		err := fwd.fastdp.ipsec.Destroy(
			fwd.fastdp.localPeer.Name, fwd.remotePeer.Name, fwd.connUID,
			localIP, fwd.remoteAddr.IP, fwd.remoteAddr.Port,
		)
		if err != nil {
			log.Errorf("ipsec destroy failed: %s", err)
		}
	}

	// stop the heartbeat goroutine
	if !fwd.stopped {
		fwd.stopped = true
		close(fwd.stopChan)
	}
}

func (fastdp *FastDatapath) addForwarder(peer mesh.PeerName, fwd *fastDatapathForwarder) {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	// We shouldn't have two confirmed forwarders to the same
	// remotePeer, due to the checks in LocalPeer AddConnection.
	fastdp.forwarders[peer] = fwd
}

func (fastdp *FastDatapath) removeForwarder(peer mesh.PeerName, fwd *fastDatapathForwarder) {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()
	if fastdp.forwarders[peer] == fwd {
		delete(fastdp.forwarders, peer)
	}
}

func (fastdp *FastDatapath) deleteFlows() error {
	fastdp.deleteFlowsCount++

	flows, err := fastdp.dp.EnumerateFlows()
	if err != nil {
		return err
	}

	for _, flow := range flows {
		err = fastdp.dp.DeleteFlow(flow.FlowKeys)
		if err != nil && !odp.IsNoSuchFlowError(err) {
			return err
		}
	}

	return nil
}

func (fastdp *FastDatapath) deleteVxlanVports() error {
	vports, err := fastdp.dp.EnumerateVports()
	if err != nil {
		return err
	}

	for _, vport := range vports {
		if vport.Spec.TypeName() != "vxlan" {
			continue
		}

		err = fastdp.dp.DeleteVport(vport.ID)
		if err != nil && !odp.IsNoSuchVportError(err) {
			return err
		}
	}

	return nil
}

func (fastdp *FastDatapath) run() {
	expireMACsCh := time.Tick(10 * time.Minute)
	expireFlowsCh := time.Tick(5 * time.Minute)

	for {
		select {
		case <-expireMACsCh:
			fastdp.expireMACs()

		case <-expireFlowsCh:
			fastdp.expireFlows()
		}
	}
}

func (fastdp *FastDatapath) expireMACs() {
	lock := fastdp.startLock()
	defer lock.unlock()

	for mac := range fastdp.sendToMAC {
		if _, present := fastdp.seenMACs[mac]; !present {
			delete(fastdp.sendToMAC, mac)
		}
	}

	fastdp.seenMACs = make(map[MAC]struct{})
}

func (fastdp *FastDatapath) expireFlows() {
	lock := fastdp.startLock()
	defer lock.unlock()

	flows, err := fastdp.dp.EnumerateFlows()
	checkWarn(err)

	for _, flow := range flows {
		if flow.Used == 0 {
			log.Debug("Expiring flow ", flow.FlowSpec)
			err = fastdp.dp.DeleteFlow(flow.FlowKeys)
		} else {
			fastdp.touchFlow(flow.FlowKeys, &lock)
			err = fastdp.dp.ClearFlow(flow.FlowSpec)
		}

		if err != nil && !odp.IsNoSuchFlowError(err) {
			log.Warn(err)
		}
	}
}

// The router needs to know which flows are active in order to
// maintain its MAC->peer table.  We do this by querying the router
// without an actual packet being involved.  Maybe it's
// worth devising a more unified approach in the future.
func (fastdp *FastDatapath) touchFlow(fks odp.FlowKeys, lock *fastDatapathLock) {
	// All the flows we create should have an ingress key, but we
	// check here just in case we encounter one from somewhere
	// else.
	ingressKey, present := fks[odp.OVS_KEY_ATTR_IN_PORT]
	if present {
		ingress := ingressKey.(odp.InPortFlowKey).VportID()
		handler := fastdp.getMissHandler(ingress)
		if handler != nil {
			handler(fks, lock)
			lock.relock()
		}
	}
}

func (fastdp *FastDatapath) Error(err error, stopped bool) {
	if stopped {
		log.Fatal("Error while listeniing on ODP datapath: ", err)
	}

	log.Error("Error while listening on ODP datapath: ", err)
}

func (fastdp *FastDatapath) Miss(packet []byte, fks odp.FlowKeys) error {
	ingress := fks[odp.OVS_KEY_ATTR_IN_PORT].(odp.InPortFlowKey).VportID()

	lock := fastdp.startLock()
	defer lock.unlock()

	fastdp.missCount++

	handler := fastdp.getMissHandler(ingress)
	if handler == nil {
		log.Debug("ODP miss (no handler): ", fks, " on port ", ingress)
		return nil
	}

	// Always include the ingress vport in the flow key.  While
	// this is not strictly necessary in some cases (e.g. for
	// delivery to a local netdev based on the dest MAC),
	// including the ingress in every flow makes things simpler
	// in touchFlow.
	mfop := NewMultiFlowOp(false, handler(fks, &lock), odpFlowKey(odp.NewInPortFlowKey(ingress)))
	fastdp.send(mfop, packet, &lock)
	return nil
}

func (fastdp *FastDatapath) getMissHandler(ingress odp.VportID) missHandler {
	handler := fastdp.missHandlers[ingress]
	if handler == nil {
		vport, err := fastdp.dp.LookupVport(ingress)
		if err != nil {
			log.Error(err)
			return nil
		}

		fastdp.makeBridgeVport(vport)
	}

	return handler
}

func (fastdp *FastDatapath) VportCreated(dpid odp.DatapathID, vport odp.Vport) error {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	if _, present := fastdp.missHandlers[vport.ID]; !present {
		fastdp.makeBridgeVport(vport)
	}

	return nil
}

func (fastdp *FastDatapath) VportDeleted(dpid odp.DatapathID, vport odp.Vport) error {
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()

	// there might be flow rules that still refer to the id of
	// this vport.  But we just allow them to expire.  Unless we
	// want prompt migration of MAC addresses, that should be
	// fine.
	delete(fastdp.missHandlers, vport.ID)
	fastdp.deleteSendToPort(bridgePortID{vport: vport.ID})
	return nil
}

func (fastdp *FastDatapath) makeBridgeVport(vport odp.Vport) {
	// Set up a bridge port for netdev and internal vports.  vxlan
	// vports are handled separately, as they do not correspond to
	// bridge ports (we set up the miss handler for them in
	// getVxlanVportID).
	typ := vport.Spec.TypeName()
	if typ != "netdev" && typ != "internal" {
		return
	}

	vportID := vport.ID

	// Sending to the bridge port outputs on the vport:
	fastdp.addSendToPort(bridgePortID{vport: vportID},
		func(_ PacketKey, _ *fastDatapathLock) FlowOp {
			return fastdp.odpActions(odp.NewOutputAction(vportID))
		})

	// Delete flows, in order to recalculate flows for broadcasts
	// on the bridge.
	checkWarn(fastdp.deleteFlows())

	// Packets coming from the netdev are processed by the bridge
	fastdp.missHandlers[vportID] = func(flowKeys odp.FlowKeys, lock *fastDatapathLock) FlowOp {
		return fastdp.bridge(bridgePortID{vport: vportID}, flowKeysToPacketKey(flowKeys), lock)
	}
}

func flowKeysToPacketKey(fks odp.FlowKeys) PacketKey {
	eth := fks[odp.OVS_KEY_ATTR_ETHERNET].(odp.EthernetFlowKey).Key()
	return PacketKey{SrcMAC: eth.EthSrc, DstMAC: eth.EthDst}
}

// The sendToPort map is read-only, so this method does the copy in
// order to add an entry.
func (fastdp *FastDatapath) addSendToPort(portID bridgePortID, sender bridgeSender) {
	sendToPort := map[bridgePortID]bridgeSender{portID: sender}
	for id, sender := range fastdp.sendToPort {
		sendToPort[id] = sender
	}
	fastdp.sendToPort = sendToPort
}

func (fastdp *FastDatapath) deleteSendToPort(portID bridgePortID) {
	sendToPort := make(map[bridgePortID]bridgeSender)
	for id, sender := range fastdp.sendToPort {
		if id != portID {
			sendToPort[id] = sender
		}
	}
	fastdp.sendToPort = sendToPort
}

// Send a packet, creating a corresponding ODP flow rule if possible
func (fastdp *FastDatapath) send(fops FlowOp, frame []byte, lock *fastDatapathLock) {
	// Gather the actions from actionFlowOps, execute any others
	var dec *EthernetDecoder
	flow := odp.NewFlowSpec()
	createFlow := true

	for _, xfop := range FlattenFlowOp(fops) {
		switch fop := xfop.(type) {
		case interface {
			updateFlowSpec(*odp.FlowSpec)
		}:
			fop.updateFlowSpec(&flow)
		case vetoFlowCreationFlowOp:
			createFlow = false
		default:
			if xfop.Discards() {
				continue
			}

			// A foreign FlowOp (e.g. a sleeve forwarding
			// FlowOp), so send the packet through the
			// FlowOp interface, decoding the packet
			// lazily.
			if dec == nil {
				dec = fastdp.takeDecoder(lock)
				dec.DecodeLayers(frame)

				// If we are sending the packet
				// through the FlowOp interface, we
				// mustn't create a flow, as that
				// could prevent the proper handling
				// of similar packets in the future.
				createFlow = false
			}

			if len(dec.decoded) != 0 {
				lock.unlock()
				fop.Process(frame, dec, false)
			}
		}
	}

	if fastdp.isHairpinFlow(&flow) {
		log.Error("Vetoed installation of hairpin flow ", flow)
		return
	}

	if dec != nil {
		// put the decoder back
		lock.relock()
		fastdp.dec = dec
	}

	if len(flow.Actions) != 0 {
		lock.relock()
		checkWarn(fastdp.dp.Execute(frame, nil, flow.Actions))
	}

	if createFlow {
		lock.relock()
		// if the fastdp's deleteFlowsCount changed since we
		// initially locked it, then we might have created a
		// flow on the basis of stale information.  It's fine
		// to handle one packet like that, but it would be bad
		// to introduce a stale flow.
		if lock.deleteFlowsCount == fastdp.deleteFlowsCount {
			log.Debug("Creating ODP flow ", flow)
			checkWarn(fastdp.dp.CreateFlow(flow))
		}
	}
}

// Get the EthernetDecoder from the singleton pool
func (fastdp *FastDatapath) takeDecoder(lock *fastDatapathLock) *EthernetDecoder {
	lock.relock()
	dec := fastdp.dec
	if dec == nil {
		dec = NewEthernetDecoder()
	} else {
		fastdp.dec = nil
	}
	return dec
}

// A isHairpinFlow checks whether the flow is created due to enabled hairpin
// mode on the weave bridge port which attaches the datapath. Such flow is
// identified by either:
//
// * in_port == out_port, where in_port is non-vxlan vport;
// * a packet is sent back to a vxlan tunnel it has been received from and
//   the tunnel id is either the same or dstPeer and srcPeer are reversed.
func (fastdp *FastDatapath) isHairpinFlow(flow *odp.FlowSpec) bool {
	var (
		vxlanKey odp.TunnelAttrs
		inVport  odp.VportID
		inVxlan  bool
	)

	for _, key := range flow.FlowKeys {
		switch k := key.(type) {
		case odp.InPortFlowKey:
			inVport = k.VportID()
		case odp.TunnelFlowKey:
			inVxlan = true
			vxlanKey = k.Key()
		}
	}

	for _, action := range flow.Actions {
		switch a := action.(type) {
		case odp.SetTunnelAction:
			if inVxlan && a.TunnelAttrs.TunnelId == vxlanKey.TunnelId &&
				a.TunnelAttrs.Ipv4Src == vxlanKey.Ipv4Dst &&
				a.TunnelAttrs.Ipv4Dst == vxlanKey.Ipv4Src {
				return true
			}
		case odp.OutputAction:
			if a.VportID() == inVport {
				if _, ok := fastdp.vxlanVportIDs[a.VportID()]; !ok {
					return true
				}
			}
		}
	}

	return false
}

type odpActionsFlowOp struct {
	NonDiscardingFlowOp
	fastdp  *FastDatapath
	actions []odp.Action
}

func (fastdp *FastDatapath) odpActions(actions ...odp.Action) FlowOp {
	return odpActionsFlowOp{
		fastdp:  fastdp,
		actions: actions,
	}
}

func (fop odpActionsFlowOp) updateFlowSpec(flow *odp.FlowSpec) {
	flow.AddActions(fop.actions)
}

func (fop odpActionsFlowOp) Process(frame []byte, dec *EthernetDecoder, broadcast bool) {
	fastdp := fop.fastdp
	fastdp.lock.Lock()
	defer fastdp.lock.Unlock()
	checkWarn(fastdp.dp.Execute(frame, nil, fop.actions))
}

// A vetoFlowCreationFlowOp flags that no flow should be created
type vetoFlowCreationFlowOp struct {
	DiscardingFlowOp
}

// A odpFlowKeyFlowOp adds a FlowKey to the resulting flow
type odpFlowKeyFlowOp struct {
	DiscardingFlowOp
	key odp.FlowKey
}

func odpFlowKey(key odp.FlowKey) FlowOp {
	return odpFlowKeyFlowOp{key: key}
}

func (fop odpFlowKeyFlowOp) updateFlowSpec(flow *odp.FlowSpec) {
	flow.AddKey(fop.key)
}

func odpEthernetFlowKey(key PacketKey) FlowOp {
	fk := odp.NewEthernetFlowKey()
	fk.SetEthSrc(key.SrcMAC)
	fk.SetEthDst(key.DstMAC)
	return odpFlowKeyFlowOp{key: fk}
}
