/*
Copyright 2015 Google Inc. All Rights Reserved.

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

/*
Package bttest contains test helpers for working with the bigtable package.

To use a Server, create it, and then connect to it with no security:
(The project/instance values are ignored.)
	srv, err := bttest.NewServer("127.0.0.1:0")
	...
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	...
	client, err := bigtable.NewClient(ctx, proj, instance,
		cloud.WithBaseGRPC(conn))
	...
*/
package bttest // import "google.golang.org/cloud/bigtable/bttest"

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	emptypb "github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/net/context"
	btdpb "google.golang.org/cloud/bigtable/internal/data_proto"
	rpcpb "google.golang.org/cloud/bigtable/internal/rpc_status_proto"
	btspb "google.golang.org/cloud/bigtable/internal/service_proto"
	bttdpb "google.golang.org/cloud/bigtable/internal/table_data_proto"
	bttspb "google.golang.org/cloud/bigtable/internal/table_service_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Server is an in-memory Cloud Bigtable fake.
// It is unauthenticated, and only a rough approximation.
type Server struct {
	Addr string

	l   net.Listener
	srv *grpc.Server
	s   *server
}

// server is the real implementation of the fake.
// It is a separate and unexported type so the API won't be cluttered with
// methods that are only relevant to the fake's implementation.
type server struct {
	mu     sync.Mutex
	tables map[string]*table // keyed by fully qualified name
	gcc    chan int          // set when gcloop starts, closed when server shuts down

	// Any unimplemented methods will cause a panic.
	bttspb.BigtableTableAdminServer
	btspb.BigtableServer
}

// NewServer creates a new Server.
// The Server will be listening for gRPC connections, without TLS,
// on the provided address. The resolved address is named by the Addr field.
func NewServer(laddr string) (*Server, error) {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Addr: l.Addr().String(),
		l:    l,
		srv:  grpc.NewServer(),
		s: &server{
			tables: make(map[string]*table),
		},
	}
	bttspb.RegisterBigtableTableAdminServer(s.srv, s.s)
	btspb.RegisterBigtableServer(s.srv, s.s)

	go s.srv.Serve(s.l)

	return s, nil
}

// Close shuts down the server.
func (s *Server) Close() {
	s.s.mu.Lock()
	if s.s.gcc != nil {
		close(s.s.gcc)
	}
	s.s.mu.Unlock()

	s.srv.Stop()
	s.l.Close()
}

func (s *server) CreateTable(ctx context.Context, req *bttspb.CreateTableRequest) (*bttdpb.Table, error) {
	tbl := req.Name + "/tables/" + req.TableId

	s.mu.Lock()
	if _, ok := s.tables[tbl]; ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("table %q already exists", tbl)
	}
	s.tables[tbl] = newTable()
	s.mu.Unlock()

	return &bttdpb.Table{Name: tbl}, nil
}

func (s *server) ListTables(ctx context.Context, req *bttspb.ListTablesRequest) (*bttspb.ListTablesResponse, error) {
	res := &bttspb.ListTablesResponse{}
	prefix := req.Name + "/tables/"

	s.mu.Lock()
	for tbl := range s.tables {
		if strings.HasPrefix(tbl, prefix) {
			res.Tables = append(res.Tables, &bttdpb.Table{Name: tbl})
		}
	}
	s.mu.Unlock()

	return res, nil
}

func (s *server) GetTable(ctx context.Context, req *bttspb.GetTableRequest) (*bttdpb.Table, error) {
	tbl := req.Name

	s.mu.Lock()
	tblIns, ok := s.tables[tbl]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("table %q not found", tbl)
	}

	return &bttdpb.Table{
		Name:           tbl,
		ColumnFamilies: toColumnFamilies(tblIns.columnFamilies()),
	}, nil
}

func (s *server) DeleteTable(ctx context.Context, req *bttspb.DeleteTableRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tables[req.Name]; !ok {
		return nil, fmt.Errorf("no such table %q", req.Name)
	}
	delete(s.tables, req.Name)
	return &emptypb.Empty{}, nil
}

func (s *server) ModifyColumnFamilies(ctx context.Context, req *bttspb.ModifyColumnFamiliesRequest) (*bttdpb.Table, error) {
	tblName := req.Name[strings.LastIndex(req.Name, "/")+1:]

	s.mu.Lock()
	tbl, ok := s.tables[req.Name]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such table %q", req.Name)
	}

	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	for _, mod := range req.Modifications {
		if create := mod.GetCreate(); create != nil {
			if _, ok := tbl.families[mod.Id]; ok {
				return nil, fmt.Errorf("family %q already exists", mod.Id)
			}
			newcf := &columnFamily{
				name:   req.Name + "/columnFamilies/" + mod.Id,
				gcRule: create.GcRule,
			}
			tbl.families[mod.Id] = newcf
		} else if mod.GetDrop() {
			if _, ok := tbl.families[mod.Id]; !ok {
				return nil, fmt.Errorf("can't delete unknown family %q", mod.Id)
			}
			delete(tbl.families, mod.Id)
		} else if modify := mod.GetUpdate(); modify != nil {
			if _, ok := tbl.families[mod.Id]; !ok {
				return nil, fmt.Errorf("no such family %q", mod.Id)
			}
			newcf := &columnFamily{
				name:   req.Name + "/columnFamilies/" + mod.Id,
				gcRule: modify.GcRule,
			}
			// assume that we ALWAYS want to replace by the new setting
			// we may need partial update through
			tbl.families[mod.Id] = newcf
		}
	}

	s.needGC()
	return &bttdpb.Table{
		Name:           tblName,
		ColumnFamilies: toColumnFamilies(tbl.families),
	}, nil
}

func (s *server) ReadRows(req *btspb.ReadRowsRequest, stream btspb.Bigtable_ReadRowsServer) error {
	s.mu.Lock()
	tbl, ok := s.tables[req.TableName]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no such table %q", req.TableName)
	}

	// Rows to read can be specified by a set of row keys and/or a set of row ranges.
	// Output is a stream of sorted, de-duped rows.
	tbl.mu.RLock()

	rowSet := make(map[string]*row)
	// Add the explicitly given keys
	for _, key := range req.Rows.RowKeys {
		start := string(key)
		addRows(start, start+"\x00", tbl, rowSet)
	}

	// Add keys from row ranges
	for _, rr := range req.Rows.RowRanges {
		var start, end string
		switch sk := rr.StartKey.(type) {
		case *btdpb.RowRange_StartKeyClosed:
			start = string(sk.StartKeyClosed)
		case *btdpb.RowRange_StartKeyOpen:
			start = string(sk.StartKeyOpen) + "\x00"
		}
		switch ek := rr.EndKey.(type) {
		case *btdpb.RowRange_EndKeyClosed:
			end = string(ek.EndKeyClosed) + "\x00"
		case *btdpb.RowRange_EndKeyOpen:
			end = string(ek.EndKeyOpen)
		}

		addRows(start, end, tbl, rowSet)
	}
	tbl.mu.RUnlock()

	rows := make([]*row, 0, len(rowSet))
	for _, r := range rowSet {
		rows = append(rows, r)
	}
	sort.Sort(byRowKey(rows))

	for _, r := range rows {
		if err := streamRow(stream, r, req.Filter); err != nil {
			return err
		}
	}
	return nil
}

func addRows(start, end string, tbl *table, rowSet map[string]*row) {
	si, ei := 0, len(tbl.rows) // half-open interval
	if start != "" {
		si = sort.Search(len(tbl.rows), func(i int) bool { return tbl.rows[i].key >= start })
	}
	if end != "" {
		ei = sort.Search(len(tbl.rows), func(i int) bool { return tbl.rows[i].key >= end })
	}
	if si < ei {
		for _, row := range tbl.rows[si:ei] {
			rowSet[row.key] = row
		}
	}
}

func streamRow(stream btspb.Bigtable_ReadRowsServer, r *row, f *btdpb.RowFilter) error {
	r.mu.Lock()
	nr := r.copy()
	r.mu.Unlock()
	r = nr

	filterRow(f, r)

	rrr := &btspb.ReadRowsResponse{}
	for col, cells := range r.cells {
		i := strings.Index(col, ":") // guaranteed to exist
		fam, col := col[:i], col[i+1:]
		if len(cells) == 0 {
			continue
		}
		// TODO(dsymonds): Apply transformers.
		for _, cell := range cells {
			rrr.Chunks = append(rrr.Chunks, &btspb.ReadRowsResponse_CellChunk{
				RowKey:          []byte(r.key),
				FamilyName:      &wrappers.StringValue{Value: fam},
				Qualifier:       &wrappers.BytesValue{Value: []byte(col)},
				TimestampMicros: cell.ts,
				Value:           cell.value,
			})
		}
	}
	// We can't have a cell with just COMMIT set, which would imply a new empty cell.
	// So modify the last cell to have the COMMIT flag set.
	if len(rrr.Chunks) > 0 {
		rrr.Chunks[len(rrr.Chunks)-1].RowStatus = &btspb.ReadRowsResponse_CellChunk_CommitRow{CommitRow: true}
	}

	return stream.Send(rrr)
}

// filterRow modifies a row with the given filter.
func filterRow(f *btdpb.RowFilter, r *row) {
	if f == nil {
		return
	}
	// Handle filters that apply beyond just including/excluding cells.
	switch f := f.Filter.(type) {
	case *btdpb.RowFilter_Chain_:
		for _, sub := range f.Chain.Filters {
			filterRow(sub, r)
		}
		return
	case *btdpb.RowFilter_Interleave_:
		srs := make([]*row, 0, len(f.Interleave.Filters))
		for _, sub := range f.Interleave.Filters {
			sr := r.copy()
			filterRow(sub, sr)
			srs = append(srs, sr)
		}
		// merge
		// TODO(dsymonds): is this correct?
		r.cells = make(map[string][]cell)
		for _, sr := range srs {
			for col, cs := range sr.cells {
				r.cells[col] = append(r.cells[col], cs...)
			}
		}
		for _, cs := range r.cells {
			sort.Sort(byDescTS(cs))
		}
		return
	case *btdpb.RowFilter_CellsPerColumnLimitFilter:
		lim := int(f.CellsPerColumnLimitFilter)
		for col, cs := range r.cells {
			if len(cs) > lim {
				r.cells[col] = cs[:lim]
			}
		}
		return
	}

	// Any other case, operate on a per-cell basis.
	for key, cs := range r.cells {
		i := strings.Index(key, ":") // guaranteed to exist
		fam, col := key[:i], key[i+1:]
		r.cells[key] = filterCells(f, fam, col, cs)
	}
}

func filterCells(f *btdpb.RowFilter, fam, col string, cs []cell) []cell {
	var ret []cell
	for _, cell := range cs {
		if includeCell(f, fam, col, cell) {
			ret = append(ret, cell)
		}
	}
	return ret
}

func includeCell(f *btdpb.RowFilter, fam, col string, cell cell) bool {
	if f == nil {
		return true
	}
	// TODO(dsymonds): Implement many more filters.
	switch f := f.Filter.(type) {
	default:
		log.Printf("WARNING: don't know how to handle filter of type %T (ignoring it)", f)
		return true
	case *btdpb.RowFilter_FamilyNameRegexFilter:
		pat := string(f.FamilyNameRegexFilter)
		rx, err := regexp.Compile(pat)
		if err != nil {
			log.Printf("Bad family_name_regex_filter pattern %q: %v", pat, err)
			return false
		}
		return rx.MatchString(fam)
	case *btdpb.RowFilter_ColumnQualifierRegexFilter:
		pat := string(f.ColumnQualifierRegexFilter)
		rx, err := regexp.Compile(pat)
		if err != nil {
			log.Printf("Bad column_qualifier_regex_filter pattern %q: %v", pat, err)
			return false
		}
		return rx.MatchString(col)
	case *btdpb.RowFilter_ValueRegexFilter:
		pat := string(f.ValueRegexFilter)
		rx, err := regexp.Compile(pat)
		if err != nil {
			log.Printf("Bad value_regex_filter pattern %q: %v", pat, err)
			return false
		}
		return rx.Match(cell.value)
	}
}

func (s *server) MutateRow(ctx context.Context, req *btspb.MutateRowRequest) (*btspb.MutateRowResponse, error) {
	s.mu.Lock()
	tbl, ok := s.tables[req.TableName]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such table %q", req.TableName)
	}

	f := tbl.columnFamiliesSet()
	r := tbl.mutableRow(string(req.RowKey))
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := applyMutations(tbl, r, req.Mutations, f); err != nil {
		return nil, err
	}
	return &btspb.MutateRowResponse{}, nil
}

func (s *server) MutateRows(req *btspb.MutateRowsRequest, stream btspb.Bigtable_MutateRowsServer) error {
	s.mu.Lock()
	tbl, ok := s.tables[req.TableName]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no such table %q", req.TableName)
	}

	res := &btspb.MutateRowsResponse{Entries: make([]*btspb.MutateRowsResponse_Entry, len(req.Entries))}

	f := tbl.columnFamiliesSet()

	for i, entry := range req.Entries {
		r := tbl.mutableRow(string(entry.RowKey))
		r.mu.Lock()
		code, msg := int32(codes.OK), ""
		if err := applyMutations(tbl, r, entry.Mutations, f); err != nil {
			code = int32(codes.Internal)
			msg = err.Error()
		}
		res.Entries[i] = &btspb.MutateRowsResponse_Entry{
			Index:  int64(i),
			Status: &rpcpb.Status{Code: code, Message: msg},
		}
		r.mu.Unlock()
	}
	stream.Send(res)
	return nil
}

func (s *server) CheckAndMutateRow(ctx context.Context, req *btspb.CheckAndMutateRowRequest) (*btspb.CheckAndMutateRowResponse, error) {
	s.mu.Lock()
	tbl, ok := s.tables[req.TableName]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such table %q", req.TableName)
	}

	res := &btspb.CheckAndMutateRowResponse{}

	f := tbl.columnFamiliesSet()

	r := tbl.mutableRow(string(req.RowKey))
	r.mu.Lock()
	defer r.mu.Unlock()

	// Figure out which mutation to apply.
	whichMut := false
	if req.PredicateFilter == nil {
		// Use true_mutations iff row contains any cells.
		whichMut = len(r.cells) > 0
	} else {
		// Use true_mutations iff any cells in the row match the filter.
		// TODO(dsymonds): This could be cheaper.
		nr := r.copy()
		filterRow(req.PredicateFilter, nr)
		for _, cs := range nr.cells {
			if len(cs) > 0 {
				whichMut = true
				break
			}
		}
		// TODO(dsymonds): Figure out if this is supposed to be set
		// even when there's no predicate filter.
		res.PredicateMatched = whichMut
	}
	muts := req.FalseMutations
	if whichMut {
		muts = req.TrueMutations
	}

	if err := applyMutations(tbl, r, muts, f); err != nil {
		return nil, err
	}
	return res, nil
}

// applyMutations applies a sequence of mutations to a row.
// fam should be a snapshot of the keys of tbl.families.
// It assumes r.mu is locked.
func applyMutations(tbl *table, r *row, muts []*btdpb.Mutation, fam map[string]bool) error {
	for _, mut := range muts {
		switch mut := mut.Mutation.(type) {
		default:
			return fmt.Errorf("can't handle mutation type %T", mut)
		case *btdpb.Mutation_SetCell_:
			set := mut.SetCell
			if !fam[set.FamilyName] {
				return fmt.Errorf("unknown family %q", set.FamilyName)
			}
			ts := set.TimestampMicros
			if ts == -1 { // bigtable.ServerTime
				ts = newTimestamp()
			}
			if !tbl.validTimestamp(ts) {
				return fmt.Errorf("invalid timestamp %d", ts)
			}
			col := fmt.Sprintf("%s:%s", set.FamilyName, set.ColumnQualifier)

			newCell := cell{ts: ts, value: set.Value}
			r.cells[col] = appendOrReplaceCell(r.cells[col], newCell)
		case *btdpb.Mutation_DeleteFromColumn_:
			del := mut.DeleteFromColumn
			col := fmt.Sprintf("%s:%s", del.FamilyName, del.ColumnQualifier)

			cs := r.cells[col]
			if del.TimeRange != nil {
				tsr := del.TimeRange
				if !tbl.validTimestamp(tsr.StartTimestampMicros) {
					return fmt.Errorf("invalid timestamp %d", tsr.StartTimestampMicros)
				}
				if !tbl.validTimestamp(tsr.EndTimestampMicros) {
					return fmt.Errorf("invalid timestamp %d", tsr.EndTimestampMicros)
				}
				// Find half-open interval to remove.
				// Cells are in descending timestamp order,
				// so the predicates to sort.Search are inverted.
				si, ei := 0, len(cs)
				if tsr.StartTimestampMicros > 0 {
					ei = sort.Search(len(cs), func(i int) bool { return cs[i].ts < tsr.StartTimestampMicros })
				}
				if tsr.EndTimestampMicros > 0 {
					si = sort.Search(len(cs), func(i int) bool { return cs[i].ts < tsr.EndTimestampMicros })
				}
				if si < ei {
					copy(cs[si:], cs[ei:])
					cs = cs[:len(cs)-(ei-si)]
				}
			} else {
				cs = nil
			}
			if len(cs) == 0 {
				delete(r.cells, col)
			} else {
				r.cells[col] = cs
			}
		case *btdpb.Mutation_DeleteFromRow_:
			r.cells = make(map[string][]cell)
		}
	}
	return nil
}

func maxTimestamp(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func newTimestamp() int64 {
	ts := time.Now().UnixNano() / 1e3
	ts -= ts % 1000 // round to millisecond granularity
	return ts
}

func appendOrReplaceCell(cs []cell, newCell cell) []cell {
	replaced := false
	for i, cell := range cs {
		if cell.ts == newCell.ts {
			cs[i] = newCell
			replaced = true
			break
		}
	}
	if !replaced {
		cs = append(cs, newCell)
	}
	sort.Sort(byDescTS(cs))
	return cs
}

func (s *server) ReadModifyWriteRow(ctx context.Context, req *btspb.ReadModifyWriteRowRequest) (*btspb.ReadModifyWriteRowResponse, error) {
	s.mu.Lock()
	tbl, ok := s.tables[req.TableName]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no such table %q", req.TableName)
	}

	updates := make(map[string]cell) // copy of updated cells; keyed by full column name

	r := tbl.mutableRow(string(req.RowKey))
	r.mu.Lock()
	defer r.mu.Unlock()
	// Assume all mutations apply to the most recent version of the cell.
	// TODO(dsymonds): Verify this assumption and document it in the proto.
	for _, rule := range req.Rules {
		tbl.mu.RLock()
		_, famOK := tbl.families[rule.FamilyName]
		tbl.mu.RUnlock()
		if !famOK {
			return nil, fmt.Errorf("unknown family %q", rule.FamilyName)
		}

		key := fmt.Sprintf("%s:%s", rule.FamilyName, rule.ColumnQualifier)

		cells := r.cells[key]
		ts := newTimestamp()
		var newCell, prevCell cell
		isEmpty := len(cells) == 0
		if !isEmpty {
			prevCell = cells[0]

			// ts is the max of now or the prev cell's timestamp in case the
			// prev cell is in the future
			ts = maxTimestamp(ts, prevCell.ts)
		}

		switch rule := rule.Rule.(type) {
		default:
			return nil, fmt.Errorf("unknown RMW rule oneof %T", rule)
		case *btdpb.ReadModifyWriteRule_AppendValue:
			newCell = cell{ts: ts, value: append(prevCell.value, rule.AppendValue...)}
		case *btdpb.ReadModifyWriteRule_IncrementAmount:
			var v int64
			if !isEmpty {
				prevVal := prevCell.value
				if len(prevVal) != 8 {
					return nil, fmt.Errorf("increment on non-64-bit value")
				}
				v = int64(binary.BigEndian.Uint64(prevVal))
			}
			v += rule.IncrementAmount
			var val [8]byte
			binary.BigEndian.PutUint64(val[:], uint64(v))
			newCell = cell{ts: ts, value: val[:]}
		}
		updates[key] = newCell
		r.cells[key] = appendOrReplaceCell(r.cells[key], newCell)
	}

	res := &btdpb.Row{
		Key: req.RowKey,
	}
	for col, cell := range updates {
		i := strings.Index(col, ":")
		fam, qual := col[:i], col[i+1:]
		var f *btdpb.Family
		for _, ff := range res.Families {
			if ff.Name == fam {
				f = ff
				break
			}
		}
		if f == nil {
			f = &btdpb.Family{Name: fam}
			res.Families = append(res.Families, f)
		}
		f.Columns = append(f.Columns, &btdpb.Column{
			Qualifier: []byte(qual),
			Cells: []*btdpb.Cell{{
				Value: cell.value,
			}},
		})
	}
	return &btspb.ReadModifyWriteRowResponse{Row: res}, nil
}

// needGC is invoked whenever the server needs gcloop running.
func (s *server) needGC() {
	s.mu.Lock()
	if s.gcc == nil {
		s.gcc = make(chan int)
		go s.gcloop(s.gcc)
	}
	s.mu.Unlock()
}

func (s *server) gcloop(done <-chan int) {
	const (
		minWait = 500  // ms
		maxWait = 1500 // ms
	)

	for {
		// Wait for a random time interval.
		d := time.Duration(minWait+rand.Intn(maxWait-minWait)) * time.Millisecond
		select {
		case <-time.After(d):
		case <-done:
			return // server has been closed
		}

		// Do a GC pass over all tables.
		var tables []*table
		s.mu.Lock()
		for _, tbl := range s.tables {
			tables = append(tables, tbl)
		}
		s.mu.Unlock()
		for _, tbl := range tables {
			tbl.gc()
		}
	}
}

type table struct {
	mu       sync.RWMutex
	families map[string]*columnFamily // keyed by plain family name
	rows     []*row                   // sorted by row key
	rowIndex map[string]*row          // indexed by row key
}

func newTable() *table {
	return &table{
		families: make(map[string]*columnFamily),
		rowIndex: make(map[string]*row),
	}
}

func (t *table) validTimestamp(ts int64) bool {
	// Assume millisecond granularity is required.
	return ts%1000 == 0
}

func (t *table) columnFamilies() map[string]*columnFamily {
	cp := make(map[string]*columnFamily)
	t.mu.RLock()
	for fam, cf := range t.families {
		cp[fam] = cf
	}
	t.mu.RUnlock()
	return cp
}

func (t *table) columnFamiliesSet() map[string]bool {
	f := make(map[string]bool)
	for fam := range t.columnFamilies() {
		f[fam] = true
	}
	return f
}

func (t *table) mutableRow(row string) *row {
	// Try fast path first.
	t.mu.RLock()
	r := t.rowIndex[row]
	t.mu.RUnlock()
	if r != nil {
		return r
	}

	// We probably need to create the row.
	t.mu.Lock()
	r = t.rowIndex[row]
	if r == nil {
		r = newRow(row)
		t.rowIndex[row] = r
		t.rows = append(t.rows, r)
		sort.Sort(byRowKey(t.rows)) // yay, inefficient!
	}
	t.mu.Unlock()
	return r
}

func (t *table) gc() {
	// This method doesn't add or remove rows, so we only need a read lock for the table.
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Gather GC rules we'll apply.
	rules := make(map[string]*bttdpb.GcRule) // keyed by "fam"
	for fam, cf := range t.families {
		if cf.gcRule != nil {
			rules[fam] = cf.gcRule
		}
	}
	if len(rules) == 0 {
		return
	}

	for _, r := range t.rows {
		r.mu.Lock()
		r.gc(rules)
		r.mu.Unlock()
	}
}

type byRowKey []*row

func (b byRowKey) Len() int           { return len(b) }
func (b byRowKey) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byRowKey) Less(i, j int) bool { return b[i].key < b[j].key }

type row struct {
	key string

	mu    sync.Mutex
	cells map[string][]cell // keyed by full column name; cells are in descending timestamp order
}

func newRow(key string) *row {
	return &row{
		key:   key,
		cells: make(map[string][]cell),
	}
}

// copy returns a copy of the row.
// Cell values are aliased.
// r.mu should be held.
func (r *row) copy() *row {
	nr := &row{
		key:   r.key,
		cells: make(map[string][]cell, len(r.cells)),
	}
	for col, cs := range r.cells {
		// Copy the []cell slice, but not the []byte inside each cell.
		nr.cells[col] = append([]cell(nil), cs...)
	}
	return nr
}

// gc applies the given GC rules to the row.
// r.mu should be held.
func (r *row) gc(rules map[string]*bttdpb.GcRule) {
	for col, cs := range r.cells {
		fam := col[:strings.Index(col, ":")]
		rule, ok := rules[fam]
		if !ok {
			continue
		}
		r.cells[col] = applyGC(cs, rule)
	}
}

var gcTypeWarn sync.Once

// applyGC applies the given GC rule to the cells.
func applyGC(cells []cell, rule *bttdpb.GcRule) []cell {
	switch rule := rule.Rule.(type) {
	default:
		// TODO(dsymonds): Support GcRule_Intersection_
		gcTypeWarn.Do(func() {
			log.Printf("Unsupported GC rule type %T", rule)
		})
	case *bttdpb.GcRule_Union_:
		for _, sub := range rule.Union.Rules {
			cells = applyGC(cells, sub)
		}
		return cells
	case *bttdpb.GcRule_MaxAge:
		// Timestamps are in microseconds.
		cutoff := time.Now().UnixNano() / 1e3
		cutoff -= rule.MaxAge.Seconds * 1e6
		cutoff -= int64(rule.MaxAge.Nanos) / 1e3
		// The slice of cells in in descending timestamp order.
		// This sort.Search will return the index of the first cell whose timestamp is chronologically before the cutoff.
		si := sort.Search(len(cells), func(i int) bool { return cells[i].ts < cutoff })
		if si < len(cells) {
			log.Printf("bttest: GC MaxAge(%v) deleted %d cells.", rule.MaxAge, len(cells)-si)
		}
		return cells[:si]
	case *bttdpb.GcRule_MaxNumVersions:
		n := int(rule.MaxNumVersions)
		if len(cells) > n {
			cells = cells[:n]
		}
		return cells
	}
	return cells
}

type cell struct {
	ts    int64
	value []byte
}

type byDescTS []cell

func (b byDescTS) Len() int           { return len(b) }
func (b byDescTS) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byDescTS) Less(i, j int) bool { return b[i].ts > b[j].ts }

type columnFamily struct {
	name   string
	gcRule *bttdpb.GcRule
}

func (c *columnFamily) proto() *bttdpb.ColumnFamily {
	return &bttdpb.ColumnFamily{
		GcRule: c.gcRule,
	}
}

func toColumnFamilies(families map[string]*columnFamily) map[string]*bttdpb.ColumnFamily {
	f := make(map[string]*bttdpb.ColumnFamily)
	for k, v := range families {
		f[k] = v.proto()
	}
	return f
}
