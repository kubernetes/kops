package hcloud

import (
	"fmt"
	"net"
)

// ErrorCode represents an error code returned from the API.
type ErrorCode string

// Error codes returned from the API.
const (
	ErrorCodeServiceError          ErrorCode = "service_error"           // Generic service error
	ErrorCodeRateLimitExceeded     ErrorCode = "rate_limit_exceeded"     // Rate limit exceeded
	ErrorCodeUnknownError          ErrorCode = "unknown_error"           // Unknown error
	ErrorCodeNotFound              ErrorCode = "not_found"               // Resource not found
	ErrorCodeInvalidInput          ErrorCode = "invalid_input"           // Validation error
	ErrorCodeForbidden             ErrorCode = "forbidden"               // Insufficient permissions
	ErrorCodeUnauthorized          ErrorCode = "unauthorized"            // Request was made with an invalid or unknown token
	ErrorCodeJSONError             ErrorCode = "json_error"              // Invalid JSON in request
	ErrorCodeLocked                ErrorCode = "locked"                  // Item is locked (Another action is running)
	ErrorCodeResourceLimitExceeded ErrorCode = "resource_limit_exceeded" // Resource limit exceeded
	ErrorCodeResourceUnavailable   ErrorCode = "resource_unavailable"    // Resource currently unavailable
	ErrorCodeUniquenessError       ErrorCode = "uniqueness_error"        // One or more fields must be unique
	ErrorCodeProtected             ErrorCode = "protected"               // The actions you are trying is protected
	ErrorCodeMaintenance           ErrorCode = "maintenance"             // Cannot perform operation due to maintenance
	ErrorCodeConflict              ErrorCode = "conflict"                // The resource has changed during the request, please retry
	ErrorCodeRobotUnavailable      ErrorCode = "robot_unavailable"       // Robot was not available. The caller may retry the operation after a short delay
	ErrorCodeResourceLocked        ErrorCode = "resource_locked"         // The resource is locked. The caller should contact support
	ErrorUnsupportedError          ErrorCode = "unsupported_error"       // The given resource does not support this

	// Server related error codes.

	ErrorCodeInvalidServerType     ErrorCode = "invalid_server_type"     // The server type does not fit for the given server or is deprecated
	ErrorCodeServerNotStopped      ErrorCode = "server_not_stopped"      // The action requires a stopped server
	ErrorCodeNetworksOverlap       ErrorCode = "networks_overlap"        // The network IP range overlaps with one of the server networks
	ErrorCodePlacementError        ErrorCode = "placement_error"         // An error during the placement occurred
	ErrorCodeServerAlreadyAttached ErrorCode = "server_already_attached" // The server is already attached to the resource

	// Load Balancer related error codes.

	ErrorCodeIPNotOwned                       ErrorCode = "ip_not_owned"                          // The IP you are trying to add as a target is not owned by the Project owner
	ErrorCodeSourcePortAlreadyUsed            ErrorCode = "source_port_already_used"              // The source port you are trying to add is already in use
	ErrorCodeCloudResourceIPNotAllowed        ErrorCode = "cloud_resource_ip_not_allowed"         // The IP you are trying to add as a target belongs to a Hetzner Cloud resource
	ErrorCodeServerNotAttachedToNetwork       ErrorCode = "server_not_attached_to_network"        // The server you are trying to add as a target is not attached to the same network as the Load Balancer
	ErrorCodeTargetAlreadyDefined             ErrorCode = "target_already_defined"                // The Load Balancer target you are trying to define is already defined
	ErrorCodeInvalidLoadBalancerType          ErrorCode = "invalid_load_balancer_type"            // The Load Balancer type does not fit for the given Load Balancer
	ErrorCodeLoadBalancerAlreadyAttached      ErrorCode = "load_balancer_already_attached"        // The Load Balancer is already attached to a network
	ErrorCodeTargetsWithoutUsePrivateIP       ErrorCode = "targets_without_use_private_ip"        // The Load Balancer has targets that use the public IP instead of the private IP
	ErrorCodeLoadBalancerNotAttachedToNetwork ErrorCode = "load_balancer_not_attached_to_network" // The Load Balancer is not attached to a network

	// Network related error codes.

	ErrorCodeIPNotAvailable     ErrorCode = "ip_not_available"        // The provided Network IP is not available
	ErrorCodeNoSubnetAvailable  ErrorCode = "no_subnet_available"     // No Subnet or IP is available for the Load Balancer/Server within the network
	ErrorCodeVSwitchAlreadyUsed ErrorCode = "vswitch_id_already_used" // The given Robot vSwitch ID is already registered in another network

	// Volume related error codes.

	ErrorCodeNoSpaceLeftInLocation ErrorCode = "no_space_left_in_location" // There is no volume space left in the given location
	ErrorCodeVolumeAlreadyAttached ErrorCode = "volume_already_attached"   // Volume is already attached to a server, detach first

	// Firewall related error codes.

	ErrorCodeFirewallAlreadyApplied   ErrorCode = "firewall_already_applied"    // Firewall was already applied on resource
	ErrorCodeFirewallAlreadyRemoved   ErrorCode = "firewall_already_removed"    // Firewall was already removed from the resource
	ErrorCodeIncompatibleNetworkType  ErrorCode = "incompatible_network_type"   // The Network type is incompatible for the given resource
	ErrorCodeResourceInUse            ErrorCode = "resource_in_use"             // Firewall must not be in use to be deleted
	ErrorCodeServerAlreadyAdded       ErrorCode = "server_already_added"        // Server added more than one time to resource
	ErrorCodeFirewallResourceNotFound ErrorCode = "firewall_resource_not_found" // Resource a firewall should be attached to / detached from not found

	// Certificate related error codes.

	ErrorCodeCAARecordDoesNotAllowCA                        ErrorCode = "caa_record_does_not_allow_ca"                          // CAA record does not allow certificate authority
	ErrorCodeCADNSValidationFailed                          ErrorCode = "ca_dns_validation_failed"                              // Certificate Authority: DNS validation failed
	ErrorCodeCATooManyAuthorizationsFailedRecently          ErrorCode = "ca_too_many_authorizations_failed_recently"            // Certificate Authority: Too many authorizations failed recently
	ErrorCodeCATooManyCertificatedIssuedForRegisteredDomain ErrorCode = "ca_too_many_certificates_issued_for_registered_domain" // Certificate Authority: Too many certificates issued for registered domain
	ErrorCodeCATooManyDuplicateCertificates                 ErrorCode = "ca_too_many_duplicate_certificates"                    // Certificate Authority: Too many duplicate certificates
	ErrorCodeCloudNotVerifyDomainDelegatedToZone            ErrorCode = "could_not_verify_domain_delegated_to_zone"             // Could not verify domain delegated to zone
	ErrorCodeDNSZoneNotFound                                ErrorCode = "dns_zone_not_found"                                    // DNS zone not found

	// Deprecated error codes.

	// Deprecated: The actual value of this error code is limit_reached. The
	// new error code rate_limit_exceeded for rate limiting was introduced
	// before Hetzner Cloud launched into the public. To make clients using the
	// old error code still work as expected, we set the value of the old error
	// code to that of the new error code.
	ErrorCodeLimitReached = ErrorCodeRateLimitExceeded
)

// Error is an error returned from the API.
type Error struct {
	Code    ErrorCode
	Message string
	Details interface{}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

// ErrorDetailsInvalidInput contains the details of an 'invalid_input' error.
type ErrorDetailsInvalidInput struct {
	Fields []ErrorDetailsInvalidInputField
}

// ErrorDetailsInvalidInputField contains the validation errors reported on a field.
type ErrorDetailsInvalidInputField struct {
	Name     string
	Messages []string
}

// IsError returns whether err is an API error with the given error code.
func IsError(err error, code ErrorCode) bool {
	apiErr, ok := err.(Error)
	return ok && apiErr.Code == code
}

type InvalidIPError struct {
	IP string
}

func (e InvalidIPError) Error() string {
	return fmt.Sprintf("could not parse ip address %s", e.IP)
}

type DNSNotFoundError struct {
	IP net.IP
}

func (e DNSNotFoundError) Error() string {
	return fmt.Sprintf("dns for ip %s not found", e.IP.String())
}
