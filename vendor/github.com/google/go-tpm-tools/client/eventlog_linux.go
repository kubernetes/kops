package client

import (
	"io/ioutil"
)

func getRealEventLog() ([]byte, error) {
	return ioutil.ReadFile("/sys/kernel/security/tpm0/binary_bios_measurements")
}
