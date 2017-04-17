package provider

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type resourceRecordChangeset struct {
	resourceRecordSets *resourceRecordSets
	zone               *zone
	add                []dnsprovider.ResourceRecordSet
	remove             []dnsprovider.ResourceRecordSet
	upsert             []dnsprovider.ResourceRecordSet
}

var _ dnsprovider.ResourceRecordChangeset = &resourceRecordChangeset{}

// Add adds the creation of a ResourceRecordSet in the Zone to the changeset
func (c *resourceRecordChangeset) Add(rrs dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.add = append(c.add, rrs)
	return c
}

// Remove adds the removal of a ResourceRecordSet in the Zone to the changeset
// The supplied ResourceRecordSet must match one of the existing recordsets (obtained via List()) exactly.
func (c *resourceRecordChangeset) Remove(rrs dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.remove = append(c.remove, rrs)
	return c
}

// Upsert adds an "create or update" operation for the ResourceRecordSet in the Zone to the changeset
// Note: the implementation may translate this into a Remove followed by an Add operation.
// If you have the pre-image, it will likely be more efficient to call Remove and Add.
func (c *resourceRecordChangeset) Upsert(rrs dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.upsert = append(c.upsert, rrs)
	return c
}

// Apply applies the accumulated operations to the Zone.
func (c *resourceRecordChangeset) Apply() error {
	return c.zone.applyChangeset(c)
}

// IsEmpty returns true if there are no accumulated operations.
func (c *resourceRecordChangeset) IsEmpty() bool {
	return len(c.add) == 0 && len(c.remove) == 0 && len(c.upsert) == 0
}

// ResourceRecordSets returns the parent ResourceRecordSets
func (c *resourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return c.resourceRecordSets
}
