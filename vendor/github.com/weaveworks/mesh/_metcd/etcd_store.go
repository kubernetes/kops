package metcd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/lease"
	"github.com/coreos/etcd/mvcc"
	"github.com/coreos/etcd/mvcc/backend"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/gogo/protobuf/proto"
	"github.com/weaveworks/mesh"
	"golang.org/x/net/context"
)

// Transport-agnostic reimplementation of coreos/etcd/etcdserver. The original
// is unsuitable because it is tightly coupled to persistent storage, an HTTP
// transport, etc. Implements selected etcd V3 API (gRPC) methods.
type etcdStore struct {
	proposalc   chan<- []byte
	snapshotc   <-chan raftpb.Snapshot
	entryc      <-chan raftpb.Entry
	confentryc  chan<- raftpb.Entry
	actionc     chan func()
	quitc       chan struct{}
	terminatedc chan struct{}
	logger      mesh.Logger

	dbPath string // please os.RemoveAll on exit
	kv     mvcc.KV
	lessor lease.Lessor
	index  *consistentIndex // see comment on type

	idgen   <-chan uint64
	pending map[uint64]responseChans
}

var _ Server = &etcdStore{}

func newEtcdStore(
	proposalc chan<- []byte,
	snapshotc <-chan raftpb.Snapshot,
	entryc <-chan raftpb.Entry,
	confentryc chan<- raftpb.Entry,
	logger mesh.Logger,
) *etcdStore {
	// It would be much better if we could have a proper in-memory backend. Alas:
	// backend.Backend is tightly coupled to bolt.DB, and both are tightly coupled
	// to os.Open &c. So we'd need to fork both Bolt and backend. A task for
	// another day.
	f, err := ioutil.TempFile(os.TempDir(), "mesh_etcd_backend_")
	if err != nil {
		panic(err)
	}
	dbPath := f.Name()
	f.Close()
	logger.Printf("etcd store: using %s", dbPath)

	b := backend.NewDefaultBackend(dbPath)
	lessor := lease.NewLessor(b)
	index := &consistentIndex{0}
	kv := mvcc.New(b, lessor, index)

	s := &etcdStore{
		proposalc:   proposalc,
		snapshotc:   snapshotc,
		entryc:      entryc,
		confentryc:  confentryc,
		actionc:     make(chan func()),
		quitc:       make(chan struct{}),
		terminatedc: make(chan struct{}),
		logger:      logger,

		dbPath: dbPath,
		kv:     kv,
		lessor: lessor,
		index:  index,

		idgen:   makeIDGen(),
		pending: map[uint64]responseChans{},
	}
	go s.loop()
	return s
}

// Range implements gRPC KVServer.
// Range gets the keys in the range from the store.
func (s *etcdStore) Range(ctx context.Context, req *etcdserverpb.RangeRequest) (*etcdserverpb.RangeResponse, error) {
	ireq := etcdserverpb.InternalRaftRequest{ID: <-s.idgen, Range: req}
	msgc, errc, err := s.proposeInternalRaftRequest(ireq)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		s.cancelInternalRaftRequest(ireq)
		return nil, ctx.Err()
	case msg := <-msgc:
		return msg.(*etcdserverpb.RangeResponse), nil
	case err := <-errc:
		return nil, err
	case <-s.quitc:
		return nil, errStopped
	}
}

// Put implements gRPC KVServer.
// Put puts the given key into the store.
// A put request increases the revision of the store,
// and generates one event in the event history.
func (s *etcdStore) Put(ctx context.Context, req *etcdserverpb.PutRequest) (*etcdserverpb.PutResponse, error) {
	ireq := etcdserverpb.InternalRaftRequest{ID: <-s.idgen, Put: req}
	msgc, errc, err := s.proposeInternalRaftRequest(ireq)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		s.cancelInternalRaftRequest(ireq)
		return nil, ctx.Err()
	case msg := <-msgc:
		return msg.(*etcdserverpb.PutResponse), nil
	case err := <-errc:
		return nil, err
	case <-s.quitc:
		return nil, errStopped
	}
}

// Delete implements gRPC KVServer.
// Delete deletes the given range from the store.
// A delete request increase the revision of the store,
// and generates one event in the event history.
func (s *etcdStore) DeleteRange(ctx context.Context, req *etcdserverpb.DeleteRangeRequest) (*etcdserverpb.DeleteRangeResponse, error) {
	ireq := etcdserverpb.InternalRaftRequest{ID: <-s.idgen, DeleteRange: req}
	msgc, errc, err := s.proposeInternalRaftRequest(ireq)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		s.cancelInternalRaftRequest(ireq)
		return nil, ctx.Err()
	case msg := <-msgc:
		return msg.(*etcdserverpb.DeleteRangeResponse), nil
	case err := <-errc:
		return nil, err
	case <-s.quitc:
		return nil, errStopped
	}
}

// Txn implements gRPC KVServer.
// Txn processes all the requests in one transaction.
// A txn request increases the revision of the store,
// and generates events with the same revision in the event history.
// It is not allowed to modify the same key several times within one txn.
func (s *etcdStore) Txn(ctx context.Context, req *etcdserverpb.TxnRequest) (*etcdserverpb.TxnResponse, error) {
	ireq := etcdserverpb.InternalRaftRequest{ID: <-s.idgen, Txn: req}
	msgc, errc, err := s.proposeInternalRaftRequest(ireq)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		s.cancelInternalRaftRequest(ireq)
		return nil, ctx.Err()
	case msg := <-msgc:
		return msg.(*etcdserverpb.TxnResponse), nil
	case err := <-errc:
		return nil, err
	case <-s.quitc:
		return nil, errStopped
	}
}

// Compact implements gRPC KVServer.
// Compact compacts the event history in s. User should compact the
// event history periodically, or it will grow infinitely.
func (s *etcdStore) Compact(ctx context.Context, req *etcdserverpb.CompactionRequest) (*etcdserverpb.CompactionResponse, error) {
	// We don't have snapshotting yet, so compact just puts us in a bad state.
	// TODO(pb): fix this when we implement snapshotting.
	return nil, errors.New("not implemented")
}

// The "consistent index" is the index number of the most recent committed
// entry. This logical value is duplicated and tracked in multiple places
// throughout the etcd server and storage code.
//
// For our part, we are expected to store one instance of this number, setting
// it whenever we receive a committed entry via entryc, and making it available
// for queries.
//
// The etcd storage backend is given a reference to this instance in the form of
// a ConsistentIndexGetter interface. In addition, it tracks its own view of the
// consistent index in a special bucket+key. See package etcd/mvcc, type
// consistentWatchableStore, method consistentIndex.
//
// Whenever a user makes an e.g. Put request, these values are compared. If
// there is some inconsistency, the transaction is marked as "skip" and becomes
// a no-op. This happens transparently to the user. See package etcd/mvcc,
// type consistentWatchableStore, method TxnBegin.
//
// tl;dr: (ಠ_ಠ)
type consistentIndex struct{ i uint64 }

func (i *consistentIndex) ConsistentIndex() uint64 { return i.i }
func (i *consistentIndex) set(index uint64)        { i.i = index }

func makeIDGen() <-chan uint64 {
	c := make(chan uint64)
	go func() {
		var i uint64 = 1
		for {
			c <- i
			i++
		}
	}()
	return c
}

const (
	maxRequestBytes = 8192
)

var (
	errStopped  = errors.New("etcd store was stopped")
	errTooBig   = errors.New("request too large to send")
	errCanceled = errors.New("request canceled")
)

type responseChans struct {
	msgc chan<- proto.Message
	errc chan<- error
}

func (s *etcdStore) loop() {
	defer close(s.terminatedc)
	defer s.removeDB()

	for {
		select {
		case snapshot := <-s.snapshotc:
			if err := s.applySnapshot(snapshot); err != nil {
				s.logger.Printf("etcd store: apply snapshot: %v", err)
			}

		case entry := <-s.entryc:
			if err := s.applyCommittedEntry(entry); err != nil {
				s.logger.Printf("etcd store: apply committed entry: %v", err)
			}

		case f := <-s.actionc:
			f()

		case <-s.quitc:
			return
		}
	}
}

func (s *etcdStore) stop() {
	close(s.quitc)
	<-s.terminatedc
}

func (s *etcdStore) applySnapshot(snapshot raftpb.Snapshot) error {
	if len(snapshot.Data) == 0 {
		//s.logger.Printf("etcd store: apply snapshot with empty snapshot; skipping")
		return nil
	}

	s.logger.Printf("etcd store: applying snapshot: size %d", len(snapshot.Data))
	s.logger.Printf("etcd store: applying snapshot: metadata %s", snapshot.Metadata.String())
	s.logger.Printf("etcd store: applying snapshot: TODO") // TODO(pb)

	return nil
}

func (s *etcdStore) applyCommittedEntry(entry raftpb.Entry) error {
	// Set the consistent index regardless of the outcome. Because we need to do
	// this for all committed entries, we need to receive all committed entries,
	// and must therefore take responsibility to demux the conf changes to the
	// configurator via confentryc.
	//
	// This requirement is unique to the etcd store. But for symmetry, we assign
	// the same responsibility to the simple store.
	s.index.set(entry.Index)

	switch entry.Type {
	case raftpb.EntryNormal:
		break
	case raftpb.EntryConfChange:
		s.logger.Printf("etcd store: forwarding ConfChange entry")
		s.confentryc <- entry
		return nil
	default:
		s.logger.Printf("etcd store: got unknown entry type %s", entry.Type)
		return fmt.Errorf("unknown entry type %d", entry.Type)
	}

	// entry.Size can be nonzero when len(entry.Data) == 0
	if len(entry.Data) <= 0 {
		s.logger.Printf("etcd store: got empty committed entry (term %d, index %d, type %s); skipping", entry.Term, entry.Index, entry.Type)
		return nil
	}

	var req etcdserverpb.InternalRaftRequest
	if err := req.Unmarshal(entry.Data); err != nil {
		s.logger.Printf("etcd store: unmarshaling entry data: %v", err)
		return err
	}

	msg, err := s.applyInternalRaftRequest(req)
	if err != nil {
		s.logger.Printf("etcd store: applying internal Raft request %d: %v", req.ID, err)
		s.cancelPending(req.ID, err)
		return err
	}

	s.signalPending(req.ID, msg)

	return nil
}

// From public API method to proposalc.
func (s *etcdStore) proposeInternalRaftRequest(req etcdserverpb.InternalRaftRequest) (<-chan proto.Message, <-chan error, error) {
	data, err := req.Marshal()
	if err != nil {
		return nil, nil, err
	}
	if len(data) > maxRequestBytes {
		return nil, nil, errTooBig
	}
	msgc, errc, err := s.registerPending(req.ID)
	if err != nil {
		return nil, nil, err
	}
	s.proposalc <- data
	return msgc, errc, nil
}

func (s *etcdStore) cancelInternalRaftRequest(req etcdserverpb.InternalRaftRequest) {
	s.cancelPending(req.ID, errCanceled)
}

// From committed entryc, back to public API method.
// etcdserver/v3demo_server.go applyV3Result
func (s *etcdStore) applyInternalRaftRequest(req etcdserverpb.InternalRaftRequest) (proto.Message, error) {
	switch {
	case req.Range != nil:
		return applyRange(noTxn, s.kv, req.Range)
	case req.Put != nil:
		return applyPut(noTxn, s.kv, s.lessor, req.Put)
	case req.DeleteRange != nil:
		return applyDeleteRange(noTxn, s.kv, req.DeleteRange)
	case req.Txn != nil:
		return applyTransaction(s.kv, s.lessor, req.Txn)
	case req.Compaction != nil:
		return applyCompaction(s.kv, req.Compaction)
	case req.LeaseGrant != nil:
		return applyLeaseGrant(s.lessor, req.LeaseGrant)
	case req.LeaseRevoke != nil:
		return applyLeaseRevoke(s.lessor, req.LeaseRevoke)
	default:
		return nil, fmt.Errorf("internal Raft request type not implemented")
	}
}

func (s *etcdStore) registerPending(id uint64) (<-chan proto.Message, <-chan error, error) {
	if _, ok := s.pending[id]; ok {
		return nil, nil, fmt.Errorf("pending ID %d already registered", id)
	}
	msgc := make(chan proto.Message)
	errc := make(chan error)
	s.pending[id] = responseChans{msgc, errc}
	return msgc, errc, nil
}

func (s *etcdStore) signalPending(id uint64, msg proto.Message) {
	rc, ok := s.pending[id]
	if !ok {
		// InternalRaftRequests are replicated via Raft. So all peers will
		// invoke this method for all messages on commit. But only the peer that
		// serviced the API request will have an operating pending. So, this is
		// a normal "failure" mode.
		return
	}
	rc.msgc <- msg
	delete(s.pending, id)
}

func (s *etcdStore) cancelPending(id uint64, err error) {
	rc, ok := s.pending[id]
	if !ok {
		s.logger.Printf("etcd store: cancel pending ID %d, but nothing was pending; strange", id)
		return
	}
	rc.errc <- err
	delete(s.pending, id)
}

func (s *etcdStore) removeDB() {
	s.logger.Printf("etcd store: removing tmp DB %s", s.dbPath)
	if err := os.RemoveAll(s.dbPath); err != nil {
		s.logger.Printf("etcd store: removing tmp DB %s: %v", s.dbPath, err)
	}
}

// Sentinel value to indicate the operation is not part of a transaction.
const noTxn = -1

// isGteRange determines if the range end is a >= range. This works around grpc
// sending empty byte strings as nil; >= is encoded in the range end as '\0'.
func isGteRange(rangeEnd []byte) bool {
	return len(rangeEnd) == 1 && rangeEnd[0] == 0
}

func applyRange(txnID int64, kv mvcc.KV, r *etcdserverpb.RangeRequest) (*etcdserverpb.RangeResponse, error) {
	resp := &etcdserverpb.RangeResponse{}
	resp.Header = &etcdserverpb.ResponseHeader{}

	var (
		kvs []mvccpb.KeyValue
		rev int64
		err error
	)

	if isGteRange(r.RangeEnd) {
		r.RangeEnd = []byte{}
	}

	limit := r.Limit
	if r.SortOrder != etcdserverpb.RangeRequest_NONE {
		// fetch everything; sort and truncate afterwards
		limit = 0
	}
	if limit > 0 {
		// fetch one extra for 'more' flag
		limit = limit + 1
	}

	if txnID != noTxn {
		kvs, rev, err = kv.TxnRange(txnID, r.Key, r.RangeEnd, limit, r.Revision)
		if err != nil {
			return nil, err
		}
	} else {
		kvs, rev, err = kv.Range(r.Key, r.RangeEnd, limit, r.Revision)
		if err != nil {
			return nil, err
		}
	}

	if r.SortOrder != etcdserverpb.RangeRequest_NONE {
		var sorter sort.Interface
		switch {
		case r.SortTarget == etcdserverpb.RangeRequest_KEY:
			sorter = &kvSortByKey{&kvSort{kvs}}
		case r.SortTarget == etcdserverpb.RangeRequest_VERSION:
			sorter = &kvSortByVersion{&kvSort{kvs}}
		case r.SortTarget == etcdserverpb.RangeRequest_CREATE:
			sorter = &kvSortByCreate{&kvSort{kvs}}
		case r.SortTarget == etcdserverpb.RangeRequest_MOD:
			sorter = &kvSortByMod{&kvSort{kvs}}
		case r.SortTarget == etcdserverpb.RangeRequest_VALUE:
			sorter = &kvSortByValue{&kvSort{kvs}}
		}
		switch {
		case r.SortOrder == etcdserverpb.RangeRequest_ASCEND:
			sort.Sort(sorter)
		case r.SortOrder == etcdserverpb.RangeRequest_DESCEND:
			sort.Sort(sort.Reverse(sorter))
		}
	}

	if r.Limit > 0 && len(kvs) > int(r.Limit) {
		kvs = kvs[:r.Limit]
		resp.More = true
	}

	resp.Header.Revision = rev
	for i := range kvs {
		resp.Kvs = append(resp.Kvs, &kvs[i])
	}
	return resp, nil
}

type kvSort struct{ kvs []mvccpb.KeyValue }

func (s *kvSort) Swap(i, j int) {
	t := s.kvs[i]
	s.kvs[i] = s.kvs[j]
	s.kvs[j] = t
}
func (s *kvSort) Len() int { return len(s.kvs) }

type kvSortByKey struct{ *kvSort }

func (s *kvSortByKey) Less(i, j int) bool {
	return bytes.Compare(s.kvs[i].Key, s.kvs[j].Key) < 0
}

type kvSortByVersion struct{ *kvSort }

func (s *kvSortByVersion) Less(i, j int) bool {
	return (s.kvs[i].Version - s.kvs[j].Version) < 0
}

type kvSortByCreate struct{ *kvSort }

func (s *kvSortByCreate) Less(i, j int) bool {
	return (s.kvs[i].CreateRevision - s.kvs[j].CreateRevision) < 0
}

type kvSortByMod struct{ *kvSort }

func (s *kvSortByMod) Less(i, j int) bool {
	return (s.kvs[i].ModRevision - s.kvs[j].ModRevision) < 0
}

type kvSortByValue struct{ *kvSort }

func (s *kvSortByValue) Less(i, j int) bool {
	return bytes.Compare(s.kvs[i].Value, s.kvs[j].Value) < 0
}

func applyPut(txnID int64, kv mvcc.KV, lessor lease.Lessor, req *etcdserverpb.PutRequest) (*etcdserverpb.PutResponse, error) {
	resp := &etcdserverpb.PutResponse{}
	resp.Header = &etcdserverpb.ResponseHeader{}
	var (
		rev int64
		err error
	)
	if txnID != noTxn {
		rev, err = kv.TxnPut(txnID, req.Key, req.Value, lease.LeaseID(req.Lease))
		if err != nil {
			return nil, err
		}
	} else {
		leaseID := lease.LeaseID(req.Lease)
		if leaseID != lease.NoLease {
			if l := lessor.Lookup(leaseID); l == nil {
				return nil, lease.ErrLeaseNotFound
			}
		}
		rev = kv.Put(req.Key, req.Value, leaseID)
	}
	resp.Header.Revision = rev
	return resp, nil
}

func applyDeleteRange(txnID int64, kv mvcc.KV, req *etcdserverpb.DeleteRangeRequest) (*etcdserverpb.DeleteRangeResponse, error) {
	resp := &etcdserverpb.DeleteRangeResponse{}
	resp.Header = &etcdserverpb.ResponseHeader{}

	var (
		n   int64
		rev int64
		err error
	)

	if isGteRange(req.RangeEnd) {
		req.RangeEnd = []byte{}
	}

	if txnID != noTxn {
		n, rev, err = kv.TxnDeleteRange(txnID, req.Key, req.RangeEnd)
		if err != nil {
			return nil, err
		}
	} else {
		n, rev = kv.DeleteRange(req.Key, req.RangeEnd)
	}

	resp.Deleted = n
	resp.Header.Revision = rev
	return resp, nil
}

func applyTransaction(kv mvcc.KV, lessor lease.Lessor, req *etcdserverpb.TxnRequest) (*etcdserverpb.TxnResponse, error) {
	var revision int64

	ok := true
	for _, c := range req.Compare {
		if revision, ok = applyCompare(kv, c); !ok {
			break
		}
	}

	var reqs []*etcdserverpb.RequestUnion
	if ok {
		reqs = req.Success
	} else {
		reqs = req.Failure
	}

	if err := checkRequestLeases(lessor, reqs); err != nil {
		return nil, err
	}
	if err := checkRequestRange(kv, reqs); err != nil {
		return nil, err
	}

	// When executing the operations of txn, we need to hold the txn lock.
	// So the reader will not see any intermediate results.
	txnID := kv.TxnBegin()
	defer func() {
		err := kv.TxnEnd(txnID)
		if err != nil {
			panic(fmt.Sprint("unexpected error when closing txn", txnID))
		}
	}()

	resps := make([]*etcdserverpb.ResponseUnion, len(reqs))
	for i := range reqs {
		resps[i] = applyUnion(txnID, kv, reqs[i])
	}

	if len(resps) != 0 {
		revision++
	}

	txnResp := &etcdserverpb.TxnResponse{}
	txnResp.Header = &etcdserverpb.ResponseHeader{}
	txnResp.Header.Revision = revision
	txnResp.Responses = resps
	txnResp.Succeeded = ok
	return txnResp, nil
}

func checkRequestLeases(le lease.Lessor, reqs []*etcdserverpb.RequestUnion) error {
	for _, requ := range reqs {
		tv, ok := requ.Request.(*etcdserverpb.RequestUnion_RequestPut)
		if !ok {
			continue
		}
		preq := tv.RequestPut
		if preq == nil || lease.LeaseID(preq.Lease) == lease.NoLease {
			continue
		}
		if l := le.Lookup(lease.LeaseID(preq.Lease)); l == nil {
			return lease.ErrLeaseNotFound
		}
	}
	return nil
}

func checkRequestRange(kv mvcc.KV, reqs []*etcdserverpb.RequestUnion) error {
	for _, requ := range reqs {
		tv, ok := requ.Request.(*etcdserverpb.RequestUnion_RequestRange)
		if !ok {
			continue
		}
		greq := tv.RequestRange
		if greq == nil || greq.Revision == 0 {
			continue
		}

		if greq.Revision > kv.Rev() {
			return mvcc.ErrFutureRev
		}
		if greq.Revision < kv.FirstRev() {
			return mvcc.ErrCompacted
		}
	}
	return nil
}

func applyUnion(txnID int64, kv mvcc.KV, union *etcdserverpb.RequestUnion) *etcdserverpb.ResponseUnion {
	switch tv := union.Request.(type) {
	case *etcdserverpb.RequestUnion_RequestRange:
		if tv.RequestRange != nil {
			resp, err := applyRange(txnID, kv, tv.RequestRange)
			if err != nil {
				panic("unexpected error during txn")
			}
			return &etcdserverpb.ResponseUnion{Response: &etcdserverpb.ResponseUnion_ResponseRange{ResponseRange: resp}}
		}
	case *etcdserverpb.RequestUnion_RequestPut:
		if tv.RequestPut != nil {
			resp, err := applyPut(txnID, kv, nil, tv.RequestPut)
			if err != nil {
				panic("unexpected error during txn")
			}
			return &etcdserverpb.ResponseUnion{Response: &etcdserverpb.ResponseUnion_ResponsePut{ResponsePut: resp}}
		}
	case *etcdserverpb.RequestUnion_RequestDeleteRange:
		if tv.RequestDeleteRange != nil {
			resp, err := applyDeleteRange(txnID, kv, tv.RequestDeleteRange)
			if err != nil {
				panic("unexpected error during txn")
			}
			return &etcdserverpb.ResponseUnion{Response: &etcdserverpb.ResponseUnion_ResponseDeleteRange{ResponseDeleteRange: resp}}
		}
	default:
		// empty union
		return nil
	}
	return nil
}

// applyCompare applies the compare request.
// It returns the revision at which the comparison happens. If the comparison
// succeeds, the it returns true. Otherwise it returns false.
func applyCompare(kv mvcc.KV, c *etcdserverpb.Compare) (int64, bool) {
	ckvs, rev, err := kv.Range(c.Key, nil, 1, 0)
	if err != nil {
		if err == mvcc.ErrTxnIDMismatch {
			panic("unexpected txn ID mismatch error")
		}
		return rev, false
	}
	var ckv mvccpb.KeyValue
	if len(ckvs) != 0 {
		ckv = ckvs[0]
	} else {
		// Use the zero value of ckv normally. However...
		if c.Target == etcdserverpb.Compare_VALUE {
			// Always fail if we're comparing a value on a key that doesn't exist.
			// We can treat non-existence as the empty set explicitly, such that
			// even a key with a value of length 0 bytes is still a real key
			// that was written that way
			return rev, false
		}
	}

	// -1 is less, 0 is equal, 1 is greater
	var result int
	switch c.Target {
	case etcdserverpb.Compare_VALUE:
		tv, _ := c.TargetUnion.(*etcdserverpb.Compare_Value)
		if tv != nil {
			result = bytes.Compare(ckv.Value, tv.Value)
		}
	case etcdserverpb.Compare_CREATE:
		tv, _ := c.TargetUnion.(*etcdserverpb.Compare_CreateRevision)
		if tv != nil {
			result = compareInt64(ckv.CreateRevision, tv.CreateRevision)
		}

	case etcdserverpb.Compare_MOD:
		tv, _ := c.TargetUnion.(*etcdserverpb.Compare_ModRevision)
		if tv != nil {
			result = compareInt64(ckv.ModRevision, tv.ModRevision)
		}
	case etcdserverpb.Compare_VERSION:
		tv, _ := c.TargetUnion.(*etcdserverpb.Compare_Version)
		if tv != nil {
			result = compareInt64(ckv.Version, tv.Version)
		}
	}

	switch c.Result {
	case etcdserverpb.Compare_EQUAL:
		if result != 0 {
			return rev, false
		}
	case etcdserverpb.Compare_GREATER:
		if result != 1 {
			return rev, false
		}
	case etcdserverpb.Compare_LESS:
		if result != -1 {
			return rev, false
		}
	}
	return rev, true
}

func compareInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func applyCompaction(kv mvcc.KV, req *etcdserverpb.CompactionRequest) (*etcdserverpb.CompactionResponse, error) {
	resp := &etcdserverpb.CompactionResponse{}
	resp.Header = &etcdserverpb.ResponseHeader{}
	_, err := kv.Compact(req.Revision)
	if err != nil {
		return nil, err
	}
	// get the current revision. which key to get is not important.
	_, resp.Header.Revision, _ = kv.Range([]byte("compaction"), nil, 1, 0)
	return resp, err
}

func applyLeaseGrant(lessor lease.Lessor, req *etcdserverpb.LeaseGrantRequest) (*etcdserverpb.LeaseGrantResponse, error) {
	l, err := lessor.Grant(lease.LeaseID(req.ID), req.TTL)
	resp := &etcdserverpb.LeaseGrantResponse{}
	if err == nil {
		resp.ID = int64(l.ID)
		resp.TTL = l.TTL
	}
	return resp, err
}

func applyLeaseRevoke(lessor lease.Lessor, req *etcdserverpb.LeaseRevokeRequest) (*etcdserverpb.LeaseRevokeResponse, error) {
	err := lessor.Revoke(lease.LeaseID(req.ID))
	return &etcdserverpb.LeaseRevokeResponse{}, err
}
