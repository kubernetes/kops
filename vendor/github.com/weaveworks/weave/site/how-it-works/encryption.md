---
title: Encryption and Weave Net
menu_order: 60
---


Weave Net peers
[communicate with each other over both TCP and UDP](/site/how-it-works.md). This
communication
[can be encrypted](/site/using-weave/security-untrusted-networks.md).

Encryption of control plane traffic (TCP) and data plane traffic (UDP) of sleeve
overlay is accomplished using the [NaCl](http://nacl.cr.yp.to/)
crypto libraries, employing Curve25519, XSalsa20 and Poly1305 to
encrypt and authenticate messages. Weave Net protects against
injection and replay attacks for traffic forwarded between peers.

NaCl was selected because of its good reputation both in terms of
selection and implementation of ciphers, but equally importantly, its
clear APIs, good documentation and high-quality
[go implementation](https://godoc.org/golang.org/x/crypto/nacl). It is
quite difficult to use NaCl incorrectly. Contrast this with libraries
such as OpenSSL where the library and its APIs are vast in size,
poorly documented, and easily used wrongly.

There are some similarities between Weave Net's crypto and
[TLS](https://tools.ietf.org/html/rfc4346). Weave Net does not need to cater
for multiple cipher suites, certificate exchange and other
requirements emanating from X509, and a number of other features. This
simplifies the protocol and implementation considerably. On the other
hand, Weave Net needs to support UDP transports, and while there are
extensions to TLS such as [DTLS](https://tools.ietf.org/html/rfc4347)
which can operate over UDP, these are not widely implemented and
deployed.

In the case of fast datapath, data plane traffic is encrypted by using
[ESP of IPsec](https://tools.ietf.org/html/rfc2406).
The process of encryption is handled by the Linux kernel and is controlled via
the IP transformation framework (XFRM).

**See Also**

 * [How Weave Implements Encryption](/site/how-it-works/encryption-implementation.md)
 * [Securing Containers Across Untrusted Networks](/site/using-weave/security-untrusted-networks.md)
