package router

import (
	"time"

	"github.com/weaveworks/mesh"
)

type NetworkRouterStatus struct {
	*mesh.Status
	Interface    string
	CaptureStats map[string]int
	MACs         []MACStatus
}

type MACStatus struct {
	Mac      string
	Name     string
	NickName string
	LastSeen time.Time
}

func NewNetworkRouterStatus(router *NetworkRouter) *NetworkRouterStatus {
	return &NetworkRouterStatus{
		mesh.NewStatus(router.Router),
		router.Bridge.String(),
		router.Bridge.Stats(),
		NewMACStatusSlice(router.Macs)}
}

func NewMACStatusSlice(cache *MacCache) []MACStatus {
	cache.RLock()
	defer cache.RUnlock()

	var slice []MACStatus
	for key, entry := range cache.table {
		slice = append(slice, MACStatus{
			intmac(key).String(),
			entry.peer.Name.String(),
			entry.peer.NickName,
			entry.lastSeen})
	}

	return slice
}
