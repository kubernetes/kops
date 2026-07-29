package hcloud

import "time"

// Deprecatable is a shared interface implemented by all Resources that have a defined deprecation workflow.
type Deprecatable interface {
	// IsDeprecated returns true if the resource is marked as deprecated.
	IsDeprecated() bool

	// UnavailableAfter returns the time that the deprecated resource will be removed from the API.
	// This only returns a valid value if [Deprecatable.IsDeprecated] returned true.
	UnavailableAfter() time.Time

	// DeprecationAnnounced returns the time that the deprecation of this resource was announced.
	// This only returns a valid value if [Deprecatable.IsDeprecated] returned true.
	DeprecationAnnounced() time.Time
}

// DeprecationInfo contains the information published when a resource is actually deprecated.
type DeprecationInfo struct {
	Announced        time.Time
	UnavailableAfter time.Time
}

// DeprecatableResource implements the [Deprecatable] interface and can be embedded in structs for Resources that can
// be deprecated.
type DeprecatableResource struct {
	Deprecation *DeprecationInfo
}

// IsDeprecated returns true if the resource is marked as deprecated.
func (o DeprecatableResource) IsDeprecated() bool {
	return o.Deprecation != nil
}

// UnavailableAfter returns the time that the deprecated resource will be removed from the API.
// This only returns a valid value if [Deprecatable.IsDeprecated] returned true.
func (o DeprecatableResource) UnavailableAfter() time.Time {
	if !o.IsDeprecated() {
		// Return "null" time if resource is not deprecated
		return time.Unix(0, 0)
	}

	return o.Deprecation.UnavailableAfter
}

// DeprecationAnnounced returns the time that the deprecation of this resource was announced.
// This only returns a valid value if [Deprecatable.IsDeprecated] returned true.
func (o DeprecatableResource) DeprecationAnnounced() time.Time {
	if !o.IsDeprecated() {
		// Return "null" time if resource is not deprecated
		return time.Unix(0, 0)
	}

	return o.Deprecation.Announced
}

// Make sure that all expected Resources actually implement the interface.
var _ Deprecatable = ServerType{}
var _ Deprecatable = LoadBalancerType{}
