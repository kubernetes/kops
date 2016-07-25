package dns

// Context represents a state of the world for DNS.
// It is grouped by scopes & named keys, and controllers will replace those groups
// The DNS controller will then merge all those record sets, resolve aliases etc,
// and then call into a dns backend to match the desired state of the world.
type Context interface {
	// CreateScope creates a new scope, which holds a set of records.
	// MarkReady must be called on every scope before any changes will be applied.
	// Records from all the scopes will be merged
	CreateScope(name string) (Scope, error)
}

type Scope interface {
	// Replace sets the records for recordName to the provided set of records.
	Replace(recordName string, records []Record)

	// MarkReady should be called when a controller has populated all the records for a particular scope
	MarkReady()
}
