package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"k8s.io/kops/tools/triage/testsoup/pkg/testdata"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	klog.InitFlags(nil)

	listen := ":9999"
	flag.StringVar(&listen, "listen", listen, "endpoint on which to listen")
	flag.Parse()

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %w", listen, err)
	}

	klog.Infof("listening on %q", listen)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := &testdata.Server{}
	testdata.RegisterTestDataServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("error serving GRPC requests: %w", err)
	}

	return nil
}
