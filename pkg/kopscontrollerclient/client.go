/*
Copyright 2020 The Kubernetes Authors.

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

package kopscontrollerclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcresolver "google.golang.org/grpc/resolver"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	pb "k8s.io/kops/proto/generated/kops/kopscontroller/v1"
)

type Client struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator bootstrap.Authenticator
	// CAs are the CA certificates for kops-controller.
	CAs []byte

	// BaseURL is the base URL for the server
	BaseURL url.URL

	// Resolver is a custom resolver that supports resolution of hostnames without requiring DNS.
	// In particular, this supports gossip mode.
	Resolver resolver.Resolver

	httpClient *http.Client
}

// dial implements a DialContext resolver function, for when a custom resolver is in use
func (b *Client) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	var errors []error

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot split host and port from %q: %w", addr, err)
	}

	// TODO: cache?
	addresses, err := b.Resolver.Resolve(ctx, host)
	if err != nil {
		return nil, err
	}

	klog.Infof("resolved %q to %v", host, addresses)

	for _, addr := range addresses {
		timeout := 5 * time.Second
		conn, err := net.DialTimeout(network, addr+":"+port, timeout)
		if err == nil {
			return conn, nil
		}
		if err != nil {
			klog.Warningf("failed to dial %q: %v", addr, err)
			errors = append(errors, err)
		}
	}
	if len(errors) == 0 {
		return nil, fmt.Errorf("no addresses for %q", addr)
	}
	return nil, errors[0]
}

func (b *Client) buildTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(b.CACertificates)
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}
	tlsConfig.Certificates = b.ClientCertificates
	return tlsConfig
}

func (b *Client) getHTTPClient(ctx context.Context) (*http.Client, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.cachedHTTPClient == nil {
		tlsConfig := b.buildTLSConfig()

		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		if b.Resolver != nil {
			transport.DialContext = b.dial
		}

		httpClient := &http.Client{
			Timeout:   time.Duration(15) * time.Second,
			Transport: transport,
		}

		b.httpClient = httpClient
	}

	// Sanity-check DNS to provide clearer diagnostic messages.
	if b.Resolver != nil {
		// Don't check DNS when there's a custom resolver.
	} else if ips, err := net.LookupIP(b.BaseURL.Hostname()); err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (not found: %v)", dnsErr))
		}
		return err
	} else if len(ips) == 1 && (ips[0].String() == cloudup.PlaceholderIP || ips[0].String() == cloudup.PlaceholderIPv6) {
		return fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (placeholder IP found: %v)", ips))
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bootstrapURL := b.BaseURL
	bootstrapURL.Path = path.Join(bootstrapURL.Path, "/bootstrap")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := b.Authenticator.CreateToken(reqBytes)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", token)

	response, err := b.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	if response.StatusCode != http.StatusOK {
		detail := ""
		if response.Body != nil {
			scanner := bufio.NewScanner(response.Body)
			if scanner.Scan() {
				detail = scanner.Text()
			}
		}
		return fmt.Errorf("kops-controller returned status code %d: %s", response.StatusCode, detail)
	}

	return json.NewDecoder(response.Body).Decode(resp)
}

func (b *Client) DiscoverHosts(ctx context.Context, req *pb.DiscoverHostsRequest) (pb.KopsControllerService_DiscoverHostsClient, error) {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(b.CAs)
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	grpcURL := b.BaseURL.String()

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var opts []grpc.DialOption

	if b.Resolver != nil {
		resolverBuilder := &grpcResolverBuilder{kopsResolver: b.Resolver}
		if !strings.HasPrefix(grpcURL, "https://") {
			return nil, fmt.Errorf("expected kops-controller url to have https:// scheme, was %q", grpcURL)
		}
		grpcURL = strings.Replace(grpcURL, "https://", "kops://", 1)
		opts = append(opts, grpc.WithResolvers(resolverBuilder))
	}

	tlsTransportCredentials := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(tlsTransportCredentials))

	rpcCredentials := &grpcPerRPCCredentials{
		Authenticator: b.Authenticator,
	}
	opts = append(opts, grpc.WithPerRPCCredentials(rpcCredentials))

	conn, err := grpc.DialContext(dialCtx, grpcURL, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pb.NewKopsControllerServiceClient(conn)

	stream, err := client.DiscoverHosts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error from DiscoverHosts request to kops-controller: %w", err)
	}
	return stream, nil
}

type grpcPerRPCCredentials struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator bootstrap.Authenticator
}

func (c grpcPerRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	ri, _ := credentials.RequestInfoFromContext(ctx)
	if err := credentials.CheckSecurityLevel(ri.AuthInfo, credentials.PrivacyAndIntegrity); err != nil {
		return nil, fmt.Errorf("unable to transfer grpcPerRPCCredentials PerRPCCredentials: %w", err)
	}

	uid := string(uuid.NewUUID())

	token, err := c.Authenticator.CreateToken([]byte(uid))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"uid":           uid,
		"authorization": token,
	}, nil
}

func (c grpcPerRPCCredentials) RequireTransportSecurity() bool {
	return true
}

type grpcResolverBuilder struct {
	kopsResolver resolver.Resolver
}

var _ grpcresolver.Builder = &grpcResolverBuilder{}

// Scheme implements grpcresolver.Builder
func (r *grpcResolverBuilder) Scheme() string {
	return "kops"
}

// Build implements grpcresolver.Builder
func (r *grpcResolverBuilder) Build(target grpcresolver.Target, clientConn grpcresolver.ClientConn, opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	return &grpcResolver{
		kopsResolver: r.kopsResolver,
		clientConn:   clientConn,
		url:          target.URL,
	}, nil
}

type grpcResolver struct {
	kopsResolver resolver.Resolver
	url          url.URL

	mutex      sync.Mutex
	clientConn grpcresolver.ClientConn
}

var _ grpcresolver.Resolver = &grpcResolver{}

// ResolveNow implements grpcresolver.Resolver
func (r *grpcResolver) ResolveNow(opt grpcresolver.ResolveNowOptions) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ctx := context.TODO()

	addresses, err := r.resolveAddresses(ctx)
	if err != nil {
		r.clientConn.ReportError(err)
	} else {
		r.clientConn.UpdateState(grpcresolver.State{
			Addresses: addresses,
		})
	}
}

func (r *grpcResolver) resolveAddresses(ctx context.Context) ([]grpcresolver.Address, error) {
	host, _, err := net.SplitHostPort(r.url.Host)
	if err != nil {
		return nil, fmt.Errorf("cannot split host and port from %q: %w", r.url.Host, err)
	}

	// TODO: cache?
	addresses, err := r.kopsResolver.Resolve(ctx, host)
	if err != nil {
		return nil, err
	}

	klog.Infof("resolved %q to %v", host, addresses)

	var grpcAddresses []grpcresolver.Address
	for _, address := range addresses {
		grpcAddresses = append(grpcAddresses, grpcresolver.Address{
			Addr: address,
		})
	}

	return grpcAddresses, nil
}

// Close implements grpcresolver.Resolver
func (r *grpcResolver) Close() {
}
