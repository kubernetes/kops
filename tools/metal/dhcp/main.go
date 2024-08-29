/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	// this function will just print the received DHCPv4 message, without replying
	if m == nil {
		log.Printf("Packet is nil!")
		return
	}

	fmt.Printf("got dhcp packet: messageType=%v, mac=%v\n", m.MessageType(), m.ClientHWAddr)

	if m.OpCode != dhcpv4.OpcodeBootRequest {
		log.Printf("Not a BootRequest!")
		return
	}
	modifiers := []dhcpv4.Modifier{}

	switch mt := m.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		modifiers = append(modifiers, dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer))
		// reply.UpdateOption(dhcpv4.OptServerIdentifier(net.IP{10, 20, 30, 1}))
		// reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
	case dhcpv4.MessageTypeRequest:
		modifiers = append(modifiers, dhcpv4.WithMessageType(dhcpv4.MessageTypeAck))
		// reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	default:
		log.Printf("Unhandled message type: %v", mt)
		return
	}

	if len(m.ClientHWAddr) != 6 {
		log.Printf("unexpected mac address %v (expected length 6)", m.ClientHWAddr)
		return
	}

	clientIP := net.IP{10, 123, 45, m.ClientHWAddr[5]}
	serverIP := net.IP{10, 123, 45, 1}

	router := serverIP

	modifiers = append(modifiers, dhcpv4.WithYourIP(clientIP))
	modifiers = append(modifiers, dhcpv4.WithDNS(net.IP{8, 8, 8, 8}))
	modifiers = append(modifiers, dhcpv4.WithNetmask(net.IPMask{255, 255, 255, 0}))
	modifiers = append(modifiers, dhcpv4.WithRouter(router))
	// modifiers = append(modifiers, dhcpv4.WithGatewayIP(net.IP{10, 20, 30, 1}))
	modifiers = append(modifiers, dhcpv4.WithLeaseTime(60*60*4))
	modifiers = append(modifiers, dhcpv4.WithServerIP(serverIP))
	modifiers = append(modifiers, dhcpv4.WithOption(dhcpv4.OptServerIdentifier(serverIP)))

	reply, err := dhcpv4.NewReplyFromRequest(m, modifiers...)
	if err != nil {
		log.Printf("NewReplyFromRequest failed: %v", err)
		return
	}
	// reply.UpdateOption(dhcpv4.OptServerIdentifier(net.IP{10, 20, 30, 1}))
	// reply.UpdateOption(dhcpv4.OptClientIdentifier(net.IP{10, 20, 30, 4}))

	if _, err := conn.WriteTo(reply.ToBytes(), peer); err != nil {
		log.Printf("Cannot reply to client: %v", err)
	}
}

func run(ctx context.Context) error {
	// listenIP := "10.123.45.1"
	listenIP := "0.0.0.0"
	// listenInterface := "loop-br0"
	listenInterface := "br0"

	ip := net.ParseIP(listenIP)
	if ip == nil {
		return fmt.Errorf("unable to parse IP %q", listenIP)
	}
	laddr := net.UDPAddr{
		IP:   ip,
		Port: 67,
	}
	fmt.Printf("listening on %v\n", listenInterface)
	server, err := server4.NewServer(listenInterface, &laddr, handler)
	if err != nil {
		log.Fatal(err)
	}

	// This never returns. If you want to do other stuff, dump it into a
	// goroutine.
	server.Serve()

	return nil
}
