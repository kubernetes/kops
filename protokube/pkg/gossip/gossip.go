package gossip

type GossipStateSnapshot struct {
	Values  map[string]string
	Version uint64
}

type GossipState interface {
	Snapshot() *GossipStateSnapshot
	UpdateValues(removeKeys []string, putKeys map[string]string) error
}
