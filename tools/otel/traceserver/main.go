/*
Copyright 2023 The Kubernetes Authors.

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
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v11 "go.opentelemetry.io/proto/otlp/common/v1"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/klog/v2"
	v2 "k8s.io/kops/tools/otel/traceserver/pb/jaeger/api/v2"
	storagev1 "k8s.io/kops/tools/otel/traceserver/pb/jaeger/storage/v1"
	"k8s.io/kops/util/pkg/vfs"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	listen := "127.0.0.1:12345"
	run := ""
	src := ""
	flag.StringVar(&src, "src", src, "tracefile to load")
	flag.StringVar(&listen, "listen", listen, "endpoint on which to serve grpc")
	flag.StringVar(&run, "run", run, "visualization program to run [jaeger, docker-jaeger]")
	klog.InitFlags(nil)
	flag.Parse()

	if src == "" {
		return fmt.Errorf("--src is required")
	}

	if run != "" {
		switch run {
		case "jaeger", "docker-jaeger":
			go func() {
				opt := RunJaegerOptions{
					StorageServer: listen,
					UseDocker:     run == "docker-jaeger",
				}
				err := runJaeger(ctx, opt)
				if err != nil {
					klog.Warningf("error starting jaeger: %w", err)
				}
			}()
		default:
			return fmt.Errorf("run=%q not known (valid values: jaeger, docker-jaeger)", run)
		}
	}

	vfsContext := vfs.NewVFSContext()
	srcPath, err := vfsContext.BuildVfsPath(src)
	if err != nil {
		return fmt.Errorf("parsing path %q: %w", src, err)
	}

	klog.Infof("listing files under %v", srcPath)
	srcFiles, err := srcPath.ReadTree()
	if err != nil {
		return fmt.Errorf("reading tree %q: %w", srcFiles, err)
	}

	var traceFiles []*TraceFile

	for _, srcFile := range srcFiles {
		traceFile, err := ReadTraceFile(ctx, srcFile)
		if err != nil {
			return fmt.Errorf("reading %q: %w", srcFile, err)
		}
		traceFiles = append(traceFiles, traceFile)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listening on %q: %w", listen, err)
	}

	s := &Server{
		traceFiles: traceFiles,
	}
	grpcServer := grpc.NewServer()
	storagev1.RegisterPluginCapabilitiesServer(grpcServer, s)
	storagev1.RegisterSpanReaderPluginServer(grpcServer, s)
	storagev1.RegisterDependenciesReaderPluginServer(grpcServer, s)
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serving %q: %w", listen, err)
	}

	return nil
}

type Server struct {
	traceFiles []*TraceFile

	storagev1.UnimplementedPluginCapabilitiesServer
	storagev1.UnimplementedSpanReaderPluginServer
	storagev1.UnimplementedDependenciesReaderPluginServer
}

func (s *Server) Capabilities(ctx context.Context, req *storagev1.CapabilitiesRequest) (*storagev1.CapabilitiesResponse, error) {
	klog.V(2).Infof("Capabilities %v", prototext.Format(req))
	response := &storagev1.CapabilitiesResponse{
		ArchiveSpanReader:   false,
		ArchiveSpanWriter:   false,
		StreamingSpanWriter: false,
	}
	klog.V(4).Infof("<-Capabilities %v", prototext.Format(response))
	return response, nil
}

func (s *Server) GetTrace(req *storagev1.GetTraceRequest, stream storagev1.SpanReaderPlugin_GetTraceServer) error {
	ctx := stream.Context()

	klog.V(2).Infof("GetTrace %v", prototext.Format(req))

	var opt FilterOptions

	return s.visitSpans(ctx, opt, func(serviceName string, resourceSpans *v1.ResourceSpans) error {
		chunk := &storagev1.SpansResponseChunk{}
		for _, scopeSpan := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				if !bytes.Equal(span.TraceId, req.TraceId) {
					continue
				}

				out := convertToJaeger(serviceName, span)
				chunk.Spans = append(chunk.Spans, out)
			}
		}
		klog.V(4).Infof("<-GetTrace %v", prototext.Format(chunk))
		if err := stream.Send(chunk); err != nil {
			klog.Warningf("error sending chunk: %w", err)
			return err
		}

		return nil
	})
}
func (s *Server) GetServices(ctx context.Context, req *storagev1.GetServicesRequest) (*storagev1.GetServicesResponse, error) {
	klog.V(2).Infof("GetServices %v", prototext.Format(req))

	services := make(map[string]struct{})

	for _, traceFile := range s.traceFiles {
		for _, data := range traceFile.data {
			// klog.Infof("data %v", prototext.Format(data))
			for _, span := range data.ResourceSpans {
				if span.Resource != nil {
					for _, attr := range span.Resource.Attributes {
						if attr.GetKey() == "service.name" {
							serviceName := attr.GetValue().GetStringValue()
							if serviceName != "" {
								services[serviceName] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	response := &storagev1.GetServicesResponse{}

	for k := range services {
		response.Services = append(response.Services, k)
	}
	klog.V(4).Infof("<-GetServices %v", prototext.Format(response))
	return response, nil
}

func (s *Server) GetOperations(ctx context.Context, req *storagev1.GetOperationsRequest) (*storagev1.GetOperationsResponse, error) {
	klog.V(2).Infof("GetOperations %v", prototext.Format(req))

	var opt FilterOptions
	opt.ServiceName = req.GetService()

	operations := make(map[string]struct{})

	if err := s.visitSpans(ctx, opt, func(serviceName string, resourceSpans *v1.ResourceSpans) error {
		for _, scopeSpan := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				operations[span.Name] = struct{}{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	response := &storagev1.GetOperationsResponse{}
	for operation := range operations {
		response.Operations = append(response.Operations, &storagev1.Operation{
			Name:     operation,
			SpanKind: "client", // TODO: How do we know?
		})
	}
	klog.V(4).Infof("<-GetOperations %v", prototext.Format(response))
	return response, nil
}

func (s *Server) FindTraces(req *storagev1.FindTracesRequest, stream storagev1.SpanReaderPlugin_FindTracesServer) error {
	ctx := stream.Context()

	klog.V(2).Infof("FindTraces %v", prototext.Format(req))

	var opt FilterOptions
	opt.ServiceName = req.GetQuery().GetServiceName()

	// Note that we must return these in order of traceid, because that is what jaeger expects:
	// https://github.com/jaegertracing/jaeger/blob/7aeb457c21eed28b58cd34021eff14727229aa69/plugin/storage/grpc/shared/grpc_client.go#L201-L217
	traces := make(map[string]*storagev1.SpansResponseChunk)
	if err := s.visitSpans(ctx, opt, func(serviceName string, resourceSpans *v1.ResourceSpans) error {
		for _, scopeSpan := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				if operationName := req.GetQuery().GetOperationName(); operationName != "" {
					if operationName != span.Name {
						continue
					}
				}
				out := convertToJaeger(serviceName, span)

				key := string(span.TraceId)
				trace := traces[key]
				if trace == nil {
					trace = &storagev1.SpansResponseChunk{}
					traces[key] = trace
				}
				trace.Spans = append(trace.Spans, out)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	for _, trace := range traces {
		chunk := trace
		klog.V(4).Infof("<-FindTraces %v", prototext.Format(chunk))
		if err := stream.Send(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) FindTraceIDs(ctx context.Context, req *storagev1.FindTraceIDsRequest) (*storagev1.FindTraceIDsResponse, error) {
	klog.Warningf("FindTraceIDs not implemented: %v", prototext.Format(req))
	return nil, status.Errorf(codes.Unimplemented, "method FindTraceIDs not implemented")
}

func (s *Server) GetDependencies(ctx context.Context, req *storagev1.GetDependenciesRequest) (*storagev1.GetDependenciesResponse, error) {
	klog.Warningf("GetDependencies not implemented: %v", prototext.Format(req))
	return nil, status.Errorf(codes.Unimplemented, "method GetDependencies not implemented")
}

type TraceFile struct {
	data []*coltracepb.ExportTraceServiceRequest
}

type FilterOptions struct {
	ServiceName string
}

func (s *Server) visitSpans(ctx context.Context, opt FilterOptions, callback func(string, *v1.ResourceSpans) error) error {
	for _, traceFile := range s.traceFiles {
		if err := traceFile.visitSpans(ctx, opt, callback); err != nil {
			return err
		}
	}
	return nil
}

func (f *TraceFile) visitSpans(ctx context.Context, opt FilterOptions, callback func(string, *v1.ResourceSpans) error) error {
	for _, data := range f.data {
		// klog.Infof("data %v", prototext.Format(data))
		for _, span := range data.ResourceSpans {
			serviceName := ""
			for _, attr := range span.GetResource().GetAttributes() {
				if attr.GetKey() == "service.name" {
					serviceName = attr.GetValue().GetStringValue()
					break
				}
			}

			if opt.ServiceName != "" && serviceName != opt.ServiceName {
				continue
			}

			if err := callback(serviceName, span); err != nil {
				return err
			}

		}
	}
	return nil
}

func ReadTraceFile(ctx context.Context, p vfs.Path) (*TraceFile, error) {
	out := &TraceFile{}

	klog.Infof("reading file %v", p)
	// TODO: Caching & streaming
	b, err := p.ReadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading %v: %w", p, err)
	}

	r := bytes.NewReader(b)

	for {
		header := make([]byte, 16)
		if _, err := io.ReadFull(r, header); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading header: %w", err)
		}

		payloadLength := binary.BigEndian.Uint32(header[0:4])
		//checksum := binary.BigEndian.Uint32(header[4:8])
		flags := binary.BigEndian.Uint32(header[8:12])
		typeCode := binary.BigEndian.Uint32(header[12:16])

		// TODO: Verify checksum

		if flags != 0 {
			return nil, fmt.Errorf("unexpected flags value %v", flags)
		}

		// TODO: Sanity-check payloadLength

		payload := make([]byte, payloadLength)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, fmt.Errorf("reading payload: %w", err)
		}

		// TODO: Better typeCode parsing
		if typeCode == 1 {
			// TODO: process type definitions
		} else if typeCode == 32 {
			obj := &coltracepb.ExportTraceServiceRequest{}
			if err := proto.Unmarshal(payload, obj); err != nil {
				return nil, fmt.Errorf("parsing ExportTraceServiceRequest: %w", err)
			}
			out.data = append(out.data, obj)
		} else {
			return nil, fmt.Errorf("unexpected typecode ")
		}

	}

	return out, nil
}

func startTimeForSpan(span *v1.Span) *timestamppb.Timestamp {
	nanos := span.StartTimeUnixNano

	if nanos == 0 {
		return nil
	}
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &timestamppb.Timestamp{
		Seconds: int64(secs),
		Nanos:   int32(nanos),
	}
}

func durationForSpan(span *v1.Span) *durationpb.Duration {
	if span.EndTimeUnixNano == 0 || span.StartTimeUnixNano == 0 {
		return nil
	}
	nanos := span.EndTimeUnixNano - span.StartTimeUnixNano
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &durationpb.Duration{Seconds: int64(secs), Nanos: int32(nanos)}
}

func convertToJaeger(serviceName string, span *v1.Span) *v2.Span {

	out := &v2.Span{}
	out.TraceId = span.TraceId
	out.SpanId = span.SpanId
	out.OperationName = span.Name
	out.StartTime = startTimeForSpan(span)
	out.Duration = durationForSpan(span)
	out.Process = &v2.Process{
		ServiceName: serviceName,
	}
	if span.ParentSpanId != nil {
		out.References = append(out.References, &v2.SpanRef{
			TraceId: span.TraceId,
			SpanId:  span.ParentSpanId,
			RefType: v2.SpanRefType_CHILD_OF,
		})
	}
	// References    []*SpanRef             `protobuf:"bytes,4,rep,name=references,proto3" json:"references,omitempty"`
	// Flags         uint32                 `protobuf:"varint,5,opt,name=flags,proto3" json:"flags,omitempty"`
	// Tags          []*KeyValue            `protobuf:"bytes,8,rep,name=tags,proto3" json:"tags,omitempty"`
	// Logs          []*Log                 `protobuf:"bytes,9,rep,name=logs,proto3" json:"logs,omitempty"`
	// Process       *Process               `protobuf:"bytes,10,opt,name=process,proto3" json:"process,omitempty"`
	// ProcessId     string                 `protobuf:"bytes,11,opt,name=process_id,json=processId,proto3" json:"process_id,omitempty"`
	// Warnings      []string               `protobuf:"bytes,12,rep,name=warnings,proto3" json:"warnings,omitempty"`

	for _, attr := range span.Attributes {
		tag := &v2.KeyValue{
			Key: attr.GetKey(),
		}
		switch v := attr.GetValue().Value.(type) {
		case *v11.AnyValue_StringValue:
			tag.VStr = v.StringValue
		case *v11.AnyValue_IntValue:
			tag.VInt64 = v.IntValue
		case *v11.AnyValue_ArrayValue:
			s, err := attributeValueAsString(attr.GetValue())
			if err != nil {
				klog.Warningf("error converting array value: %v", err)
				s = "<?error>"
			}
			tag.VStr = s
		default:
			klog.Warningf("unhandled attribute type %T", v)
		}
		out.Tags = append(out.Tags, tag)
	}
	return out
}

func attributeValueAsString(v *v11.AnyValue) (string, error) {
	switch v := v.Value.(type) {
	case *v11.AnyValue_StringValue:
		return v.StringValue, nil
	case *v11.AnyValue_ArrayValue:
		var values []string
		for _, a := range v.ArrayValue.GetValues() {
			s, err := attributeValueAsString(a)
			if err != nil {
				klog.Warningf("error converting array value: %v", err)
				s = "<?error>"
			}
			values = append(values, s)
		}
		return "[" + strings.Join(values, ",") + "]", nil
	default:
		return "", fmt.Errorf("unhandled attribute type %T", v)
	}
}

// RunJaegerOptions are the options for runJaeger
type RunJaegerOptions struct {
	StorageServer string
	UseDocker     bool
}

// runJaeger starts the jaeger query & visualizer, binding to our storage server
func runJaeger(ctx context.Context, opt RunJaegerOptions) error {
	jaegerURL := "http://127.0.0.1:16686/"

	var jaeger *exec.Cmd
	{
		klog.Infof("starting jaeger")

		var c *exec.Cmd
		if opt.UseDocker {
			args := []string{
				"docker", "run", "--rm", "--network=host", "--name=jaeger",
				"-e=SPAN_STORAGE_TYPE=grpc-plugin",
				"jaegertracing/jaeger-query",
				"--grpc-storage.server=" + opt.StorageServer,
			}
			c = exec.CommandContext(ctx, args[0], args[1:]...)
		} else {
			args := []string{
				"jaeger-query",
				"--grpc-storage.server=" + opt.StorageServer,
			}
			c = exec.CommandContext(ctx, args[0], args[1:]...)
			c.Env = append(os.Environ(), "SPAN_STORAGE_TYPE=grpc-plugin")
		}

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Start(); err != nil {
			return fmt.Errorf("starting jaeger (%s): %w", strings.Join(c.Args, " "), err)
		}
		jaeger = c
	}

	{
		fmt.Fprintf(os.Stdout, "open browser to %s\n", jaegerURL)
		args := []string{"xdg-open", jaegerURL}
		c := exec.CommandContext(ctx, args[0], args[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("opening webbrowser (%s): %w", strings.Join(args, " "), err)
		}
	}

	if err := jaeger.Wait(); err != nil {
		return fmt.Errorf("waiting for jaeger to exit: %w", err)
	}

	return nil
}
