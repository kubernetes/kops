package main

import (
	"fmt"
	"log"
	"os"

	"github.com/docker/machine/libmachine/cert"
)

const (
	org            = "vagrant"
	bits           = 2048
	caKeyPath      = "ca-key.pem"
	caCertPath     = "ca.pem"
	clientKeyPath  = "key.pem"
	clientCertPath = "cert.pem"
	serverKeyPath  = "server-key.pem"
	serverCertPath = "server.pem"
)

func main() {
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Printf("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caKeyPath); err == nil {
			log.Fatalf("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := cert.GenerateCACertificate(caCertPath, caKeyPath, org, bits); err != nil {
			log.Printf("Error generating CA certificate: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Printf("Creating client certificate: %s", clientCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			log.Fatalf("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		err = cert.GenerateCert(
			[]string{""},
			clientCertPath,
			clientKeyPath,
			caCertPath,
			caKeyPath,
			org,
			bits,
		)
		if err != nil {
			log.Fatalf("Error generating client certificate: %s", err)
		}
	}

	if len(os.Args) <= 1 {
		return
	}
	for _, ip := range os.Args[1:] {
		serverKeyPath := fmt.Sprintf("%s-key.pem", ip)
		serverCertPath := fmt.Sprintf("%s.pem", ip)

		if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
			log.Printf("Creating server certificate: %s", serverCertPath)
			err = cert.GenerateCert(
				[]string{ip},
				serverCertPath,
				serverKeyPath,
				caCertPath,
				caKeyPath,
				org,
				bits,
			)
			if err != nil {
				log.Fatalf("error generating server cert: %s", err)
			}
		}
	}
}
