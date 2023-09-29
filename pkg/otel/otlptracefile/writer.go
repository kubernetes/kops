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

package otlptracefile

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
	"k8s.io/kops/pkg/otel/otlptracefile/pb"
)

type writer struct {
	fileMutex sync.Mutex
	f         *os.File

	typeCodesMutex sync.Mutex
	nextTypeCode   TypeCode
	typeCodes      map[string]TypeCode
}

type TypeCode uint32

func newWriter(cfg Config) (*writer, error) {
	f, err := os.OpenFile(cfg.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", cfg.path, err)
	}
	w := &writer{
		f: f,
	}
	w.nextTypeCode = 32
	w.typeCodes = make(map[string]TypeCode)
	w.recordWellKnownType(pb.WellKnownTypeCode_WellKnownTypeCode_ObjectType, &pb.ObjectType{})

	return w, nil
}

// writeTraces is called by the otel libraries to write a set of trace records.
func (w *writer) writeTraces(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) error {
	return w.writeObject(ctx, req)
}

// codeForType returns the integer code value for objects of obj.
// If this is the first time we've seen the type, this method will assign a code value, write it to the file and return it.
func (w *writer) codeForType(ctx context.Context, obj proto.Message) (TypeCode, error) {
	typeName := string(obj.ProtoReflect().Descriptor().FullName())

	w.typeCodesMutex.Lock()
	defer w.typeCodesMutex.Unlock()

	typeCode, found := w.typeCodes[typeName]
	if found {
		return typeCode, nil
	}

	typeCode = w.nextTypeCode
	w.nextTypeCode++

	record := &pb.ObjectType{
		TypeCode: uint32(typeCode),
		TypeName: typeName,
	}
	if err := w.writeObjectWithTypeCode(ctx, TypeCode(pb.WellKnownTypeCode_WellKnownTypeCode_ObjectType), record); err != nil {
		return 0, err
	}

	w.typeCodes[typeName] = typeCode
	return typeCode, nil
}

// recordWellKnownType is used to insert a "system" type into the table of type codes.
// This is used for the types that are needed to e.g. record the type code table itself.
func (w *writer) recordWellKnownType(typeCode pb.WellKnownTypeCode, obj proto.Message) {
	typeName := string(obj.ProtoReflect().Descriptor().FullName())

	w.typeCodesMutex.Lock()
	defer w.typeCodesMutex.Unlock()

	w.typeCodes[typeName] = TypeCode(typeCode)
}

// writeObject appends an object to the file
func (w *writer) writeObject(ctx context.Context, obj proto.Message) error {
	typeCode, err := w.codeForType(ctx, obj)
	if err != nil {
		return err
	}

	return w.writeObjectWithTypeCode(ctx, typeCode, obj)
}

// writeObjectWithTypeCode is the key function here.  We encode and write the object.
// We include a header that identifies the object using the provided typeCode.
func (w *writer) writeObjectWithTypeCode(ctx context.Context, typeCode TypeCode, obj proto.Message) error {
	buf, err := proto.Marshal(obj)
	if err != nil {
		return fmt.Errorf("converting to proto: %w", err)
	}

	crc32q := crc32.MakeTable(crc32.Castagnoli)
	checksum := crc32.Checksum(buf, crc32q)

	flags := uint32(0)

	w.fileMutex.Lock()
	defer w.fileMutex.Unlock()

	if w.f == nil {
		return fmt.Errorf("already closed")
	}

	// write the object with a header.
	header := make([]byte, 16)
	binary.BigEndian.PutUint32(header[0:4], uint32(len(buf)))
	binary.BigEndian.PutUint32(header[4:8], checksum)
	binary.BigEndian.PutUint32(header[8:12], flags)
	binary.BigEndian.PutUint32(header[12:16], uint32(typeCode))

	if _, err := w.f.Write(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	if _, err := w.f.Write(buf); err != nil {
		// TODO: Rotate file?
		return fmt.Errorf("writing body: %w", err)
	}

	return nil
}

// Close closes the output file.
func (w *writer) Close() error {
	w.fileMutex.Lock()
	defer w.fileMutex.Unlock()

	if w.f != nil {
		if err := w.f.Close(); err != nil {
			return err
		}
		w.f = nil
	}

	return nil
}
