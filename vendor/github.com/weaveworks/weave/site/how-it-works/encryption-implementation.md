---
title: How Weave Net Implements Encryption
menu_order: 70
---

This section describes some details of Weave Net's built-in
[encryption](/site/how-it-works/encryption.md):

 * [Establishing the Ephemeral Session Key](#ephemeral-key)
 * [Key Generation](#csprng)
 * [Encypting and Decrypting TCP Messages](#tcp)
 * [Encypting and Decrypting UDP Messages](#udp)



#### <a name="ephemeral-key"></a>Establishing the Ephemeral Session Key

For every connection between peers, a fresh public/private key pair is
created at both ends, using NaCl's `GenerateKey` function. The public
key portion is sent to the other end as part of the initial handshake
performed over TCP. Peers that were started with a password do not
continue with connection establishment unless they receive a public
key from the remote peer. Thus either all peers in a weave network
must be supplied with a password, or none.

When a peer has received a public key from the remote peer, it uses
this to form the ephemeral session key for this connection. The public
key from the remote peer is combined with the private key for the
local peer in the usual [Diffie-Hellman way](https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange), 
resulting in both peers arriving at the same shared key. To this is appended the supplied
password, and the result is hashed through SHA256, to form the final
ephemeral session key. 

The supplied password is never exchanged directly, and is thoroughly 
mixed into the shared secret. Furthermore, the rate at which TCP connections 
are accepted is limited by Weave to 10Hz, which thwarts online 
dictionary attacks on reasonably strong passwords.

The shared key formed by Diffie-Hellman is 256 bits long. Appending
the password to this obviously makes it longer by an unknown amount,
and the use of SHA256 reduces this back to 256 bits, to form the final
ephemeral session key. This late combination with the password
eliminates any "Man In The Middle" attacks: sniffing the public key
exchange between the two peers and faking their responses will not
grant an attacker knowledge of the password, and therefore, an attacker would
not be able to form valid ephemeral session keys.

The same ephemeral session key is used for both TCP and UDP traffic
between two peers.

### <a name="csprng"></a> Key Generation and The Linux CSPRNG

Generating fresh keys for every connection
provides forward secrecy at the cost of placing a demand on the Linux
CSPRNG (accessed by `GenerateKey` via `/dev/urandom`) proportional to
the number of inbound connection attempts. Weave Net has accept throttling
to mitigate against denial of service attacks that seek to deplete the
CSPRNG entropy pool, however even at the lower bound of ten requests
per second, there may not be enough entropy gathered on a headless
system to keep pace.

Under such conditions, the consequences will be limited to slowing
down processes reading from the blocking `/dev/random` device as the
kernel waits for enough new entropy to be harvested. It is important
to note that contrary to intuition, this low entropy state does not
compromise the ongoing use of `/dev/urandom`. [Expert
opinion](http://blog.cr.yp.to/20140205-entropy.html)
asserts that as long as the CSPRNG is seeded with enough entropy (for example,
256 bits) before random number generation commences, then the output is
entirely safe for use as key material.

By way of comparison, this is exactly how OpenSSL works - it reads 256
bits of entropy at startup, and uses that to seed an internal CSPRNG,
which is used to generate keys. While Weave Net could have taken
the same approach and built a custom CSPRNG to work around the
potential `/dev/random` blocking issue, the decision was made to rely
on the [heavily scrutinized](http://eprint.iacr.org/2012/251.pdf) Linux random number
generator as [advised
here](http://cr.yp.to/highspeed/coolnacl-20120725.pdf) (page 10,
'Centralizing randomness'). 

>**Note:**The aforementioned notwithstanding, if
Weave Net's demand on `/dev/urandom` is causing you problems with blocking
`/dev/random` reads, please get in touch with us - we'd love to hear
about your use case.

#### <a name="tcp"></a>Encypting and Decrypting TCP Messages

TCP connection are only used to exchange topology information between
peers, via a message-based protocol. Encryption of each message is
carried out by NaCl's `secretbox.Seal` function using the ephemeral
session key and a nonce. The nonce contains the message sequence
number, which is incremented for every message sent, and a bit
indicating the polarity of the connection at the sender ('1' for
outbound). The latter is required by the
[NaCl Security Model](http://nacl.cr.yp.to/box.html) in order to
ensure that the two ends of the connection do not use the same nonces.

Decryption of a message at the receiver is carried out by NaCl's
`secretbox.Open` function using the ephemeral session key and a
nonce. The receiver maintains its own message sequence number, which
it increments for every message it decrypted successfully. The nonce
is constructed from that sequence number and the connection
polarity. As a result the receiver will only be able to decrypt a
message if it has the expected sequence number. This prevents replay
attacks.

#### <a name="udp"></a>Encrypting and Decrypting UDP Packets

##### Sleeve

UDP connections carry captured traffic between peers. For a UDP packet
sent between peers that are using crypto, the encapsulation looks as
follows:

    +-----------------------------------+
    | Name of sending peer              |
    +-----------------------------------+
    | Message Sequence No and flags     |
    +-----------------------------------+
    | NaCl SecretBox overheads          |
    +-----------------------------------+ -+
    | Frame 1: Name of capturing peer   |  |
    +-----------------------------------+  | This section is encrypted
    | Frame 1: Name of destination peer |  | using the ephemeral session
    +-----------------------------------+  | key between the weave peers
    | Frame 1: Captured payload length  |  | sending and receiving this
    +-----------------------------------+  | packet.
    | Frame 1: Captured payload         |  |
    +-----------------------------------+  |
    | Frame 2: Name of capturing peer   |  |
    +-----------------------------------+  |
    | Frame 2: Name of destination peer |  |
    +-----------------------------------+  |
    | Frame 2: Captured payload length  |  |
    +-----------------------------------+  |
    | Frame 2: Captured payload         |  |
    +-----------------------------------+  |
    |                ...                |  |
    +-----------------------------------+  |
    | Frame N: Name of capturing peer   |  |
    +-----------------------------------+  |
    | Frame N: Name of destination peer |  |
    +-----------------------------------+  |
    | Frame N: Captured payload length  |  |
    +-----------------------------------+  |
    | Frame N: Captured payload         |  |
    +-----------------------------------+ -+

This is very similar to the [non-crypto encapsulation](/site/how-it-works/router-encapsulation.md).

All of the frames on a connection are encrypted with the same
ephemeral session key, and a nonce constructed from a message sequence
number, flags and the connection polarity. This is very similar to the
TCP encryption scheme, and encryption is again done with the NaCl
`secretbox.Seal` function. The main difference is that the message
sequence number and flags are transmitted as part of the message,
unencrypted.

The receiver uses the name of the sending peer to determine which
ephemeral session key and local cryptographic state to use for
decryption. Frames which are to be forwarded on to some further peer
will be re-encrypted with the relevant ephemeral session keys for the
onward connections. Thus all traffic is fully decrypted on every peer
it passes through.

Decryption is once again carried out by NaCl's `secretbox.Open`
function using the ephemeral session key and nonce. The latter is
constructed from the message sequence number and flags that appeared
in the unencrypted portion of the received message, and the connection
polarity.

To guard against replay attacks, the receiver maintains some state in
which it remembers the highest message sequence number seen. It could
simply reject messages with lower sequence numbers, but that could
result in excessive message loss when messages are re-ordered. The
receiver therefore additionally maintains a set of received message
sequence numbers in a window below the highest number seen, and only
rejects messages with a sequence number below that window, or
contained in the set. The window spans at least 2^20 message sequence
numbers, and hence any re-ordering between the most recent ~1 million
messages is handled without dropping messages.

##### Fast Datapath

Encryption in fastdp uses [the ESP protocol of IPsec](https://tools.ietf.org/html/rfc2406)
in the transport mode. Each VXLAN packet is encrypted with
[AES in GCM mode](https://tools.ietf.org/html/rfc4106aesgcm), with 32 byte key and
4 byte salt. This combo provides the following security properties:

* Data confidentiality.
* Data origin authentication.
* Integrity.
* Anti-replay.
* Limited traffic flow confidentiality as VXLAN packets are fully encrypted.

For each connection direction, a different AES-GCM key and salt is used.
The pairs are derived with [HKDF](https://tools.ietf.org/html/rfc5869)
to which we pass a randomly generated 32 byte salt transferred over the encrypted
control plane channel between peers.

To prevent from replay attacks, which are possible because of the size of
sequence number field in ESP (4 bytes), we use extended sequence numbers
implemented by [ESN](https://tools.ietf.org/html/rfc4304).

Authentication of ESP packet integrity and origin is ensured by 16 byte
Integrity Check Value of AES-GCM.

**See Also**

 * [architecture documentation](https://github.com/weaveworks/weave/blob/master/docs/architecture.txt)
 * [fastdp encryption](https://github.com/weaveworks/weave/blob/master/docs/fastdp-crypto.md)
 * [Securing Containers Across Untrusted Networks](/site/using-weave/security-untrusted-networks.md)
