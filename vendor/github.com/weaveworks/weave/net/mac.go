package net

import (
	"crypto/rand"
	"crypto/sha256"
	"net"
)

func RandomMAC() (net.HardwareAddr, error) {
	mac := make([]byte, 6)
	if _, err := rand.Read(mac); err != nil {
		return nil, err
	}

	setUnicastAndLocal(mac)

	return net.HardwareAddr(mac), nil
}

func MACfromUUID(uuid []byte) net.HardwareAddr {
	hash := sha256.New()
	// We salt the input just as a precaution to avoid clashes with
	// other applications who might have had the bright idea of
	// generating MACs in the same way.
	hash.Write([]byte("9oBJ0Jmip-"))
	hash.Write(uuid)
	sum := hash.Sum(nil)

	setUnicastAndLocal(sum)

	return net.HardwareAddr(sum[:6])
}

// In the first byte of the MAC, the 'multicast' bit should be
// clear and 'locally administered' bit should be set.
func setUnicastAndLocal(mac []byte) {
	mac[0] = (mac[0] & 0xFE) | 0x02
}
