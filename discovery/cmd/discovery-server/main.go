/*
Copyright 2025 The Kubernetes Authors.

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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"k8s.io/kops/discovery/pkg/discovery"
)

func main() {
	certFile := os.Getenv("TLS_CERT")
	flag.StringVar(&certFile, "tls-cert", certFile, "Path to server TLS certificate")

	keyFile := os.Getenv("TLS_KEY")
	flag.StringVar(&keyFile, "tls-key", keyFile, "Path to server TLS key")

	addr := flag.String("listen", ":8443", "Address to listen on")
	storageType := flag.String("storage", "memory", "Storage backend (memory, gcs)")
	flag.Parse()

	if certFile == "" || keyFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --tls-cert and --tls-key are required\n")
		flag.Usage()
		os.Exit(1)
	}

	var store discovery.Store

	switch *storageType {
	case "memory":
		store = discovery.NewMemoryStore()
	default:
		log.Fatalf("Unknown storage type: %s", *storageType)
	}

	handler := discovery.NewServer(store)

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequestClientCert,
		// We do not set ClientCAs because we accept any CA and use it to define the universe.
		MinVersion: tls.VersionTLS12,
	}

	server := &http.Server{
		Addr:      *addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	log.Printf("Discovery server listening on %s using %s storage", *addr, *storageType)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
