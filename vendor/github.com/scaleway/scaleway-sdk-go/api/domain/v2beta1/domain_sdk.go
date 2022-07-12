// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package domain provides methods and message types of the domain v2beta1 API.
package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: DNS API
//
// Manage your DNS zones and records.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

// RegistrarAPI: domains registrar API
//
// Manage your domains and contacts.
type RegistrarAPI struct {
	client *scw.Client
}

// NewRegistrarAPI returns a RegistrarAPI object from a Scaleway client.
func NewRegistrarAPI(client *scw.Client) *RegistrarAPI {
	return &RegistrarAPI{
		client: client,
	}
}

type ContactCivility string

const (
	// ContactCivilityCivilityUnknown is [insert doc].
	ContactCivilityCivilityUnknown = ContactCivility("civility_unknown")
	// ContactCivilityMr is [insert doc].
	ContactCivilityMr = ContactCivility("mr")
	// ContactCivilityMrs is [insert doc].
	ContactCivilityMrs = ContactCivility("mrs")
)

func (enum ContactCivility) String() string {
	if enum == "" {
		// return default value if empty
		return "civility_unknown"
	}
	return string(enum)
}

func (enum ContactCivility) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ContactCivility) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ContactCivility(ContactCivility(tmp).String())
	return nil
}

type ContactEmailStatus string

const (
	// ContactEmailStatusEmailStatusUnknown is [insert doc].
	ContactEmailStatusEmailStatusUnknown = ContactEmailStatus("email_status_unknown")
	// ContactEmailStatusValidated is [insert doc].
	ContactEmailStatusValidated = ContactEmailStatus("validated")
	// ContactEmailStatusNotValidated is [insert doc].
	ContactEmailStatusNotValidated = ContactEmailStatus("not_validated")
	// ContactEmailStatusInvalidEmail is [insert doc].
	ContactEmailStatusInvalidEmail = ContactEmailStatus("invalid_email")
)

func (enum ContactEmailStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "email_status_unknown"
	}
	return string(enum)
}

func (enum ContactEmailStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ContactEmailStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ContactEmailStatus(ContactEmailStatus(tmp).String())
	return nil
}

type ContactExtensionFRMode string

const (
	// ContactExtensionFRModeModeUnknown is [insert doc].
	ContactExtensionFRModeModeUnknown = ContactExtensionFRMode("mode_unknown")
	// ContactExtensionFRModeParticular is [insert doc].
	ContactExtensionFRModeParticular = ContactExtensionFRMode("particular")
	// ContactExtensionFRModeCompanyIdentificationCode is [insert doc].
	ContactExtensionFRModeCompanyIdentificationCode = ContactExtensionFRMode("company_identification_code")
	// ContactExtensionFRModeDuns is [insert doc].
	ContactExtensionFRModeDuns = ContactExtensionFRMode("duns")
	// ContactExtensionFRModeLocal is [insert doc].
	ContactExtensionFRModeLocal = ContactExtensionFRMode("local")
	// ContactExtensionFRModeAssociation is [insert doc].
	ContactExtensionFRModeAssociation = ContactExtensionFRMode("association")
	// ContactExtensionFRModeBrand is [insert doc].
	ContactExtensionFRModeBrand = ContactExtensionFRMode("brand")
	// ContactExtensionFRModeCodeAuthAfnic is [insert doc].
	ContactExtensionFRModeCodeAuthAfnic = ContactExtensionFRMode("code_auth_afnic")
)

func (enum ContactExtensionFRMode) String() string {
	if enum == "" {
		// return default value if empty
		return "mode_unknown"
	}
	return string(enum)
}

func (enum ContactExtensionFRMode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ContactExtensionFRMode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ContactExtensionFRMode(ContactExtensionFRMode(tmp).String())
	return nil
}

type ContactLegalForm string

const (
	// ContactLegalFormLegalFormUnknown is [insert doc].
	ContactLegalFormLegalFormUnknown = ContactLegalForm("legal_form_unknown")
	// ContactLegalFormParticular is [insert doc].
	ContactLegalFormParticular = ContactLegalForm("particular")
	// ContactLegalFormSociety is [insert doc].
	ContactLegalFormSociety = ContactLegalForm("society")
	// ContactLegalFormAssociation is [insert doc].
	ContactLegalFormAssociation = ContactLegalForm("association")
	// ContactLegalFormOther is [insert doc].
	ContactLegalFormOther = ContactLegalForm("other")
)

func (enum ContactLegalForm) String() string {
	if enum == "" {
		// return default value if empty
		return "legal_form_unknown"
	}
	return string(enum)
}

func (enum ContactLegalForm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ContactLegalForm) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ContactLegalForm(ContactLegalForm(tmp).String())
	return nil
}

type DNSZoneStatus string

const (
	// DNSZoneStatusUnknown is [insert doc].
	DNSZoneStatusUnknown = DNSZoneStatus("unknown")
	// DNSZoneStatusActive is [insert doc].
	DNSZoneStatusActive = DNSZoneStatus("active")
	// DNSZoneStatusPending is [insert doc].
	DNSZoneStatusPending = DNSZoneStatus("pending")
	// DNSZoneStatusError is [insert doc].
	DNSZoneStatusError = DNSZoneStatus("error")
	// DNSZoneStatusLocked is [insert doc].
	DNSZoneStatusLocked = DNSZoneStatus("locked")
)

func (enum DNSZoneStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum DNSZoneStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DNSZoneStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DNSZoneStatus(DNSZoneStatus(tmp).String())
	return nil
}

type DSRecordAlgorithm string

const (
	// DSRecordAlgorithmRsamd5 is [insert doc].
	DSRecordAlgorithmRsamd5 = DSRecordAlgorithm("rsamd5")
	// DSRecordAlgorithmDh is [insert doc].
	DSRecordAlgorithmDh = DSRecordAlgorithm("dh")
	// DSRecordAlgorithmDsa is [insert doc].
	DSRecordAlgorithmDsa = DSRecordAlgorithm("dsa")
	// DSRecordAlgorithmRsasha1 is [insert doc].
	DSRecordAlgorithmRsasha1 = DSRecordAlgorithm("rsasha1")
	// DSRecordAlgorithmDsaNsec3Sha1 is [insert doc].
	DSRecordAlgorithmDsaNsec3Sha1 = DSRecordAlgorithm("dsa_nsec3_sha1")
	// DSRecordAlgorithmRsasha1Nsec3Sha1 is [insert doc].
	DSRecordAlgorithmRsasha1Nsec3Sha1 = DSRecordAlgorithm("rsasha1_nsec3_sha1")
	// DSRecordAlgorithmRsasha256 is [insert doc].
	DSRecordAlgorithmRsasha256 = DSRecordAlgorithm("rsasha256")
	// DSRecordAlgorithmRsasha512 is [insert doc].
	DSRecordAlgorithmRsasha512 = DSRecordAlgorithm("rsasha512")
	// DSRecordAlgorithmEccGost is [insert doc].
	DSRecordAlgorithmEccGost = DSRecordAlgorithm("ecc_gost")
	// DSRecordAlgorithmEcdsap256sha256 is [insert doc].
	DSRecordAlgorithmEcdsap256sha256 = DSRecordAlgorithm("ecdsap256sha256")
	// DSRecordAlgorithmEcdsap384sha384 is [insert doc].
	DSRecordAlgorithmEcdsap384sha384 = DSRecordAlgorithm("ecdsap384sha384")
	// DSRecordAlgorithmEd25519 is [insert doc].
	DSRecordAlgorithmEd25519 = DSRecordAlgorithm("ed25519")
	// DSRecordAlgorithmEd448 is [insert doc].
	DSRecordAlgorithmEd448 = DSRecordAlgorithm("ed448")
)

func (enum DSRecordAlgorithm) String() string {
	if enum == "" {
		// return default value if empty
		return "rsamd5"
	}
	return string(enum)
}

func (enum DSRecordAlgorithm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DSRecordAlgorithm) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DSRecordAlgorithm(DSRecordAlgorithm(tmp).String())
	return nil
}

type DSRecordDigestType string

const (
	// DSRecordDigestTypeSha1 is [insert doc].
	DSRecordDigestTypeSha1 = DSRecordDigestType("sha_1")
	// DSRecordDigestTypeSha256 is [insert doc].
	DSRecordDigestTypeSha256 = DSRecordDigestType("sha_256")
	// DSRecordDigestTypeGostR34_11_94 is [insert doc].
	DSRecordDigestTypeGostR34_11_94 = DSRecordDigestType("gost_r_34_11_94")
	// DSRecordDigestTypeSha384 is [insert doc].
	DSRecordDigestTypeSha384 = DSRecordDigestType("sha_384")
)

func (enum DSRecordDigestType) String() string {
	if enum == "" {
		// return default value if empty
		return "sha_1"
	}
	return string(enum)
}

func (enum DSRecordDigestType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DSRecordDigestType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DSRecordDigestType(DSRecordDigestType(tmp).String())
	return nil
}

type DomainFeatureStatus string

const (
	// DomainFeatureStatusFeatureStatusUnknown is [insert doc].
	DomainFeatureStatusFeatureStatusUnknown = DomainFeatureStatus("feature_status_unknown")
	// DomainFeatureStatusEnabling is [insert doc].
	DomainFeatureStatusEnabling = DomainFeatureStatus("enabling")
	// DomainFeatureStatusEnabled is [insert doc].
	DomainFeatureStatusEnabled = DomainFeatureStatus("enabled")
	// DomainFeatureStatusDisabling is [insert doc].
	DomainFeatureStatusDisabling = DomainFeatureStatus("disabling")
	// DomainFeatureStatusDisabled is [insert doc].
	DomainFeatureStatusDisabled = DomainFeatureStatus("disabled")
)

func (enum DomainFeatureStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "feature_status_unknown"
	}
	return string(enum)
}

func (enum DomainFeatureStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DomainFeatureStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DomainFeatureStatus(DomainFeatureStatus(tmp).String())
	return nil
}

type DomainRegistrationStatusTransferStatus string

const (
	// DomainRegistrationStatusTransferStatusStatusUnknown is [insert doc].
	DomainRegistrationStatusTransferStatusStatusUnknown = DomainRegistrationStatusTransferStatus("status_unknown")
	// DomainRegistrationStatusTransferStatusPending is [insert doc].
	DomainRegistrationStatusTransferStatusPending = DomainRegistrationStatusTransferStatus("pending")
	// DomainRegistrationStatusTransferStatusWaitingVote is [insert doc].
	DomainRegistrationStatusTransferStatusWaitingVote = DomainRegistrationStatusTransferStatus("waiting_vote")
	// DomainRegistrationStatusTransferStatusRejected is [insert doc].
	DomainRegistrationStatusTransferStatusRejected = DomainRegistrationStatusTransferStatus("rejected")
	// DomainRegistrationStatusTransferStatusProcessing is [insert doc].
	DomainRegistrationStatusTransferStatusProcessing = DomainRegistrationStatusTransferStatus("processing")
	// DomainRegistrationStatusTransferStatusDone is [insert doc].
	DomainRegistrationStatusTransferStatusDone = DomainRegistrationStatusTransferStatus("done")
)

func (enum DomainRegistrationStatusTransferStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "status_unknown"
	}
	return string(enum)
}

func (enum DomainRegistrationStatusTransferStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DomainRegistrationStatusTransferStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DomainRegistrationStatusTransferStatus(DomainRegistrationStatusTransferStatus(tmp).String())
	return nil
}

type DomainStatus string

const (
	// DomainStatusStatusUnknown is [insert doc].
	DomainStatusStatusUnknown = DomainStatus("status_unknown")
	// DomainStatusActive is [insert doc].
	DomainStatusActive = DomainStatus("active")
	// DomainStatusCreating is [insert doc].
	DomainStatusCreating = DomainStatus("creating")
	// DomainStatusCreateError is [insert doc].
	DomainStatusCreateError = DomainStatus("create_error")
	// DomainStatusRenewing is [insert doc].
	DomainStatusRenewing = DomainStatus("renewing")
	// DomainStatusRenewError is [insert doc].
	DomainStatusRenewError = DomainStatus("renew_error")
	// DomainStatusXfering is [insert doc].
	DomainStatusXfering = DomainStatus("xfering")
	// DomainStatusXferError is [insert doc].
	DomainStatusXferError = DomainStatus("xfer_error")
	// DomainStatusExpired is [insert doc].
	DomainStatusExpired = DomainStatus("expired")
	// DomainStatusExpiring is [insert doc].
	DomainStatusExpiring = DomainStatus("expiring")
	// DomainStatusUpdating is [insert doc].
	DomainStatusUpdating = DomainStatus("updating")
	// DomainStatusChecking is [insert doc].
	DomainStatusChecking = DomainStatus("checking")
	// DomainStatusLocked is [insert doc].
	DomainStatusLocked = DomainStatus("locked")
	// DomainStatusDeleting is [insert doc].
	DomainStatusDeleting = DomainStatus("deleting")
)

func (enum DomainStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "status_unknown"
	}
	return string(enum)
}

func (enum DomainStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DomainStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DomainStatus(DomainStatus(tmp).String())
	return nil
}

type LanguageCode string

const (
	// LanguageCodeUnknownLanguageCode is [insert doc].
	LanguageCodeUnknownLanguageCode = LanguageCode("unknown_language_code")
	// LanguageCodeEnUS is [insert doc].
	LanguageCodeEnUS = LanguageCode("en_US")
	// LanguageCodeFrFR is [insert doc].
	LanguageCodeFrFR = LanguageCode("fr_FR")
)

func (enum LanguageCode) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_language_code"
	}
	return string(enum)
}

func (enum LanguageCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LanguageCode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LanguageCode(LanguageCode(tmp).String())
	return nil
}

type ListDNSZoneRecordsRequestOrderBy string

const (
	// ListDNSZoneRecordsRequestOrderByNameAsc is [insert doc].
	ListDNSZoneRecordsRequestOrderByNameAsc = ListDNSZoneRecordsRequestOrderBy("name_asc")
	// ListDNSZoneRecordsRequestOrderByNameDesc is [insert doc].
	ListDNSZoneRecordsRequestOrderByNameDesc = ListDNSZoneRecordsRequestOrderBy("name_desc")
)

func (enum ListDNSZoneRecordsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "name_asc"
	}
	return string(enum)
}

func (enum ListDNSZoneRecordsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListDNSZoneRecordsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListDNSZoneRecordsRequestOrderBy(ListDNSZoneRecordsRequestOrderBy(tmp).String())
	return nil
}

type ListDNSZonesRequestOrderBy string

const (
	// ListDNSZonesRequestOrderByDomainAsc is [insert doc].
	ListDNSZonesRequestOrderByDomainAsc = ListDNSZonesRequestOrderBy("domain_asc")
	// ListDNSZonesRequestOrderByDomainDesc is [insert doc].
	ListDNSZonesRequestOrderByDomainDesc = ListDNSZonesRequestOrderBy("domain_desc")
	// ListDNSZonesRequestOrderBySubdomainAsc is [insert doc].
	ListDNSZonesRequestOrderBySubdomainAsc = ListDNSZonesRequestOrderBy("subdomain_asc")
	// ListDNSZonesRequestOrderBySubdomainDesc is [insert doc].
	ListDNSZonesRequestOrderBySubdomainDesc = ListDNSZonesRequestOrderBy("subdomain_desc")
)

func (enum ListDNSZonesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "domain_asc"
	}
	return string(enum)
}

func (enum ListDNSZonesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListDNSZonesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListDNSZonesRequestOrderBy(ListDNSZonesRequestOrderBy(tmp).String())
	return nil
}

type ListDomainsRequestOrderBy string

const (
	// ListDomainsRequestOrderByDomainAsc is [insert doc].
	ListDomainsRequestOrderByDomainAsc = ListDomainsRequestOrderBy("domain_asc")
	// ListDomainsRequestOrderByDomainDesc is [insert doc].
	ListDomainsRequestOrderByDomainDesc = ListDomainsRequestOrderBy("domain_desc")
)

func (enum ListDomainsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "domain_asc"
	}
	return string(enum)
}

func (enum ListDomainsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListDomainsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListDomainsRequestOrderBy(ListDomainsRequestOrderBy(tmp).String())
	return nil
}

type ListRenewableDomainsRequestOrderBy string

const (
	// ListRenewableDomainsRequestOrderByDomainAsc is [insert doc].
	ListRenewableDomainsRequestOrderByDomainAsc = ListRenewableDomainsRequestOrderBy("domain_asc")
	// ListRenewableDomainsRequestOrderByDomainDesc is [insert doc].
	ListRenewableDomainsRequestOrderByDomainDesc = ListRenewableDomainsRequestOrderBy("domain_desc")
)

func (enum ListRenewableDomainsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "domain_asc"
	}
	return string(enum)
}

func (enum ListRenewableDomainsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListRenewableDomainsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListRenewableDomainsRequestOrderBy(ListRenewableDomainsRequestOrderBy(tmp).String())
	return nil
}

type RawFormat string

const (
	// RawFormatUnknownRawFormat is [insert doc].
	RawFormatUnknownRawFormat = RawFormat("unknown_raw_format")
	// RawFormatBind is [insert doc].
	RawFormatBind = RawFormat("bind")
)

func (enum RawFormat) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_raw_format"
	}
	return string(enum)
}

func (enum RawFormat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *RawFormat) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = RawFormat(RawFormat(tmp).String())
	return nil
}

type RecordHTTPServiceConfigStrategy string

const (
	// RecordHTTPServiceConfigStrategyRandom is [insert doc].
	RecordHTTPServiceConfigStrategyRandom = RecordHTTPServiceConfigStrategy("random")
	// RecordHTTPServiceConfigStrategyHashed is [insert doc].
	RecordHTTPServiceConfigStrategyHashed = RecordHTTPServiceConfigStrategy("hashed")
	// RecordHTTPServiceConfigStrategyAll is [insert doc].
	RecordHTTPServiceConfigStrategyAll = RecordHTTPServiceConfigStrategy("all")
)

func (enum RecordHTTPServiceConfigStrategy) String() string {
	if enum == "" {
		// return default value if empty
		return "random"
	}
	return string(enum)
}

func (enum RecordHTTPServiceConfigStrategy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *RecordHTTPServiceConfigStrategy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = RecordHTTPServiceConfigStrategy(RecordHTTPServiceConfigStrategy(tmp).String())
	return nil
}

type RecordType string

const (
	// RecordTypeUnknown is [insert doc].
	RecordTypeUnknown = RecordType("unknown")
	// RecordTypeA is [insert doc].
	RecordTypeA = RecordType("A")
	// RecordTypeAAAA is [insert doc].
	RecordTypeAAAA = RecordType("AAAA")
	// RecordTypeCNAME is [insert doc].
	RecordTypeCNAME = RecordType("CNAME")
	// RecordTypeTXT is [insert doc].
	RecordTypeTXT = RecordType("TXT")
	// RecordTypeSRV is [insert doc].
	RecordTypeSRV = RecordType("SRV")
	// RecordTypeTLSA is [insert doc].
	RecordTypeTLSA = RecordType("TLSA")
	// RecordTypeMX is [insert doc].
	RecordTypeMX = RecordType("MX")
	// RecordTypeNS is [insert doc].
	RecordTypeNS = RecordType("NS")
	// RecordTypePTR is [insert doc].
	RecordTypePTR = RecordType("PTR")
	// RecordTypeCAA is [insert doc].
	RecordTypeCAA = RecordType("CAA")
	// RecordTypeALIAS is [insert doc].
	RecordTypeALIAS = RecordType("ALIAS")
	// RecordTypeLOC is [insert doc].
	RecordTypeLOC = RecordType("LOC")
	// RecordTypeSSHFP is [insert doc].
	RecordTypeSSHFP = RecordType("SSHFP")
	// RecordTypeHINFO is [insert doc].
	RecordTypeHINFO = RecordType("HINFO")
	// RecordTypeRP is [insert doc].
	RecordTypeRP = RecordType("RP")
	// RecordTypeURI is [insert doc].
	RecordTypeURI = RecordType("URI")
	// RecordTypeDS is [insert doc].
	RecordTypeDS = RecordType("DS")
	// RecordTypeNAPTR is [insert doc].
	RecordTypeNAPTR = RecordType("NAPTR")
)

func (enum RecordType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum RecordType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *RecordType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = RecordType(RecordType(tmp).String())
	return nil
}

type RenewableDomainStatus string

const (
	// RenewableDomainStatusUnknown is [insert doc].
	RenewableDomainStatusUnknown = RenewableDomainStatus("unknown")
	// RenewableDomainStatusRenewable is [insert doc].
	RenewableDomainStatusRenewable = RenewableDomainStatus("renewable")
	// RenewableDomainStatusLateReneweable is [insert doc].
	RenewableDomainStatusLateReneweable = RenewableDomainStatus("late_reneweable")
	// RenewableDomainStatusNotRenewable is [insert doc].
	RenewableDomainStatusNotRenewable = RenewableDomainStatus("not_renewable")
)

func (enum RenewableDomainStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum RenewableDomainStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *RenewableDomainStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = RenewableDomainStatus(RenewableDomainStatus(tmp).String())
	return nil
}

type SSLCertificateStatus string

const (
	// SSLCertificateStatusUnknown is [insert doc].
	SSLCertificateStatusUnknown = SSLCertificateStatus("unknown")
	// SSLCertificateStatusNew is [insert doc].
	SSLCertificateStatusNew = SSLCertificateStatus("new")
	// SSLCertificateStatusPending is [insert doc].
	SSLCertificateStatusPending = SSLCertificateStatus("pending")
	// SSLCertificateStatusSuccess is [insert doc].
	SSLCertificateStatusSuccess = SSLCertificateStatus("success")
	// SSLCertificateStatusError is [insert doc].
	SSLCertificateStatusError = SSLCertificateStatus("error")
)

func (enum SSLCertificateStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum SSLCertificateStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SSLCertificateStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SSLCertificateStatus(SSLCertificateStatus(tmp).String())
	return nil
}

type TaskStatus string

const (
	// TaskStatusUnavailable is [insert doc].
	TaskStatusUnavailable = TaskStatus("unavailable")
	// TaskStatusNew is [insert doc].
	TaskStatusNew = TaskStatus("new")
	// TaskStatusWaitingPayment is [insert doc].
	TaskStatusWaitingPayment = TaskStatus("waiting_payment")
	// TaskStatusPending is [insert doc].
	TaskStatusPending = TaskStatus("pending")
	// TaskStatusSuccess is [insert doc].
	TaskStatusSuccess = TaskStatus("success")
	// TaskStatusError is [insert doc].
	TaskStatusError = TaskStatus("error")
)

func (enum TaskStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unavailable"
	}
	return string(enum)
}

func (enum TaskStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *TaskStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = TaskStatus(TaskStatus(tmp).String())
	return nil
}

type TaskType string

const (
	// TaskTypeUnknown is [insert doc].
	TaskTypeUnknown = TaskType("unknown")
	// TaskTypeCreateDomain is [insert doc].
	TaskTypeCreateDomain = TaskType("create_domain")
	// TaskTypeCreateExternalDomain is [insert doc].
	TaskTypeCreateExternalDomain = TaskType("create_external_domain")
	// TaskTypeRenewDomain is [insert doc].
	TaskTypeRenewDomain = TaskType("renew_domain")
	// TaskTypeTransferDomain is [insert doc].
	TaskTypeTransferDomain = TaskType("transfer_domain")
	// TaskTypeTradeDomain is [insert doc].
	TaskTypeTradeDomain = TaskType("trade_domain")
	// TaskTypeLockDomainTransfer is [insert doc].
	TaskTypeLockDomainTransfer = TaskType("lock_domain_transfer")
	// TaskTypeUnlockDomainTransfer is [insert doc].
	TaskTypeUnlockDomainTransfer = TaskType("unlock_domain_transfer")
	// TaskTypeEnableDnssec is [insert doc].
	TaskTypeEnableDnssec = TaskType("enable_dnssec")
	// TaskTypeDisableDnssec is [insert doc].
	TaskTypeDisableDnssec = TaskType("disable_dnssec")
	// TaskTypeUpdateDomain is [insert doc].
	TaskTypeUpdateDomain = TaskType("update_domain")
	// TaskTypeUpdateContact is [insert doc].
	TaskTypeUpdateContact = TaskType("update_contact")
	// TaskTypeDeleteDomain is [insert doc].
	TaskTypeDeleteDomain = TaskType("delete_domain")
	// TaskTypeCancelTask is [insert doc].
	TaskTypeCancelTask = TaskType("cancel_task")
	// TaskTypeGenerateSslCertificate is [insert doc].
	TaskTypeGenerateSslCertificate = TaskType("generate_ssl_certificate")
	// TaskTypeRenewSslCertificate is [insert doc].
	TaskTypeRenewSslCertificate = TaskType("renew_ssl_certificate")
	// TaskTypeSendMessage is [insert doc].
	TaskTypeSendMessage = TaskType("send_message")
	// TaskTypeDeleteDomainExpired is [insert doc].
	TaskTypeDeleteDomainExpired = TaskType("delete_domain_expired")
	// TaskTypeDeleteExternalDomain is [insert doc].
	TaskTypeDeleteExternalDomain = TaskType("delete_external_domain")
)

func (enum TaskType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum TaskType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *TaskType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = TaskType(TaskType(tmp).String())
	return nil
}

type AvailableDomain struct {
	Domain string `json:"domain"`

	Available bool `json:"available"`

	Tld *Tld `json:"tld"`
}

// ClearDNSZoneRecordsResponse: clear dns zone records response
type ClearDNSZoneRecordsResponse struct {
}

// Contact: contact
type Contact struct {
	ID string `json:"id"`
	// LegalForm:
	//
	// Default value: legal_form_unknown
	LegalForm ContactLegalForm `json:"legal_form"`
	// Civility:
	//
	// Default value: civility_unknown
	Civility ContactCivility `json:"civility"`

	Firstname string `json:"firstname"`

	Lastname string `json:"lastname"`

	CompanyName string `json:"company_name"`

	Email string `json:"email"`

	EmailAlt string `json:"email_alt"`

	PhoneNumber string `json:"phone_number"`

	FaxNumber string `json:"fax_number"`

	AddressLine1 string `json:"address_line_1"`

	AddressLine2 string `json:"address_line_2"`

	Zip string `json:"zip"`

	City string `json:"city"`

	Country string `json:"country"`

	VatIdentificationCode string `json:"vat_identification_code"`

	CompanyIdentificationCode string `json:"company_identification_code"`
	// Lang:
	//
	// Default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale bool `json:"resale"`

	Questions []*ContactQuestion `json:"questions"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn bool `json:"whois_opt_in"`
	// EmailStatus:
	//
	// Default value: email_status_unknown
	EmailStatus ContactEmailStatus `json:"email_status"`

	State string `json:"state"`
}

type ContactExtensionEU struct {
	EuropeanCitizenship string `json:"european_citizenship"`
}

type ContactExtensionFR struct {
	// Mode:
	//
	// Default value: mode_unknown
	Mode ContactExtensionFRMode `json:"mode"`

	// Precisely one of AssociationInfos, BrandInfos, CodeAuthAfnicInfos, DunsInfos, ParticularInfos must be set.
	ParticularInfos *ContactExtensionFRParticularInfos `json:"particular_infos,omitempty"`

	// Precisely one of AssociationInfos, BrandInfos, CodeAuthAfnicInfos, DunsInfos, ParticularInfos must be set.
	DunsInfos *ContactExtensionFRDunsInfos `json:"duns_infos,omitempty"`

	// Precisely one of AssociationInfos, BrandInfos, CodeAuthAfnicInfos, DunsInfos, ParticularInfos must be set.
	AssociationInfos *ContactExtensionFRAssociationInfos `json:"association_infos,omitempty"`

	// Precisely one of AssociationInfos, BrandInfos, CodeAuthAfnicInfos, DunsInfos, ParticularInfos must be set.
	BrandInfos *ContactExtensionFRBrandInfos `json:"brand_infos,omitempty"`

	// Precisely one of AssociationInfos, BrandInfos, CodeAuthAfnicInfos, DunsInfos, ParticularInfos must be set.
	CodeAuthAfnicInfos *ContactExtensionFRCodeAuthAfnicInfos `json:"code_auth_afnic_infos,omitempty"`
}

type ContactExtensionFRAssociationInfos struct {
	PublicationJo *time.Time `json:"publication_jo"`

	PublicationJoPage uint32 `json:"publication_jo_page"`
}

type ContactExtensionFRBrandInfos struct {
	BrandInpi string `json:"brand_inpi"`
}

type ContactExtensionFRCodeAuthAfnicInfos struct {
	CodeAuthAfnic string `json:"code_auth_afnic"`
}

type ContactExtensionFRDunsInfos struct {
	DunsID string `json:"duns_id"`

	LocalID string `json:"local_id"`
}

type ContactExtensionFRParticularInfos struct {
	WhoisOptIn bool `json:"whois_opt_in"`
}

type ContactQuestion struct {
	Question string `json:"question"`

	Answer string `json:"answer"`
}

type ContactRoles struct {
	Contact *Contact `json:"contact"`

	Roles map[string]*ContactRolesRoles `json:"roles"`
}

type ContactRolesRoles struct {
	IsOwner bool `json:"is_owner"`

	IsAdministrative bool `json:"is_administrative"`

	IsTechnical bool `json:"is_technical"`
}

type DNSZone struct {
	Domain string `json:"domain"`

	Subdomain string `json:"subdomain"`

	Ns []string `json:"ns"`

	NsDefault []string `json:"ns_default"`

	NsMaster []string `json:"ns_master"`
	// Status:
	//
	// Default value: unknown
	Status DNSZoneStatus `json:"status"`

	Message *string `json:"message"`

	UpdatedAt *time.Time `json:"updated_at"`

	ProjectID string `json:"project_id"`
}

type DNSZoneVersion struct {
	ID string `json:"id"`

	CreatedAt *time.Time `json:"created_at"`
}

type DSRecord struct {
	KeyID uint32 `json:"key_id"`
	// Algorithm:
	//
	// Default value: rsamd5
	Algorithm DSRecordAlgorithm `json:"algorithm"`

	// Precisely one of Digest, PublicKey must be set.
	Digest *DSRecordDigest `json:"digest,omitempty"`

	// Precisely one of Digest, PublicKey must be set.
	PublicKey *DSRecordPublicKey `json:"public_key,omitempty"`
}

type DSRecordDigest struct {
	// Type:
	//
	// Default value: sha_1
	Type DSRecordDigestType `json:"type"`

	Digest string `json:"digest"`
}

type DSRecordPublicKey struct {
	Key string `json:"key"`
}

// DeleteDNSZoneResponse: delete dns zone response
type DeleteDNSZoneResponse struct {
}

// DeleteExternalDomainResponse: delete external domain response
type DeleteExternalDomainResponse struct {
}

// DeleteSSLCertificateResponse: delete ssl certificate response
type DeleteSSLCertificateResponse struct {
}

// Domain: domain
type Domain struct {
	Domain string `json:"domain"`

	OrganizationID string `json:"organization_id"`

	ProjectID string `json:"project_id"`
	// AutoRenewStatus:
	//
	// Default value: feature_status_unknown
	AutoRenewStatus DomainFeatureStatus `json:"auto_renew_status"`

	Dnssec *DomainDNSSEC `json:"dnssec"`

	EppCode []string `json:"epp_code"`

	ExpiredAt *time.Time `json:"expired_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	Registrar string `json:"registrar"`

	IsExternal bool `json:"is_external"`
	// Status:
	//
	// Default value: status_unknown
	Status DomainStatus `json:"status"`

	DNSZones []*DNSZone `json:"dns_zones"`

	OwnerContact *Contact `json:"owner_contact"`

	TechnicalContact *Contact `json:"technical_contact"`

	AdministrativeContact *Contact `json:"administrative_contact"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	ExternalDomainRegistrationStatus *DomainRegistrationStatusExternalDomain `json:"external_domain_registration_status,omitempty"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	TransferRegistrationStatus *DomainRegistrationStatusTransfer `json:"transfer_registration_status,omitempty"`
}

type DomainDNSSEC struct {
	// Status:
	//
	// Default value: feature_status_unknown
	Status DomainFeatureStatus `json:"status"`

	DsRecords []*DSRecord `json:"ds_records"`
}

type DomainRegistrationStatusExternalDomain struct {
	ValidationToken string `json:"validation_token"`
}

type DomainRegistrationStatusTransfer struct {
	// Status:
	//
	// Default value: status_unknown
	Status DomainRegistrationStatusTransferStatus `json:"status"`

	VoteCurrentOwner bool `json:"vote_current_owner"`

	VoteNewOwner bool `json:"vote_new_owner"`
}

type DomainSummary struct {
	Domain string `json:"domain"`

	ProjectID string `json:"project_id"`
	// AutoRenewStatus:
	//
	// Default value: feature_status_unknown
	AutoRenewStatus DomainFeatureStatus `json:"auto_renew_status"`
	// DnssecStatus:
	//
	// Default value: feature_status_unknown
	DnssecStatus DomainFeatureStatus `json:"dnssec_status"`

	EppCode []string `json:"epp_code"`

	ExpiredAt *time.Time `json:"expired_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	Registrar string `json:"registrar"`

	IsExternal bool `json:"is_external"`
	// Status:
	//
	// Default value: status_unknown
	Status DomainStatus `json:"status"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	ExternalDomainRegistrationStatus *DomainRegistrationStatusExternalDomain `json:"external_domain_registration_status,omitempty"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	TransferRegistrationStatus *DomainRegistrationStatusTransfer `json:"transfer_registration_status,omitempty"`

	OrganizationID string `json:"organization_id"`
}

// GetDNSZoneTsigKeyResponse: get dns zone tsig key response
type GetDNSZoneTsigKeyResponse struct {
	Name string `json:"name"`

	Key string `json:"key"`

	Algorithm string `json:"algorithm"`
}

// GetDNSZoneVersionDiffResponse: get dns zone version diff response
type GetDNSZoneVersionDiffResponse struct {
	Changes []*RecordChange `json:"changes"`
}

// GetDomainAuthCodeResponse: get domain auth code response
type GetDomainAuthCodeResponse struct {
	AuthCode string `json:"auth_code"`
}

type ImportProviderDNSZoneRequestOnlineV1 struct {
	Token string `json:"token"`
}

// ImportProviderDNSZoneResponse: import provider dns zone response
type ImportProviderDNSZoneResponse struct {
	Records []*Record `json:"records"`
}

type ImportRawDNSZoneRequestAXFRSource struct {
	NameServer string `json:"name_server"`

	TsigKey *ImportRawDNSZoneRequestTsigKey `json:"tsig_key"`
}

type ImportRawDNSZoneRequestBindSource struct {
	Content string `json:"content"`
}

type ImportRawDNSZoneRequestTsigKey struct {
	Name string `json:"name"`

	Key string `json:"key"`

	Algorithm string `json:"algorithm"`
}

// ImportRawDNSZoneResponse: import raw dns zone response
type ImportRawDNSZoneResponse struct {
	Records []*Record `json:"records"`
}

// ListContactsResponse: list contacts response
type ListContactsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Contacts []*ContactRoles `json:"contacts"`
}

// ListDNSZoneNameserversResponse: list dns zone nameservers response
type ListDNSZoneNameserversResponse struct {
	// Ns: the returned DNS zone nameservers
	Ns []*Nameserver `json:"ns"`
}

// ListDNSZoneRecordsResponse: list dns zone records response
type ListDNSZoneRecordsResponse struct {
	// TotalCount: the total number of DNS zone records
	TotalCount uint32 `json:"total_count"`
	// Records: the paginated returned DNS zone records
	Records []*Record `json:"records"`
}

// ListDNSZoneVersionRecordsResponse: list dns zone version records response
type ListDNSZoneVersionRecordsResponse struct {
	// TotalCount: the total number of DNS zones versions records
	TotalCount uint32 `json:"total_count"`

	Records []*Record `json:"records"`
}

// ListDNSZoneVersionsResponse: list dns zone versions response
type ListDNSZoneVersionsResponse struct {
	// TotalCount: the total number of DNS zones versions
	TotalCount uint32 `json:"total_count"`

	Versions []*DNSZoneVersion `json:"versions"`
}

// ListDNSZonesResponse: list dns zones response
type ListDNSZonesResponse struct {
	// TotalCount: the total number of DNS zones
	TotalCount uint32 `json:"total_count"`
	// DNSZones: the paginated returned DNS zones
	DNSZones []*DNSZone `json:"dns_zones"`
}

// ListDomainsResponse: list domains response
type ListDomainsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Domains []*DomainSummary `json:"domains"`
}

// ListRenewableDomainsResponse: list renewable domains response
type ListRenewableDomainsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Domains []*RenewableDomain `json:"domains"`
}

// ListSSLCertificatesResponse: list ssl certificates response
type ListSSLCertificatesResponse struct {
	TotalCount uint32 `json:"total_count"`

	Certificates []*SSLCertificate `json:"certificates"`
}

// ListTasksResponse: list tasks response
type ListTasksResponse struct {
	TotalCount uint32 `json:"total_count"`

	Tasks []*Task `json:"tasks"`
}

type Nameserver struct {
	Name string `json:"name"`

	IP []string `json:"ip"`
}

type NewContact struct {
	// LegalForm:
	//
	// Default value: legal_form_unknown
	LegalForm ContactLegalForm `json:"legal_form"`
	// Civility:
	//
	// Default value: civility_unknown
	Civility ContactCivility `json:"civility"`

	Firstname string `json:"firstname"`

	Lastname string `json:"lastname"`

	CompanyName *string `json:"company_name"`

	Email string `json:"email"`

	EmailAlt *string `json:"email_alt"`

	PhoneNumber string `json:"phone_number"`

	FaxNumber *string `json:"fax_number"`

	AddressLine1 string `json:"address_line_1"`

	AddressLine2 *string `json:"address_line_2"`

	Zip string `json:"zip"`

	City string `json:"city"`

	Country string `json:"country"`

	VatIdentificationCode *string `json:"vat_identification_code"`

	CompanyIdentificationCode *string `json:"company_identification_code"`
	// Lang:
	//
	// Default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale bool `json:"resale"`

	Questions []*ContactQuestion `json:"questions"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn bool `json:"whois_opt_in"`

	State *string `json:"state"`
}

type OrderResponse struct {
	Domains []string `json:"domains"`

	OrganizationID string `json:"organization_id"`

	ProjectID string `json:"project_id"`

	TaskID string `json:"task_id"`

	CreatedAt *time.Time `json:"created_at"`
}

type Record struct {
	Data string `json:"data"`

	Name string `json:"name"`

	Priority uint32 `json:"priority"`

	TTL uint32 `json:"ttl"`
	// Type:
	//
	// Default value: unknown
	Type RecordType `json:"type"`

	Comment *string `json:"comment"`

	// Precisely one of GeoIPConfig, HTTPServiceConfig, ViewConfig, WeightedConfig must be set.
	GeoIPConfig *RecordGeoIPConfig `json:"geo_ip_config,omitempty"`

	// Precisely one of GeoIPConfig, HTTPServiceConfig, ViewConfig, WeightedConfig must be set.
	HTTPServiceConfig *RecordHTTPServiceConfig `json:"http_service_config,omitempty"`

	// Precisely one of GeoIPConfig, HTTPServiceConfig, ViewConfig, WeightedConfig must be set.
	WeightedConfig *RecordWeightedConfig `json:"weighted_config,omitempty"`

	// Precisely one of GeoIPConfig, HTTPServiceConfig, ViewConfig, WeightedConfig must be set.
	ViewConfig *RecordViewConfig `json:"view_config,omitempty"`

	ID string `json:"id"`
}

type RecordChange struct {

	// Precisely one of Add, Clear, Delete, Set must be set.
	Add *RecordChangeAdd `json:"add,omitempty"`

	// Precisely one of Add, Clear, Delete, Set must be set.
	Set *RecordChangeSet `json:"set,omitempty"`

	// Precisely one of Add, Clear, Delete, Set must be set.
	Delete *RecordChangeDelete `json:"delete,omitempty"`

	// Precisely one of Add, Clear, Delete, Set must be set.
	Clear *RecordChangeClear `json:"clear,omitempty"`
}

type RecordChangeAdd struct {
	Records []*Record `json:"records"`
}

type RecordChangeClear struct {
}

type RecordChangeDelete struct {

	// Precisely one of ID, IDFields must be set.
	ID *string `json:"id,omitempty"`

	// Precisely one of ID, IDFields must be set.
	IDFields *RecordIdentifier `json:"id_fields,omitempty"`
}

type RecordChangeSet struct {

	// Precisely one of ID, IDFields must be set.
	ID *string `json:"id,omitempty"`

	// Precisely one of ID, IDFields must be set.
	IDFields *RecordIdentifier `json:"id_fields,omitempty"`

	Records []*Record `json:"records"`
}

type RecordGeoIPConfig struct {
	Matches []*RecordGeoIPConfigMatch `json:"matches"`

	Default string `json:"default"`
}

type RecordGeoIPConfigMatch struct {
	Countries []string `json:"countries"`

	Continents []string `json:"continents"`

	Data string `json:"data"`
}

type RecordHTTPServiceConfig struct {
	IPs []net.IP `json:"ips"`

	MustContain *string `json:"must_contain"`

	URL string `json:"url"`

	UserAgent *string `json:"user_agent"`
	// Strategy:
	//
	// Default value: random
	Strategy RecordHTTPServiceConfigStrategy `json:"strategy"`
}

type RecordIdentifier struct {
	Name string `json:"name"`
	// Type:
	//
	// Default value: unknown
	Type RecordType `json:"type"`

	Data *string `json:"data"`

	TTL *uint32 `json:"ttl"`
}

type RecordViewConfig struct {
	Views []*RecordViewConfigView `json:"views"`
}

type RecordViewConfigView struct {
	Subnet string `json:"subnet"`

	Data string `json:"data"`
}

type RecordWeightedConfig struct {
	WeightedIPs []*RecordWeightedConfigWeightedIP `json:"weighted_ips"`
}

type RecordWeightedConfigWeightedIP struct {
	IP net.IP `json:"ip"`

	Weight uint32 `json:"weight"`
}

// RefreshDNSZoneResponse: refresh dns zone response
type RefreshDNSZoneResponse struct {
	// DNSZones: the returned DNS zones
	DNSZones []*DNSZone `json:"dns_zones"`
}

type RegisterExternalDomainResponse struct {
	Domain string `json:"domain"`

	OrganizationID string `json:"organization_id"`

	ValidationToken string `json:"validation_token"`

	CreatedAt *time.Time `json:"created_at"`

	ProjectID string `json:"project_id"`
}

type RenewableDomain struct {
	Domain string `json:"domain"`

	ProjectID string `json:"project_id"`

	OrganizationID string `json:"organization_id"`
	// Status:
	//
	// Default value: unknown
	Status RenewableDomainStatus `json:"status"`

	RenewableDurationInYears *int32 `json:"renewable_duration_in_years"`

	ExpiredAt *time.Time `json:"expired_at"`
}

// RestoreDNSZoneVersionResponse: restore dns zone version response
type RestoreDNSZoneVersionResponse struct {
}

type SSLCertificate struct {
	DNSZone string `json:"dns_zone"`

	AlternativeDNSZones []string `json:"alternative_dns_zones"`
	// Status:
	//
	// Default value: unknown
	Status SSLCertificateStatus `json:"status"`

	PrivateKey string `json:"private_key"`

	CertificateChain string `json:"certificate_chain"`

	CreatedAt *time.Time `json:"created_at"`

	ExpiredAt *time.Time `json:"expired_at"`
}

// SearchAvailableDomainsResponse: search available domains response
type SearchAvailableDomainsResponse struct {
	// AvailableDomains: array of available domains
	AvailableDomains []*AvailableDomain `json:"available_domains"`
}

type Task struct {
	ID string `json:"id"`

	ProjectID string `json:"project_id"`

	OrganizationID string `json:"organization_id"`

	Domain *string `json:"domain"`
	// Type:
	//
	// Default value: unknown
	Type TaskType `json:"type"`
	// Status:
	//
	// Default value: unavailable
	Status TaskStatus `json:"status"`

	StartedAt *time.Time `json:"started_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	Message *string `json:"message"`
}

type Tld struct {
	Name string `json:"name"`

	DnssecSupport bool `json:"dnssec_support"`

	DurationInYearsMin uint32 `json:"duration_in_years_min"`

	DurationInYearsMax uint32 `json:"duration_in_years_max"`

	IdnSupport bool `json:"idn_support"`

	Offers map[string]*TldOffer `json:"offers"`
}

type TldOffer struct {
	Action string `json:"action"`

	OperationPath string `json:"operation_path"`

	Price *scw.Money `json:"price"`
}

type UpdateContactRequestQuestion struct {
	Question *string `json:"question"`

	Answer *string `json:"answer"`
}

// UpdateDNSZoneNameserversResponse: update dns zone nameservers response
type UpdateDNSZoneNameserversResponse struct {
	// Ns: the returned DNS zone nameservers
	Ns []*Nameserver `json:"ns"`
}

// UpdateDNSZoneRecordsResponse: update dns zone records response
type UpdateDNSZoneRecordsResponse struct {
	// Records: the returned DNS zone records
	Records []*Record `json:"records"`
}

// Service API

type ListDNSZonesRequest struct {
	// OrganizationID: the organization ID on which to filter the returned DNS zones
	OrganizationID *string `json:"-"`
	// ProjectID: the project ID on which to filter the returned DNS zones
	ProjectID *string `json:"-"`
	// OrderBy: the sort order of the returned DNS zones
	//
	// Default value: domain_asc
	OrderBy ListDNSZonesRequestOrderBy `json:"-"`
	// Page: the page number for the returned DNS zones
	Page *int32 `json:"-"`
	// PageSize: the maximum number of DNS zones per page
	PageSize *uint32 `json:"-"`
	// Domain: the domain on which to filter the returned DNS zones
	Domain string `json:"-"`
	// DNSZone: the DNS zone on which to filter the returned DNS zones
	DNSZone string `json:"-"`
}

// ListDNSZones: list DNS zones
//
// Returns a list of manageable DNS zones.
// You can filter the DNS zones by domain name.
//
func (s *API) ListDNSZones(req *ListDNSZonesRequest, opts ...scw.RequestOption) (*ListDNSZonesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "domain", req.Domain)
	parameter.AddToQuery(query, "dns_zone", req.DNSZone)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDNSZonesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateDNSZoneRequest struct {
	// Domain: the domain of the DNS zone to create
	Domain string `json:"domain"`
	// Subdomain: the subdomain of the DNS zone to create
	Subdomain string `json:"subdomain"`
	// ProjectID: the project ID where the DNS zone will be created
	ProjectID string `json:"project_id"`
}

// CreateDNSZone: create a DNS zone
//
// Create a new DNS zone.
func (s *API) CreateDNSZone(req *CreateDNSZoneRequest, opts ...scw.RequestOption) (*DNSZone, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DNSZone

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateDNSZoneRequest struct {
	// DNSZone: the DNS zone to update
	DNSZone string `json:"-"`
	// NewDNSZone: the new DNS zone
	NewDNSZone *string `json:"new_dns_zone"`
	// ProjectID: the project ID of the new DNS zone
	ProjectID string `json:"project_id"`
}

// UpdateDNSZone: update a DNS zone
//
// Update the name and/or the organizations for a DNS zone.
func (s *API) UpdateDNSZone(req *UpdateDNSZoneRequest, opts ...scw.RequestOption) (*DNSZone, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DNSZone

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CloneDNSZoneRequest struct {
	// DNSZone: the DNS zone to clone
	DNSZone string `json:"-"`
	// DestDNSZone: the destinaton DNS zone
	DestDNSZone string `json:"dest_dns_zone"`
	// Overwrite: whether or not the destination DNS zone will be overwritten
	Overwrite bool `json:"overwrite"`
	// ProjectID: the project ID of the destination DNS zone
	ProjectID *string `json:"project_id"`
}

// CloneDNSZone: clone a DNS zone
//
// Clone an existed DNS zone with all its records into a new one.
func (s *API) CloneDNSZone(req *CloneDNSZoneRequest, opts ...scw.RequestOption) (*DNSZone, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/clone",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DNSZone

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteDNSZoneRequest struct {
	// DNSZone: the DNS zone to delete
	DNSZone string `json:"-"`
	// ProjectID: the project ID of the DNS zone to delete
	ProjectID string `json:"-"`
}

// DeleteDNSZone: delete DNS zone
//
// Delete a DNS zone and all it's records.
func (s *API) DeleteDNSZone(req *DeleteDNSZoneRequest, opts ...scw.RequestOption) (*DeleteDNSZoneResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	query := url.Values{}
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp DeleteDNSZoneResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListDNSZoneRecordsRequest struct {
	// DNSZone: the DNS zone on which to filter the returned DNS zone records
	DNSZone string `json:"-"`
	// ProjectID: the project ID on which to filter the returned DNS zone records
	ProjectID *string `json:"-"`
	// OrderBy: the sort order of the returned DNS zone records
	//
	// Default value: name_asc
	OrderBy ListDNSZoneRecordsRequestOrderBy `json:"-"`
	// Page: the page number for the returned DNS zone records
	Page *int32 `json:"-"`
	// PageSize: the maximum number of DNS zone records per page
	PageSize *uint32 `json:"-"`
	// Name: the name on which to filter the returned DNS zone records
	Name string `json:"-"`
	// Type: the record type on which to filter the returned DNS zone records
	//
	// Default value: unknown
	Type RecordType `json:"-"`
}

// ListDNSZoneRecords: list DNS zone records
//
// Returns a list of DNS records of a DNS zone with default NS.
// You can filter the records by type and name.
//
func (s *API) ListDNSZoneRecords(req *ListDNSZoneRecordsRequest, opts ...scw.RequestOption) (*ListDNSZoneRecordsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "type", req.Type)

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/records",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDNSZoneRecordsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateDNSZoneRecordsRequest struct {
	// DNSZone: the DNS zone where the DNS zone records will be updated
	DNSZone string `json:"-"`
	// Changes: the changes made to the records
	Changes []*RecordChange `json:"changes"`
	// ReturnAllRecords: whether or not to return all the records
	ReturnAllRecords *bool `json:"return_all_records"`
	// DisallowNewZoneCreation: forbid the creation of the target zone if not existing (default action is yes)
	DisallowNewZoneCreation bool `json:"disallow_new_zone_creation"`
	// Serial: don't use the autoincremenent serial but the provided one (0 to keep the same)
	Serial *uint64 `json:"serial"`
}

// UpdateDNSZoneRecords: update DNS zone records
//
// Only available with default NS.<br/>
// Send a list of actions and records.
//
// Action can be:
//  - add:
//   - Add new record
//   - Can be more specific and add a new IP to an existing A record for example
//  - set:
//   - Edit a record
//   - Can be more specific and edit an IP from an existing A record for example
//  - delete:
//   - Delete a record
//   - Can be more specific and delete an IP from an existing A record for example
//  - clear:
//   - Delete all records from a DNS zone
//
// All edits will be versioned.
//
func (s *API) UpdateDNSZoneRecords(req *UpdateDNSZoneRecordsRequest, opts ...scw.RequestOption) (*UpdateDNSZoneRecordsResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/records",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateDNSZoneRecordsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListDNSZoneNameserversRequest struct {
	// DNSZone: the DNS zone on which to filter the returned DNS zone nameservers
	DNSZone string `json:"-"`
	// ProjectID: the project ID on which to filter the returned DNS zone nameservers
	ProjectID *string `json:"-"`
}

// ListDNSZoneNameservers: list DNS zone nameservers
//
// Returns a list of Nameservers and their optional glue records for a DNS zone.
func (s *API) ListDNSZoneNameservers(req *ListDNSZoneNameserversRequest, opts ...scw.RequestOption) (*ListDNSZoneNameserversResponse, error) {
	var err error

	query := url.Values{}
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/nameservers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDNSZoneNameserversResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateDNSZoneNameserversRequest struct {
	// DNSZone: the DNS zone where the DNS zone nameservers will be updated
	DNSZone string `json:"-"`
	// Ns: the new DNS zone nameservers
	Ns []*Nameserver `json:"ns"`
}

// UpdateDNSZoneNameservers: update DNS zone nameservers
//
// Update DNS zone nameservers and set optional glue records.
func (s *API) UpdateDNSZoneNameservers(req *UpdateDNSZoneNameserversRequest, opts ...scw.RequestOption) (*UpdateDNSZoneNameserversResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/nameservers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateDNSZoneNameserversResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ClearDNSZoneRecordsRequest struct {
	// DNSZone: the DNS zone to clear
	DNSZone string `json:"-"`
}

// ClearDNSZoneRecords: clear DNS zone records
//
// Only available with default NS.<br/>
// Delete all the records from a DNS zone.
// All edits will be versioned.
//
func (s *API) ClearDNSZoneRecords(req *ClearDNSZoneRecordsRequest, opts ...scw.RequestOption) (*ClearDNSZoneRecordsResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/records",
		Headers: http.Header{},
	}

	var resp ClearDNSZoneRecordsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ExportRawDNSZoneRequest struct {
	// DNSZone: the DNS zone to export
	DNSZone string `json:"-"`
	// Format: format for DNS zone
	//
	// Default value: bind
	Format RawFormat `json:"-"`
}

// ExportRawDNSZone: export raw DNS zone
//
// Get a DNS zone in a given format with default NS.
func (s *API) ExportRawDNSZone(req *ExportRawDNSZoneRequest, opts ...scw.RequestOption) (*scw.File, error) {
	var err error

	query := url.Values{}
	parameter.AddToQuery(query, "format", req.Format)

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/raw",
		Query:   query,
		Headers: http.Header{},
	}

	var resp scw.File

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ImportRawDNSZoneRequest struct {
	// DNSZone: the DNS zone to import
	DNSZone string `json:"-"`
	// Deprecated
	Content string `json:"content"`

	ProjectID string `json:"project_id"`
	// Deprecated: Format:
	//
	// Default value: unknown_raw_format
	Format RawFormat `json:"format"`
	// BindSource: import a bind file format
	// Precisely one of AxfrSource, BindSource must be set.
	BindSource *ImportRawDNSZoneRequestBindSource `json:"bind_source,omitempty"`
	// AxfrSource: import from the nameserver given with tsig use or not
	// Precisely one of AxfrSource, BindSource must be set.
	AxfrSource *ImportRawDNSZoneRequestAXFRSource `json:"axfr_source,omitempty"`
}

// ImportRawDNSZone: import raw DNS zone
//
// Import and replace records from a given provider format with default NS.
func (s *API) ImportRawDNSZone(req *ImportRawDNSZoneRequest, opts ...scw.RequestOption) (*ImportRawDNSZoneResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/raw",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ImportRawDNSZoneResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ImportProviderDNSZoneRequest struct {
	DNSZone string `json:"-"`

	// Precisely one of OnlineV1 must be set.
	OnlineV1 *ImportProviderDNSZoneRequestOnlineV1 `json:"online_v1,omitempty"`
}

// ImportProviderDNSZone: import provider DNS zone
//
// Import and replace records from a given provider format with default NS.
func (s *API) ImportProviderDNSZone(req *ImportProviderDNSZoneRequest, opts ...scw.RequestOption) (*ImportProviderDNSZoneResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/import-provider",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ImportProviderDNSZoneResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RefreshDNSZoneRequest struct {
	// DNSZone: the DNS zone to refresh
	DNSZone string `json:"-"`
	// RecreateDNSZone: whether or not to recreate the DNS zone
	RecreateDNSZone bool `json:"recreate_dns_zone"`
	// RecreateSubDNSZone: whether or not to recreate the sub DNS zone
	RecreateSubDNSZone bool `json:"recreate_sub_dns_zone"`
}

// RefreshDNSZone: refresh DNS zone
//
// Refresh SOA DNS zone.
// You can recreate the given DNS zone and its sub DNS zone if needed.
//
func (s *API) RefreshDNSZone(req *RefreshDNSZoneRequest, opts ...scw.RequestOption) (*RefreshDNSZoneResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/refresh",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp RefreshDNSZoneResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListDNSZoneVersionsRequest struct {
	DNSZone string `json:"-"`
	// Page: the page number for the returned DNS zones versions
	Page *int32 `json:"-"`
	// PageSize: the maximum number of DNS zones versions per page
	PageSize *uint32 `json:"-"`
}

// ListDNSZoneVersions: list DNS zone versions
//
// Get a list of DNS zone versions.<br/>
// The maximum version count is 100.<br/>
// If the count reaches this limit, the oldest version will be deleted after each new modification.
//
func (s *API) ListDNSZoneVersions(req *ListDNSZoneVersionsRequest, opts ...scw.RequestOption) (*ListDNSZoneVersionsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/versions",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDNSZoneVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListDNSZoneVersionRecordsRequest struct {
	DNSZoneVersionID string `json:"-"`
	// Page: the page number for the returned DNS zones versions records
	Page *int32 `json:"-"`
	// PageSize: the maximum number of DNS zones versions records per page
	PageSize *uint32 `json:"-"`
}

// ListDNSZoneVersionRecords: list DNS zone version records
//
// Get a list of records from a previous DNS zone version.
func (s *API) ListDNSZoneVersionRecords(req *ListDNSZoneVersionRecordsRequest, opts ...scw.RequestOption) (*ListDNSZoneVersionRecordsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.DNSZoneVersionID) == "" {
		return nil, errors.New("field DNSZoneVersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/version/" + fmt.Sprint(req.DNSZoneVersionID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDNSZoneVersionRecordsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetDNSZoneVersionDiffRequest struct {
	DNSZoneVersionID string `json:"-"`
}

// GetDNSZoneVersionDiff: get DNS zone version diff
//
// Get all differences from a previous DNS zone version.
func (s *API) GetDNSZoneVersionDiff(req *GetDNSZoneVersionDiffRequest, opts ...scw.RequestOption) (*GetDNSZoneVersionDiffResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZoneVersionID) == "" {
		return nil, errors.New("field DNSZoneVersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/version/" + fmt.Sprint(req.DNSZoneVersionID) + "/diff",
		Headers: http.Header{},
	}

	var resp GetDNSZoneVersionDiffResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RestoreDNSZoneVersionRequest struct {
	DNSZoneVersionID string `json:"-"`
}

// RestoreDNSZoneVersion: restore DNS zone version
//
// Restore and activate a previous DNS zone version.
func (s *API) RestoreDNSZoneVersion(req *RestoreDNSZoneVersionRequest, opts ...scw.RequestOption) (*RestoreDNSZoneVersionResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZoneVersionID) == "" {
		return nil, errors.New("field DNSZoneVersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/dns-zones/version/" + fmt.Sprint(req.DNSZoneVersionID) + "/restore",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp RestoreDNSZoneVersionResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSSLCertificateRequest struct {
	DNSZone string `json:"-"`
}

// GetSSLCertificate: get the zone TLS certificate if it exists
func (s *API) GetSSLCertificate(req *GetSSLCertificateRequest, opts ...scw.RequestOption) (*SSLCertificate, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/ssl-certificates/" + fmt.Sprint(req.DNSZone) + "",
		Headers: http.Header{},
	}

	var resp SSLCertificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateSSLCertificateRequest struct {
	DNSZone string `json:"dns_zone"`

	AlternativeDNSZones []string `json:"alternative_dns_zones"`
}

// CreateSSLCertificate: create or return the zone TLS certificate
func (s *API) CreateSSLCertificate(req *CreateSSLCertificateRequest, opts ...scw.RequestOption) (*SSLCertificate, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/ssl-certificates",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SSLCertificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListSSLCertificatesRequest struct {
	DNSZone string `json:"-"`

	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`

	ProjectID *string `json:"-"`
}

// ListSSLCertificates: list all user TLS certificates
func (s *API) ListSSLCertificates(req *ListSSLCertificatesRequest, opts ...scw.RequestOption) (*ListSSLCertificatesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "dns_zone", req.DNSZone)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/ssl-certificates",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSSLCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSSLCertificateRequest struct {
	DNSZone string `json:"-"`
}

// DeleteSSLCertificate: delete an TLS certificate
func (s *API) DeleteSSLCertificate(req *DeleteSSLCertificateRequest, opts ...scw.RequestOption) (*DeleteSSLCertificateResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/ssl-certificates/" + fmt.Sprint(req.DNSZone) + "",
		Headers: http.Header{},
	}

	var resp DeleteSSLCertificateResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetDNSZoneTsigKeyRequest struct {
	DNSZone string `json:"-"`
}

// GetDNSZoneTsigKey: get the DNS zone TSIG Key
//
// Get the DNS zone TSIG Key to allow AXFR request.
func (s *API) GetDNSZoneTsigKey(req *GetDNSZoneTsigKeyRequest, opts ...scw.RequestOption) (*GetDNSZoneTsigKeyResponse, error) {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return nil, errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/tsig-key",
		Headers: http.Header{},
	}

	var resp GetDNSZoneTsigKeyResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteDNSZoneTsigKeyRequest struct {
	DNSZone string `json:"-"`
}

// DeleteDNSZoneTsigKey: delete the DNS zone TSIG Key
func (s *API) DeleteDNSZoneTsigKey(req *DeleteDNSZoneTsigKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.DNSZone) == "" {
		return errors.New("field DNSZone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/dns-zones/" + fmt.Sprint(req.DNSZone) + "/tsig-key",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// Service RegistrarAPI

type RegistrarAPIListTasksRequest struct {
	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`

	Domain string `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListTasks: list tasks
//
// List all account tasks.
// You can filter the list by domain name.
//
func (s *RegistrarAPI) ListTasks(req *RegistrarAPIListTasksRequest, opts ...scw.RequestOption) (*ListTasksResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "domain", req.Domain)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/tasks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListTasksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIBuyDomainsRequest struct {
	Domains []string `json:"domains"`

	DurationInYears uint32 `json:"duration_in_years"`

	ProjectID string `json:"project_id"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContactID *string `json:"owner_contact_id,omitempty"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContact *NewContact `json:"owner_contact,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContactID *string `json:"administrative_contact_id,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContact *NewContact `json:"administrative_contact,omitempty"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContactID *string `json:"technical_contact_id,omitempty"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContact *NewContact `json:"technical_contact,omitempty"`
}

// BuyDomains: buy one or more domains
//
// Request the registration of domain names.
// You can provide an already existing domain's contact or a new contact.
//
func (s *RegistrarAPI) BuyDomains(req *RegistrarAPIBuyDomainsRequest, opts ...scw.RequestOption) (*OrderResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/buy-domains",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIRenewDomainsRequest struct {
	Domains []string `json:"domains"`

	DurationInYears uint32 `json:"duration_in_years"`

	ForceLateRenewal *bool `json:"force_late_renewal"`
}

// RenewDomains: renew one or more domains
//
// Request the renewal of domain names.
//
func (s *RegistrarAPI) RenewDomains(req *RegistrarAPIRenewDomainsRequest, opts ...scw.RequestOption) (*OrderResponse, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/renew-domains",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPITransferInDomainRequest struct {
	Domain string `json:"domain"`

	AuthCode string `json:"auth_code"`

	ProjectID string `json:"project_id"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContactID *string `json:"owner_contact_id,omitempty"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContact *NewContact `json:"owner_contact,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContactID *string `json:"administrative_contact_id,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContact *NewContact `json:"administrative_contact,omitempty"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContactID *string `json:"technical_contact_id,omitempty"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContact *NewContact `json:"technical_contact,omitempty"`
}

// TransferInDomain: transfer a domain
//
// Request the transfer from another registrar domain to Scaleway.
//
func (s *RegistrarAPI) TransferInDomain(req *RegistrarAPITransferInDomainRequest, opts ...scw.RequestOption) (*OrderResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/domain-transfers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPITradeDomainRequest struct {
	Domain string `json:"-"`

	ProjectID *string `json:"project_id"`

	// Precisely one of NewOwnerContact, NewOwnerContactID must be set.
	NewOwnerContactID *string `json:"new_owner_contact_id,omitempty"`

	// Precisely one of NewOwnerContact, NewOwnerContactID must be set.
	NewOwnerContact *NewContact `json:"new_owner_contact,omitempty"`
}

// TradeDomain: trade a domain contact
//
// Request a trade for the contact owner.<br/>
// If an `organization_id` is given, the change is from the current Scaleway account to another Scaleway account.<br/>
// If no contact is given, the first contact of the other Scaleway account is taken.<br/>
// If the other Scaleway account has no contact. An error occurs.
//
func (s *RegistrarAPI) TradeDomain(req *RegistrarAPITradeDomainRequest, opts ...scw.RequestOption) (*OrderResponse, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/trade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp OrderResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIRegisterExternalDomainRequest struct {
	Domain string `json:"domain"`

	ProjectID string `json:"project_id"`
}

// RegisterExternalDomain: register an external domain
//
// Request the registration of an external domain name.
//
func (s *RegistrarAPI) RegisterExternalDomain(req *RegistrarAPIRegisterExternalDomainRequest, opts ...scw.RequestOption) (*RegisterExternalDomainResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/external-domains",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp RegisterExternalDomainResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIDeleteExternalDomainRequest struct {
	Domain string `json:"-"`
}

// DeleteExternalDomain: delete an external domain
//
// Delete an external domain name.
//
func (s *RegistrarAPI) DeleteExternalDomain(req *RegistrarAPIDeleteExternalDomainRequest, opts ...scw.RequestOption) (*DeleteExternalDomainResponse, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/external-domains/" + fmt.Sprint(req.Domain) + "",
		Headers: http.Header{},
	}

	var resp DeleteExternalDomainResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIListContactsRequest struct {
	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`

	Domain *string `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListContacts: list contacts
//
// Return a list of contacts with their domains and roles.
// You can filter the list by domain name.
//
func (s *RegistrarAPI) ListContacts(req *RegistrarAPIListContactsRequest, opts ...scw.RequestOption) (*ListContactsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "domain", req.Domain)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/contacts",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListContactsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIGetContactRequest struct {
	ContactID string `json:"-"`
}

// GetContact: get a contact
//
// Return a contact details retrieved from the registrar using a given contact ID.
func (s *RegistrarAPI) GetContact(req *RegistrarAPIGetContactRequest, opts ...scw.RequestOption) (*Contact, error) {
	var err error

	if fmt.Sprint(req.ContactID) == "" {
		return nil, errors.New("field ContactID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/contacts/" + fmt.Sprint(req.ContactID) + "",
		Headers: http.Header{},
	}

	var resp Contact

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIUpdateContactRequest struct {
	ContactID string `json:"-"`

	Email *string `json:"email"`

	EmailAlt *string `json:"email_alt"`

	PhoneNumber *string `json:"phone_number"`

	FaxNumber *string `json:"fax_number"`

	AddressLine1 *string `json:"address_line_1"`

	AddressLine2 *string `json:"address_line_2"`

	Zip *string `json:"zip"`

	City *string `json:"city"`

	Country *string `json:"country"`

	VatIdentificationCode *string `json:"vat_identification_code"`

	CompanyIdentificationCode *string `json:"company_identification_code"`
	// Lang:
	//
	// Default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale *bool `json:"resale"`

	Questions []*UpdateContactRequestQuestion `json:"questions"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn *bool `json:"whois_opt_in"`

	State *string `json:"state"`
}

// UpdateContact: update contact
//
// You can edit the contact coordinates.
func (s *RegistrarAPI) UpdateContact(req *RegistrarAPIUpdateContactRequest, opts ...scw.RequestOption) (*Contact, error) {
	var err error

	if fmt.Sprint(req.ContactID) == "" {
		return nil, errors.New("field ContactID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/domain/v2beta1/contacts/" + fmt.Sprint(req.ContactID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Contact

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIListDomainsRequest struct {
	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`
	// OrderBy:
	//
	// Default value: domain_asc
	OrderBy ListDomainsRequestOrderBy `json:"-"`

	Registrar *string `json:"-"`
	// Status:
	//
	// Default value: status_unknown
	Status DomainStatus `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`

	IsExternal *bool `json:"-"`
}

// ListDomains: list domains
//
// Returns a list of domains owned by the user.
func (s *RegistrarAPI) ListDomains(req *RegistrarAPIListDomainsRequest, opts ...scw.RequestOption) (*ListDomainsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "registrar", req.Registrar)
	parameter.AddToQuery(query, "status", req.Status)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "is_external", req.IsExternal)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/domains",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDomainsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIListRenewableDomainsRequest struct {
	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`
	// OrderBy:
	//
	// Default value: domain_asc
	OrderBy ListRenewableDomainsRequestOrderBy `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListRenewableDomains: list scaleway domains that can or not be renewed
//
// Returns a list of domains owned by the user with a renew status and if renewable, the maximum renew duration in years.
func (s *RegistrarAPI) ListRenewableDomains(req *RegistrarAPIListRenewableDomainsRequest, opts ...scw.RequestOption) (*ListRenewableDomainsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/renewable-domains",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListRenewableDomainsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIGetDomainRequest struct {
	Domain string `json:"-"`
}

// GetDomain: get domain
//
// Returns a the domain with more informations.
func (s *RegistrarAPI) GetDomain(req *RegistrarAPIGetDomainRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "",
		Headers: http.Header{},
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIUpdateDomainRequest struct {
	Domain string `json:"-"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContactID *string `json:"technical_contact_id,omitempty"`

	// Precisely one of TechnicalContact, TechnicalContactID must be set.
	TechnicalContact *NewContact `json:"technical_contact,omitempty"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContactID *string `json:"owner_contact_id,omitempty"`

	// Precisely one of OwnerContact, OwnerContactID must be set.
	OwnerContact *NewContact `json:"owner_contact,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContactID *string `json:"administrative_contact_id,omitempty"`

	// Precisely one of AdministrativeContact, AdministrativeContactID must be set.
	AdministrativeContact *NewContact `json:"administrative_contact,omitempty"`
}

// UpdateDomain: update a domain
//
// Update the domain contacts or create a new one.<br/>
// If you add the same contact for multiple roles. Only one ID will be created and used for all of them.
//
func (s *RegistrarAPI) UpdateDomain(req *RegistrarAPIUpdateDomainRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPILockDomainTransferRequest struct {
	Domain string `json:"-"`
}

// LockDomainTransfer: lock domain transfer
//
// Lock domain transfer. A locked domain transfer can't be transferred and the auth code can't be requested.
//
func (s *RegistrarAPI) LockDomainTransfer(req *RegistrarAPILockDomainTransferRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/lock-transfer",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIUnlockDomainTransferRequest struct {
	Domain string `json:"-"`
}

// UnlockDomainTransfer: unlock domain transfer
//
// Unlock domain transfer. An unlocked domain can be transferred and the auth code can be requested for this.
//
func (s *RegistrarAPI) UnlockDomainTransfer(req *RegistrarAPIUnlockDomainTransferRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/unlock-transfer",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIEnableDomainAutoRenewRequest struct {
	Domain string `json:"-"`
}

// EnableDomainAutoRenew: enable domain auto renew
func (s *RegistrarAPI) EnableDomainAutoRenew(req *RegistrarAPIEnableDomainAutoRenewRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/enable-auto-renew",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIDisableDomainAutoRenewRequest struct {
	Domain string `json:"-"`
}

// DisableDomainAutoRenew: disable domain auto renew
func (s *RegistrarAPI) DisableDomainAutoRenew(req *RegistrarAPIDisableDomainAutoRenewRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/disable-auto-renew",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIGetDomainAuthCodeRequest struct {
	Domain string `json:"-"`
}

// GetDomainAuthCode: return domain auth code
//
// If possible, return the auth code for an unlocked domain transfer, or an error if the domain is locked.
// Some TLD may have a different procedure to retrieve the auth code, in that case, the information is given in the message field.
//
func (s *RegistrarAPI) GetDomainAuthCode(req *RegistrarAPIGetDomainAuthCodeRequest, opts ...scw.RequestOption) (*GetDomainAuthCodeResponse, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/auth-code",
		Headers: http.Header{},
	}

	var resp GetDomainAuthCodeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIEnableDomainDNSSECRequest struct {
	Domain string `json:"-"`

	DsRecord *DSRecord `json:"ds_record"`
}

// EnableDomainDNSSEC: update domain DNSSEC
//
// If your domain has the default Scaleway NS and uses another registrar, you have to update the DS record manually.
// For the algorithm, here are the code numbers for each type:
//   - 1: RSAMD5
//   - 2: DIFFIE_HELLMAN
//   - 3: DSA_SHA1
//   - 5: RSA_SHA1
//   - 6: DSA_NSEC3_SHA1
//   - 7: RSASHA1_NSEC3_SHA1
//   - 8: RSASHA256
//   - 10: RSASHA512
//   - 12: ECC_GOST
//   - 13: ECDSAP256SHA256
//   - 14: ECDSAP384SHA384
//
// And for the digest type:
//   - 1: SHA_1
//   - 2: SHA_256
//   - 3: GOST_R_34_11_94
//   - 4: SHA_384
//
func (s *RegistrarAPI) EnableDomainDNSSEC(req *RegistrarAPIEnableDomainDNSSECRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/enable-dnssec",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIDisableDomainDNSSECRequest struct {
	Domain string `json:"-"`
}

// DisableDomainDNSSEC: disable domain DNSSEC
func (s *RegistrarAPI) DisableDomainDNSSEC(req *RegistrarAPIDisableDomainDNSSECRequest, opts ...scw.RequestOption) (*Domain, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/disable-dnssec",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Domain

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPISearchAvailableDomainsRequest struct {
	// Domains: a list of domain to search, TLD is optional
	Domains []string `json:"-"`
	// Tlds: array of tlds to search on
	Tlds []string `json:"-"`
}

// SearchAvailableDomains: search available domains
//
// Search a domain (or at maximum, 10 domains).
//
// If the TLD list is empty or not set the search returns the results from the most popular TLDs.
//
func (s *RegistrarAPI) SearchAvailableDomains(req *RegistrarAPISearchAvailableDomainsRequest, opts ...scw.RequestOption) (*SearchAvailableDomainsResponse, error) {
	var err error

	query := url.Values{}
	parameter.AddToQuery(query, "domains", req.Domains)
	parameter.AddToQuery(query, "tlds", req.Tlds)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/search-domains",
		Query:   query,
		Headers: http.Header{},
	}

	var resp SearchAvailableDomainsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDNSZonesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDNSZonesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDNSZonesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.DNSZones = append(r.DNSZones, results.DNSZones...)
	r.TotalCount += uint32(len(results.DNSZones))
	return uint32(len(results.DNSZones)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDNSZoneRecordsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDNSZoneRecordsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDNSZoneRecordsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Records = append(r.Records, results.Records...)
	r.TotalCount += uint32(len(results.Records))
	return uint32(len(results.Records)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDNSZoneVersionsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDNSZoneVersionsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDNSZoneVersionsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Versions = append(r.Versions, results.Versions...)
	r.TotalCount += uint32(len(results.Versions))
	return uint32(len(results.Versions)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDNSZoneVersionRecordsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDNSZoneVersionRecordsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDNSZoneVersionRecordsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Records = append(r.Records, results.Records...)
	r.TotalCount += uint32(len(results.Records))
	return uint32(len(results.Records)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSSLCertificatesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSSLCertificatesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListSSLCertificatesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Certificates = append(r.Certificates, results.Certificates...)
	r.TotalCount += uint32(len(results.Certificates))
	return uint32(len(results.Certificates)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListTasksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListTasksResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListTasksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Tasks = append(r.Tasks, results.Tasks...)
	r.TotalCount += uint32(len(results.Tasks))
	return uint32(len(results.Tasks)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListContactsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListContactsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListContactsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Contacts = append(r.Contacts, results.Contacts...)
	r.TotalCount += uint32(len(results.Contacts))
	return uint32(len(results.Contacts)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDomainsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDomainsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDomainsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Domains = append(r.Domains, results.Domains...)
	r.TotalCount += uint32(len(results.Domains))
	return uint32(len(results.Domains)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListRenewableDomainsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListRenewableDomainsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListRenewableDomainsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Domains = append(r.Domains, results.Domains...)
	r.TotalCount += uint32(len(results.Domains))
	return uint32(len(results.Domains)), nil
}
