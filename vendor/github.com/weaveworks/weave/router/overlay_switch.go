package router

import (
	"fmt"
	"strings"
	"sync"

	"github.com/weaveworks/mesh"
)

// OverlaySwitch selects which overlay to use, from a set of
// subsidiary overlays.  First, it passes a list of supported overlays
// in the connection features, and uses that to determine which
// overlays are in common.  Then it tries those common overlays, and
// uses the best one that seems to be working.

type OverlaySwitch struct {
	overlays      map[string]NetworkOverlay
	overlayNames  []string
	compatOverlay NetworkOverlay
}

func NewOverlaySwitch() *OverlaySwitch {
	return &OverlaySwitch{overlays: make(map[string]NetworkOverlay)}
}

func (osw *OverlaySwitch) Add(name string, overlay NetworkOverlay) {
	// check for repeated names
	if _, present := osw.overlays[name]; present {
		log.Fatal("OverlaySwitch: repeated overlay name")
	}

	osw.overlays[name] = overlay
	osw.overlayNames = append(osw.overlayNames, name)
}

func (osw *OverlaySwitch) SetCompatOverlay(overlay NetworkOverlay) {
	osw.compatOverlay = overlay
}

func (osw *OverlaySwitch) AddFeaturesTo(features map[string]string) {
	features["Overlays"] = strings.Join(osw.overlayNames, " ")
}

func (osw *OverlaySwitch) Diagnostics() interface{} {
	diagnostics := make(map[string]interface{})
	for name, overlay := range osw.overlays {
		diagnostics[name] = overlay.Diagnostics()
	}
	return diagnostics
}

func (osw *OverlaySwitch) Stop() {
	for _, overlay := range osw.overlays {
		overlay.Stop()
	}
}

func (osw *OverlaySwitch) InvalidateRoutes() {
	for _, overlay := range osw.overlays {
		overlay.InvalidateRoutes()
	}
}

func (osw *OverlaySwitch) InvalidateShortIDs() {
	for _, overlay := range osw.overlays {
		overlay.InvalidateShortIDs()
	}
}

func (osw *OverlaySwitch) StartConsumingPackets(localPeer *mesh.Peer, peers *mesh.Peers, consumer OverlayConsumer) error {
	for _, overlay := range osw.overlays {
		if err := overlay.StartConsumingPackets(localPeer, peers, consumer); err != nil {
			return err
		}
	}
	return nil
}

type namedOverlay struct {
	NetworkOverlay
	name string
}

// Find the common set of overlays supported by both sides, with the
// ordering being the same on both sides too.
func (osw *OverlaySwitch) commonOverlays(params mesh.OverlayConnectionParams) ([]namedOverlay, error) {
	var peerOverlays []string
	if overlaysFeature, present := params.Features["Overlays"]; present {
		peerOverlays = strings.Split(overlaysFeature, " ")
	}

	common := make(map[string]NetworkOverlay)
	for _, name := range peerOverlays {
		if overlay := osw.overlays[name]; overlay != nil {
			common[name] = overlay
		}
	}

	if len(common) == 0 {
		return nil, fmt.Errorf("no overlays in common with peer")
	}

	// we order them according to the connecting node
	ordering := osw.overlayNames
	if params.RemoteAddr == nil {
		// we are the connectee
		ordering = peerOverlays
	}

	res := make([]namedOverlay, 0, len(common))
	for _, name := range ordering {
		overlay := common[name]
		if overlay != nil {
			res = append(res, namedOverlay{overlay, name})
		}
	}

	// we use bytes to represent forwarder indices in control
	// messages, so just in case:
	if len(res) > 256 {
		res = res[:256]
	}

	return res, nil
}

type overlaySwitchForwarder struct {
	remotePeer *mesh.Peer

	lock sync.Mutex

	// the index of the forwarder to send on
	best int

	// the subsidiary forwarders
	forwarders []subForwarder

	// closed to tell the main goroutine to stop
	stopChan chan<- struct{}

	alreadyEstablished bool
	establishedChan    chan struct{}
	errorChan          chan error
}

// A subsidiary forwarder
type subForwarder struct {
	fwd         OverlayForwarder
	overlayName string

	// Has the forwarder signalled that it is established?
	established bool

	// closed to tell the forwarder monitor goroutine to stop
	stopChan chan<- struct{}
}

// An event from a subsidiary forwarder
type subForwarderEvent struct {
	// the index of the forwarder
	index int

	// is this an "established" event?
	established bool

	// is this an error event?
	err error
}

func (osw *OverlaySwitch) PrepareConnection(params mesh.OverlayConnectionParams) (mesh.OverlayConnection, error) {
	if _, present := params.Features["Overlays"]; !present && osw.compatOverlay != nil {
		return osw.compatOverlay.PrepareConnection(params)
	}

	overlays, err := osw.commonOverlays(params)
	if err != nil {
		return nil, err
	}

	// channel to carry events from the subforwarder monitors to
	// the main goroutine
	eventsChan := make(chan subForwarderEvent)

	// channel to stop the main goroutine
	stopChan := make(chan struct{})

	fwd := &overlaySwitchForwarder{
		remotePeer: params.RemotePeer,

		best:       -1,
		forwarders: make([]subForwarder, len(overlays)),
		stopChan:   stopChan,

		establishedChan: make(chan struct{}),
		errorChan:       make(chan error, 1),
	}

	origSendControlMessage := params.SendControlMessage
	for i, overlay := range overlays {
		// Prefix control messages to indicate the relevant forwarder
		index := i
		params.SendControlMessage = func(tag byte, msg []byte) error {
			xmsg := make([]byte, len(msg)+2)
			xmsg[0] = byte(index)
			xmsg[1] = tag
			copy(xmsg[2:], msg)
			return origSendControlMessage(mesh.ProtocolOverlayControlMsg, xmsg)
		}

		subConn, err := overlay.PrepareConnection(params)
		if err != nil {
			log.Infof("Unable to use %s for connection to %s(%s): %s",
				overlay.name,
				params.RemotePeer.Name,
				params.RemotePeer.NickName,
				err)
			// failed to start subforwarder - record overlay name and continue
			fwd.forwarders[i] = subForwarder{
				overlayName: overlay.name,
			}
			continue
		}
		subFwd := subConn.(OverlayForwarder)

		subStopChan := make(chan struct{})
		go monitorForwarder(i, eventsChan, subStopChan, subFwd)
		fwd.forwarders[i] = subForwarder{
			fwd:         subFwd,
			overlayName: overlay.name,
			stopChan:    subStopChan,
		}
	}

	fwd.chooseBest()
	go fwd.run(eventsChan, stopChan)
	return fwd, nil
}

func monitorForwarder(index int, eventsChan chan<- subForwarderEvent, stopChan <-chan struct{}, fwd OverlayForwarder) {
	establishedChan := fwd.EstablishedChannel()
loop:
	for {
		e := subForwarderEvent{index: index}

		select {
		case <-establishedChan:
			e.established = true
			establishedChan = nil

		case err := <-fwd.ErrorChannel():
			e.err = err

		case <-stopChan:
			break loop
		}

		select {
		case eventsChan <- e:
		case <-stopChan:
			break loop
		}

		if e.err != nil {
			break loop
		}
	}

	fwd.Stop()
}

func (fwd *overlaySwitchForwarder) run(eventsChan <-chan subForwarderEvent, stopChan <-chan struct{}) {
loop:
	for {
		select {
		case <-stopChan:
			break loop

		case e := <-eventsChan:
			switch {
			case e.established:
				fwd.established(e.index)
			case e.err != nil:
				fwd.error(e.index, e.err)
			}
		}
	}

	fwd.lock.Lock()
	defer fwd.lock.Unlock()
	fwd.stopFrom(0)
}

func (fwd *overlaySwitchForwarder) established(index int) {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()

	fwd.forwarders[index].established = true

	if !fwd.alreadyEstablished {
		fwd.alreadyEstablished = true
		close(fwd.establishedChan)
	}

	fwd.chooseBest()
}

func (fwd *overlaySwitchForwarder) logPrefix() string {
	return fmt.Sprintf("overlay_switch ->[%s] ", fwd.remotePeer)
}

func (fwd *overlaySwitchForwarder) error(index int, err error) {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()

	log.Info(fwd.logPrefix(), fwd.forwarders[index].overlayName, " ", err)
	fwd.forwarders[index].fwd = nil
	fwd.chooseBest()
}

func (fwd *overlaySwitchForwarder) stopFrom(index int) {
	for index < len(fwd.forwarders) {
		subFwd := &fwd.forwarders[index]
		if subFwd.fwd != nil {
			subFwd.fwd = nil
			close(subFwd.stopChan)
		}
		index++
	}
}

func (fwd *overlaySwitchForwarder) chooseBest() {
	// the most preferred established forwarder is the best
	// otherwise, the most preferred working forwarder is the best
	bestEstablished := -1
	bestWorking := -1

	for i := range fwd.forwarders {
		subFwd := &fwd.forwarders[i]
		if subFwd.fwd == nil {
			continue
		}

		if bestWorking < 0 {
			bestWorking = i
		}

		if bestEstablished < 0 && subFwd.established {
			bestEstablished = i
		}
	}

	best := bestEstablished
	if best < 0 {
		if bestWorking < 0 {
			select {
			case fwd.errorChan <- fmt.Errorf("no working forwarders to %s", fwd.remotePeer):
			default:
			}

			return
		}

		best = bestWorking
	}

	if fwd.best != best {
		fwd.best = best
		log.Info(fwd.logPrefix(), "using ", fwd.forwarders[best].overlayName)
	}
}

func (fwd *overlaySwitchForwarder) Confirm() {
	var forwarders []OverlayForwarder

	fwd.lock.Lock()
	for _, subFwd := range fwd.forwarders {
		if subFwd.fwd != nil {
			forwarders = append(forwarders, subFwd.fwd)
		}
	}
	fwd.lock.Unlock()

	for _, subFwd := range forwarders {
		subFwd.Confirm()
	}
}

func (fwd *overlaySwitchForwarder) Forward(pk ForwardPacketKey) FlowOp {
	fwd.lock.Lock()

	if fwd.best >= 0 {
		for i := fwd.best; i < len(fwd.forwarders); i++ {
			best := fwd.forwarders[i].fwd
			if best != nil {
				fwd.lock.Unlock()
				if op := best.Forward(pk); op != nil {
					return op
				}
				fwd.lock.Lock()
			}
		}
	}

	fwd.lock.Unlock()

	return DiscardingFlowOp{}
}

func (fwd *overlaySwitchForwarder) EstablishedChannel() <-chan struct{} {
	return fwd.establishedChan
}

func (fwd *overlaySwitchForwarder) ErrorChannel() <-chan error {
	return fwd.errorChan
}

func (fwd *overlaySwitchForwarder) Stop() {
	fwd.lock.Lock()
	defer fwd.lock.Unlock()
	fwd.stopFrom(0)
}

func (fwd *overlaySwitchForwarder) ControlMessage(tag byte, msg []byte) {
	fwd.lock.Lock()
	subFwd := fwd.forwarders[msg[0]].fwd
	fwd.lock.Unlock()
	if subFwd != nil {
		subFwd.ControlMessage(msg[1], msg[2:])
	}
}

func (fwd *overlaySwitchForwarder) Attrs() map[string]interface{} {
	var best OverlayForwarder

	fwd.lock.Lock()
	if fwd.best >= 0 {
		best = fwd.forwarders[fwd.best].fwd
	}
	fwd.lock.Unlock()

	if best != nil {
		return best.Attrs()
	}

	return nil
}
