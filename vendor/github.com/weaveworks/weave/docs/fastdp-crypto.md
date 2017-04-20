This document describes implementation details of the fast datapath encryption.

# Overview

At the high level, we use the ESP protocol ([RFC 2406][esp]) in the Transport
mode. Each packet is encrypted with AES in GCM mode ([RFC 4106][aesgcm]), with
32 byte key and 4 byte salt. This combo provides the following security
properties:

* Data confidentiality.
* Data origin authentication.
* Integrity.
* Anti-replay.
* Limited traffic flow confidentiality as fast datapath VXLAN packets are fully
  encrypted.

## Notation

* `SAin`:  IPsec security association for inbound connections.
* `SAout`: IPsec security association for outbound connections.
* `SPout`: IPsec security policy for outbound connections. Used to match
           outbound flows to SAout.

## SA Key Derivation

For each connection direction, a different AES-GCM key and salt is used.
The pairs are derived with HKDF ([RFC 5869][hkdf]):

```
SAin[KeyAndSalt] = HKDF(sha256, ikm=sessionKey, salt=nonceIn, info=localPeerName)
SAout[KeyAndSalt] = HKDF(sha256, ikm=sessionKey, salt=nonceOut, info=remotePeerName)
```

Where:

* the mutual `sessionKey` is derived by the [Mesh][mesh] library during
  the control plane connection establishment between the local peer and
  the remote peer.
* `nonceIn` and `nonceOut` are randomly generated 32byte nonces which
  are exchanged over the encrypted control plane channel.

## SPI

A directional secure connection between two peers is identified with SPI.

The kernel requires the pair of SPI and dst IP to be unique among security
associations. Thus, to avoid collisions, we generate outbound SPI on a remote
peer.

## Connection Establishment

```
Peer A                                                        Peer B
-----------------------------------------------------------------------------------------------------------------

fastdp.fwd.Confirm():
    install iptables blocking rules,                          fastdp.fwd.Confirm():
    nonce_BA = rand(),                                            install iptables blocking rules,
    {key,salt}_BA = hkdf(sessionKey, nonce_BA, A),                nonce_AB = rand(),
    spi_BA = allocspi(),                                          {key,salt}_AB = hkdf(sessionKey, nonce_AB, B),
    create SA_BA(B<-A, spi_BA, key_BA, salt_BA),                  spi_AB = allocspi(),
    send InitSARemote(spi_BA, nonce_BA). -->                      create SA_AB(A<-B, spi_AB, key_AB, salt_AB),
                                                              <-- send InitSARemote(spi_AB, nonce_AB).


                                                          --> recv InitSARemote(spi_BA, nonce_BA):
recv InitSARemote(spi_AB, nonce_AB): <--                          {key,salt}_BA = hkdf(sessionKey, nonce_BA, A),
    {key,salt}_AB = hkdf(sessionKey, nonce_AB, B),                create SA_BA(B<-A, spi_BA, key_BA, salt_BA),
    create SA_AB(A<-B, spi_AB, key_AB, salt_AB),                  create SP_BA(B<-A, spi_BA),
    create SP_AB(A<-B, spi_AB),                                   install marking rule.
    install marking rule.
```

# Implementation Details

## XFRM

The implementation is based on the kernel IP packet transformation framework
called XFRM. Unfortunately, docs are barely existing and the complexity of
the framework is high. The best resource I found is Chapter 10 in
"Linux Kernel Networking: Implementation and Theory" by Rami Rosen.

## iptables Rules

The kernel VXLAN driver does not set a dst port of a tunnel in the ip flow
descriptor, thus XFRM policy lookup cannot match a policy (SPout) which includes
the port. This makes impossible to encrypt only tunneled traffic between
peers. To work around, we mark such outgoing packets with iptables and set
the same mark in the policy selector (funnily enough, `iptables_mangle` module
eventually sets the missing dst port in the flow descriptor). The challenge
here is to pick a mark that it would not interfere with other networking
applications before OUTPUT'ing a packet. For example, Kubernetes by default
uses 1<<14 and 1<<15 marks and we choose 1<<17 (0x20000). Additionally,
such workaround brings the requirement for at least the 4.2 kernel.

The marking rules are the following:

```
iptables -t mangle -A OUTPUT -j WEAVE-IPSEC-OUT
iptables -t mangle -A WEAVE-IPSEC-OUT -s ${LOCAL_PEER_IP} -d ${REMOTE_PEER_IP} \
         -p udp --dport ${TUNNEL_PORT} -j WEAVE-IPSEC-OUT-MARK
iptables -t mangle -A WEAVE-IPSEC-OUT-MARK --set-xmark ${MARK} -j MARK
```

As Linux does not implement [IP Security Levels][ipseclevel], we install
additional iptables rules to prevent from accidentally sending unencrypted
traffic between peers which have previously established the secure connection.

For inbound traffic, we mark each ESP packet with the mark and drop non-marked
tunnel traffic:

```
iptables -t mangle -A INPUT -j WEAVE-IPSEC-IN
iptables -t mangle -A WEAVE-IPSEC-IN -s ${REMOTE_PEER_IP} -d ${LOCAL_PEER_IP} \
         -m esp --espspi ${SPIin} -j WEAVE-IPSEC-IN-MARK
iptables -t mangle -A WEAVE-IPSEC-IN-MARK --set-xmark ${MARK} -j MARK

iptables -t filter -A INPUT -j WEAVE-IPSEC-IN
iptables -t filter -A WEAVE-IPSEC-IN -s ${REMOTE_PEER_IP} -d ${LOCAL_PEER_IP} \
         -p udp --dport ${TUNNEL_PORT} -m mark ! --mark ${MARK} -j DROP
```

For outbound traffic, we drop marked traffic which does not match any SPout:

```
iptables -t filter -A OUTPUT ! -p esp -m policy --dir out --pol none \
         -m mark --mark ${MARK} -j DROP
```

## ESN

To prevent from cycling SeqNo which makes replay attacks possible, we use
64-bit extended sequence numbers known as [ESN](esn).

## MTU

In addition to the VXLAN overhead, the MTU calculation should take into account
the ESP overhead which is 34-37 bytes (encrypted Payload is 4 bytes aligned) and
consists of:

* 4 bytes (SPI).
* 4 bytes (Sequence Number).
* 8 bytes (ESP IV).
* 1 byte (Pad Length).
* 1 byte (NextHeader).
* 16 bytes (ICV).
* 0-3 bytes (Padding).


[esp]:              https://tools.ietf.org/html/rfc2406
[aesgcm]:           https://tools.ietf.org/html/rfc4106
[hkdf]:             https://tools.ietf.org/html/rfc5869
[esn]:              https://tools.ietf.org/html/rfc4304
[ipseclevel]:       https://tools.ietf.org/html/draft-mcdonald-simple-ipsec-api-01)
[mesh]:             https://github.com/weaveworks/mesh
