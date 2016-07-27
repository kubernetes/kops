// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bttest

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"
	btdpb "google.golang.org/cloud/bigtable/internal/data_proto"
	btspb "google.golang.org/cloud/bigtable/internal/service_proto"
	bttdpb "google.golang.org/cloud/bigtable/internal/table_data_proto"
	bttspb "google.golang.org/cloud/bigtable/internal/table_service_proto"
)

func TestConcurrentMutationsAndGC(t *testing.T) {
	s := &server{
		tables: make(map[string]*table),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := s.CreateTable(
		ctx,
		&bttspb.CreateTableRequest{Name: "cluster", TableId: "t"}); err != nil {
		t.Fatal(err)
	}
	const name = `cluster/tables/t`
	tbl := s.tables[name]
	req := &bttspb.ModifyColumnFamiliesRequest{
		Name: name,
		Modifications: []*bttspb.ModifyColumnFamiliesRequest_Modification{
			{
				Id:  "cf",
				Mod: &bttspb.ModifyColumnFamiliesRequest_Modification_Create{Create: &bttdpb.ColumnFamily{}},
			},
		},
	}
	_, err := s.ModifyColumnFamilies(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	req = &bttspb.ModifyColumnFamiliesRequest{
		Name: name,
		Modifications: []*bttspb.ModifyColumnFamiliesRequest_Modification{
			{
				Id: "cf",
				Mod: &bttspb.ModifyColumnFamiliesRequest_Modification_Update{
					Update: &bttdpb.ColumnFamily{GcRule: &bttdpb.GcRule{Rule: &bttdpb.GcRule_MaxNumVersions{MaxNumVersions: 1}}}},
			},
		},
	}
	if _, err := s.ModifyColumnFamilies(ctx, req); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var ts int64
	ms := func() []*btdpb.Mutation {
		return []*btdpb.Mutation{
			{
				Mutation: &btdpb.Mutation_SetCell_{
					SetCell: &btdpb.Mutation_SetCell{
						FamilyName:      "cf",
						ColumnQualifier: []byte(`col`),
						TimestampMicros: atomic.AddInt64(&ts, 1000),
					},
				},
			},
		}
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				req := &btspb.MutateRowRequest{
					TableName: name,
					RowKey:    []byte(fmt.Sprint(rand.Intn(100))),
					Mutations: ms(),
				}
				s.MutateRow(ctx, req)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			tbl.gc()
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Concurrent mutations and GCs haven't completed after 100ms")
	}
}
