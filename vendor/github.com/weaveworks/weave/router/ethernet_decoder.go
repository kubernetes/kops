package router

import (
	"bytes"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type EthernetDecoder struct {
	Eth     layers.Ethernet
	IP      layers.IPv4
	decoded []gopacket.LayerType
	parser  *gopacket.DecodingLayerParser
}

func NewEthernetDecoder() *EthernetDecoder {
	dec := &EthernetDecoder{}
	dec.parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &dec.Eth, &dec.IP)
	return dec
}

func (dec *EthernetDecoder) DecodeLayers(data []byte) {
	// We intentionally discard the error return here, because it
	// is normal for gopacket to return an error saying that it
	// cannot decode a layer beyond the ones we specified when
	// setting up the parser.
	dec.parser.DecodeLayers(data, &dec.decoded)
}

func (dec *EthernetDecoder) PacketKey() (key PacketKey) {
	copy(key.SrcMAC[:], dec.Eth.SrcMAC)
	copy(key.DstMAC[:], dec.Eth.DstMAC)
	return
}

func (dec *EthernetDecoder) makeICMPFragNeeded(mtu int) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true}
	ipHeaderSize := int(dec.IP.IHL) * 4 // IHL is the number of 32-byte words in the header
	payload := gopacket.Payload(dec.IP.BaseLayer.Contents[:ipHeaderSize+8])
	err := gopacket.SerializeLayers(buf, opts,
		&layers.Ethernet{
			SrcMAC:       dec.Eth.DstMAC,
			DstMAC:       dec.Eth.SrcMAC,
			EthernetType: dec.Eth.EthernetType},
		&layers.IPv4{
			Version:    4,
			TOS:        dec.IP.TOS,
			Id:         0,
			Flags:      0,
			FragOffset: 0,
			TTL:        64,
			Protocol:   layers.IPProtocolICMPv4,
			DstIP:      dec.IP.SrcIP,
			SrcIP:      dec.IP.DstIP},
		&layers.ICMPv4{
			TypeCode: 0x304,
			Id:       0,
			Seq:      uint16(mtu)},
		&payload)
	if err != nil {
		return nil, err
	}

	log.Printf("Sending ICMP 3,4 (%v -> %v): PMTU=%v", dec.IP.DstIP, dec.IP.SrcIP, mtu)
	return buf.Bytes(), nil
}

var (
	zeroMAC, _ = net.ParseMAC("00:00:00:00:00:00")
)

func (dec *EthernetDecoder) IsSpecial() bool {
	return dec.Eth.Length == 0 && dec.Eth.EthernetType == layers.EthernetTypeLLC &&
		bytes.Equal(zeroMAC, dec.Eth.SrcMAC) && bytes.Equal(zeroMAC, dec.Eth.DstMAC)
}

func (dec *EthernetDecoder) DF() bool {
	return len(dec.decoded) == 2 && (dec.IP.Flags&layers.IPv4DontFragment != 0)
}
