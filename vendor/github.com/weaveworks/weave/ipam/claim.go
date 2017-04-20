package ipam

import (
	"fmt"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/api"
	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
)

type claim struct {
	resultChan       chan<- error
	ident            string       // a container ID, something like "weave:expose", or api.NoContainerID
	cidr             address.CIDR // single address being claimed
	isContainer      bool         // true if ident is a container ID
	noErrorOnUnknown bool         // if false, error or block if we don't know; if true return ok but keep trying
	hasBeenCancelled func() bool
}

// Send an error (or nil for success) back to caller listening on resultChan
func (c *claim) sendResult(result error) {
	// Make sure we only send a result once, since listener stops listening after that
	if c.resultChan != nil {
		c.resultChan <- result
		close(c.resultChan)
		c.resultChan = nil
		return
	}
	if result != nil {
		common.Log.Errorln("[allocator] " + result.Error())
	}
}

// Try returns true for success (or failure), false if we need to try again later
func (c *claim) Try(alloc *Allocator) bool {
	if c.hasBeenCancelled != nil && c.hasBeenCancelled() {
		c.Cancel()
		return true
	}

	if !alloc.ring.Contains(c.cidr.Addr) {
		// Address not within our universe; assume user knows what they are doing
		alloc.infof("Address %s claimed by %s - not in our range", c.cidr, c.ident)
		alloc.addOwned(c.ident, c.cidr, c.isContainer)
		c.sendResult(nil)
		return true
	}

	alloc.establishRing()

	// If we had heard that this container died, resurrect it
	delete(alloc.dead, c.ident) // (delete is no-op if key not in map)

	switch owner := alloc.ring.Owner(c.cidr.Addr); owner {
	case alloc.ourName:
		// success
	case mesh.UnknownPeerName:
		// If our ring doesn't know, it must be empty.
		alloc.infof("Claim %s for %s: is in the range %s, but the allocator is not initialized yet; will try later.",
			c.cidr, c.ident, alloc.universe)
		if c.noErrorOnUnknown {
			c.sendResult(nil)
		}
		return false
	default:
		alloc.debugf("requesting address %s from other peer %s", c.cidr, owner)
		err := alloc.sendSpaceRequest(owner, address.NewRange(c.cidr.Addr, 1))
		if err != nil { // can't speak to owner right now
			if c.noErrorOnUnknown {
				alloc.infof("Claim %s for %s: %s; will try later.", c.cidr, c.ident, err)
				c.sendResult(nil)
			} else { // just tell the user they can't do this.
				c.deniedBy(alloc, owner)
			}
		}
		return false
	}

	// We are the owner, check we haven't given it to another container
	existingIdent := alloc.findOwner(c.cidr.Addr)
	switch {
	case existingIdent == "":
		// Unused address, we try to claim it:
		if err := alloc.space.Claim(c.cidr.Addr); err == nil {
			alloc.debugln("Claimed", c.cidr, "for", c.ident)
			if c.ident == api.NoContainerID {
				alloc.addOwned(c.cidr.Addr.String(), c.cidr, c.isContainer)
			} else {
				alloc.addOwned(c.ident, c.cidr, c.isContainer)
			}
			c.sendResult(nil)
		} else {
			c.sendResult(err)
		}
	case (existingIdent == c.ident) || (c.ident == api.NoContainerID && existingIdent == c.cidr.Addr.String()):
		// same identifier is claiming same address; that's OK
		alloc.debugln("Re-Claimed", c.cidr, "for", c.ident)
		c.sendResult(nil)
	case existingIdent == c.cidr.Addr.String():
		// Address already allocated via api.NoContainerID name and current ID is a real container ID:
		c.sendResult(fmt.Errorf("address %s already in use", c.cidr))
	case c.ident == api.NoContainerID:
		// We do not know whether this is the same container or another one,
		// but we also cannot prove otherwise, so we let it reclaim the address:
		alloc.debugln("Re-Claimed", c.cidr, "for ID", c.ident, "having existing ID as", existingIdent)
		c.sendResult(nil)
	default:
		// Addr already owned by container on this machine
		c.sendResult(fmt.Errorf("address %s is already owned by %s", c.cidr.String(), existingIdent))
	}
	return true
}

func (c *claim) deniedBy(alloc *Allocator, owner mesh.PeerName) {
	name, found := alloc.nicknames[owner]
	if found {
		name = " (" + name + ")"
	}
	c.sendResult(fmt.Errorf("address %s is owned by other peer %s%s", c.cidr.String(), owner, name))
}

func (c *claim) Cancel() {
	c.sendResult(&errorCancelled{"Claim", c.ident})
}

func (c *claim) ForContainer(ident string) bool {
	return c.ident == ident
}
