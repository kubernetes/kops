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

// API: domains and DNS API.
// Manage your domains, DNS zones and records with the Domains and DNS API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

// RegistrarAPI: domains and DNS - Registrar API.
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

type ContactEmailStatus string

const (
	ContactEmailStatusEmailStatusUnknown = ContactEmailStatus("email_status_unknown")
	ContactEmailStatusValidated          = ContactEmailStatus("validated")
	ContactEmailStatusNotValidated       = ContactEmailStatus("not_validated")
	ContactEmailStatusInvalidEmail       = ContactEmailStatus("invalid_email")
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
	ContactExtensionFRModeModeUnknown               = ContactExtensionFRMode("mode_unknown")
	ContactExtensionFRModeIndividual                = ContactExtensionFRMode("individual")
	ContactExtensionFRModeCompanyIdentificationCode = ContactExtensionFRMode("company_identification_code")
	ContactExtensionFRModeDuns                      = ContactExtensionFRMode("duns")
	ContactExtensionFRModeLocal                     = ContactExtensionFRMode("local")
	ContactExtensionFRModeAssociation               = ContactExtensionFRMode("association")
	ContactExtensionFRModeTrademark                 = ContactExtensionFRMode("trademark")
	ContactExtensionFRModeCodeAuthAfnic             = ContactExtensionFRMode("code_auth_afnic")
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

type ContactExtensionNLLegalForm string

const (
	ContactExtensionNLLegalFormLegalFormUnknown                      = ContactExtensionNLLegalForm("legal_form_unknown")
	ContactExtensionNLLegalFormOther                                 = ContactExtensionNLLegalForm("other")
	ContactExtensionNLLegalFormNonDutchEuCompany                     = ContactExtensionNLLegalForm("non_dutch_eu_company")
	ContactExtensionNLLegalFormNonDutchLegalFormEnterpriseSubsidiary = ContactExtensionNLLegalForm("non_dutch_legal_form_enterprise_subsidiary")
	ContactExtensionNLLegalFormLimitedCompany                        = ContactExtensionNLLegalForm("limited_company")
	ContactExtensionNLLegalFormLimitedCompanyInFormation             = ContactExtensionNLLegalForm("limited_company_in_formation")
	ContactExtensionNLLegalFormCooperative                           = ContactExtensionNLLegalForm("cooperative")
	ContactExtensionNLLegalFormLimitedPartnership                    = ContactExtensionNLLegalForm("limited_partnership")
	ContactExtensionNLLegalFormSoleCompany                           = ContactExtensionNLLegalForm("sole_company")
	ContactExtensionNLLegalFormEuropeanEconomicInterestGroup         = ContactExtensionNLLegalForm("european_economic_interest_group")
	ContactExtensionNLLegalFormReligiousEntity                       = ContactExtensionNLLegalForm("religious_entity")
	ContactExtensionNLLegalFormPartnership                           = ContactExtensionNLLegalForm("partnership")
	ContactExtensionNLLegalFormPublicCompany                         = ContactExtensionNLLegalForm("public_company")
	ContactExtensionNLLegalFormMutualBenefitCompany                  = ContactExtensionNLLegalForm("mutual_benefit_company")
	ContactExtensionNLLegalFormResidential                           = ContactExtensionNLLegalForm("residential")
	ContactExtensionNLLegalFormShippingCompany                       = ContactExtensionNLLegalForm("shipping_company")
	ContactExtensionNLLegalFormFoundation                            = ContactExtensionNLLegalForm("foundation")
	ContactExtensionNLLegalFormAssociation                           = ContactExtensionNLLegalForm("association")
	ContactExtensionNLLegalFormTradingPartnership                    = ContactExtensionNLLegalForm("trading_partnership")
	ContactExtensionNLLegalFormNaturalPerson                         = ContactExtensionNLLegalForm("natural_person")
)

func (enum ContactExtensionNLLegalForm) String() string {
	if enum == "" {
		// return default value if empty
		return "legal_form_unknown"
	}
	return string(enum)
}

func (enum ContactExtensionNLLegalForm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ContactExtensionNLLegalForm) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ContactExtensionNLLegalForm(ContactExtensionNLLegalForm(tmp).String())
	return nil
}

type ContactLegalForm string

const (
	ContactLegalFormLegalFormUnknown = ContactLegalForm("legal_form_unknown")
	ContactLegalFormIndividual       = ContactLegalForm("individual")
	ContactLegalFormCorporate        = ContactLegalForm("corporate")
	ContactLegalFormAssociation      = ContactLegalForm("association")
	ContactLegalFormOther            = ContactLegalForm("other")
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
	DNSZoneStatusUnknown = DNSZoneStatus("unknown")
	DNSZoneStatusActive  = DNSZoneStatus("active")
	DNSZoneStatusPending = DNSZoneStatus("pending")
	DNSZoneStatusError   = DNSZoneStatus("error")
	DNSZoneStatusLocked  = DNSZoneStatus("locked")
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
	DSRecordAlgorithmRsamd5           = DSRecordAlgorithm("rsamd5")
	DSRecordAlgorithmDh               = DSRecordAlgorithm("dh")
	DSRecordAlgorithmDsa              = DSRecordAlgorithm("dsa")
	DSRecordAlgorithmRsasha1          = DSRecordAlgorithm("rsasha1")
	DSRecordAlgorithmDsaNsec3Sha1     = DSRecordAlgorithm("dsa_nsec3_sha1")
	DSRecordAlgorithmRsasha1Nsec3Sha1 = DSRecordAlgorithm("rsasha1_nsec3_sha1")
	DSRecordAlgorithmRsasha256        = DSRecordAlgorithm("rsasha256")
	DSRecordAlgorithmRsasha512        = DSRecordAlgorithm("rsasha512")
	DSRecordAlgorithmEccGost          = DSRecordAlgorithm("ecc_gost")
	DSRecordAlgorithmEcdsap256sha256  = DSRecordAlgorithm("ecdsap256sha256")
	DSRecordAlgorithmEcdsap384sha384  = DSRecordAlgorithm("ecdsap384sha384")
	DSRecordAlgorithmEd25519          = DSRecordAlgorithm("ed25519")
	DSRecordAlgorithmEd448            = DSRecordAlgorithm("ed448")
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
	DSRecordDigestTypeSha1          = DSRecordDigestType("sha_1")
	DSRecordDigestTypeSha256        = DSRecordDigestType("sha_256")
	DSRecordDigestTypeGostR34_11_94 = DSRecordDigestType("gost_r_34_11_94")
	DSRecordDigestTypeSha384        = DSRecordDigestType("sha_384")
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
	DomainFeatureStatusFeatureStatusUnknown = DomainFeatureStatus("feature_status_unknown")
	DomainFeatureStatusEnabling             = DomainFeatureStatus("enabling")
	DomainFeatureStatusEnabled              = DomainFeatureStatus("enabled")
	DomainFeatureStatusDisabling            = DomainFeatureStatus("disabling")
	DomainFeatureStatusDisabled             = DomainFeatureStatus("disabled")
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
	DomainRegistrationStatusTransferStatusStatusUnknown = DomainRegistrationStatusTransferStatus("status_unknown")
	DomainRegistrationStatusTransferStatusPending       = DomainRegistrationStatusTransferStatus("pending")
	DomainRegistrationStatusTransferStatusWaitingVote   = DomainRegistrationStatusTransferStatus("waiting_vote")
	DomainRegistrationStatusTransferStatusRejected      = DomainRegistrationStatusTransferStatus("rejected")
	DomainRegistrationStatusTransferStatusProcessing    = DomainRegistrationStatusTransferStatus("processing")
	DomainRegistrationStatusTransferStatusDone          = DomainRegistrationStatusTransferStatus("done")
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
	DomainStatusStatusUnknown = DomainStatus("status_unknown")
	DomainStatusActive        = DomainStatus("active")
	DomainStatusCreating      = DomainStatus("creating")
	DomainStatusCreateError   = DomainStatus("create_error")
	DomainStatusRenewing      = DomainStatus("renewing")
	DomainStatusRenewError    = DomainStatus("renew_error")
	DomainStatusXfering       = DomainStatus("xfering")
	DomainStatusXferError     = DomainStatus("xfer_error")
	DomainStatusExpired       = DomainStatus("expired")
	DomainStatusExpiring      = DomainStatus("expiring")
	DomainStatusUpdating      = DomainStatus("updating")
	DomainStatusChecking      = DomainStatus("checking")
	DomainStatusLocked        = DomainStatus("locked")
	DomainStatusDeleting      = DomainStatus("deleting")
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

type HostStatus string

const (
	HostStatusUnknownStatus = HostStatus("unknown_status")
	HostStatusActive        = HostStatus("active")
	HostStatusUpdating      = HostStatus("updating")
	HostStatusDeleting      = HostStatus("deleting")
)

func (enum HostStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_status"
	}
	return string(enum)
}

func (enum HostStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *HostStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = HostStatus(HostStatus(tmp).String())
	return nil
}

type LanguageCode string

const (
	LanguageCodeUnknownLanguageCode = LanguageCode("unknown_language_code")
	LanguageCodeEnUS                = LanguageCode("en_US")
	LanguageCodeFrFR                = LanguageCode("fr_FR")
	LanguageCodeDeDE                = LanguageCode("de_DE")
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
	ListDNSZoneRecordsRequestOrderByNameAsc  = ListDNSZoneRecordsRequestOrderBy("name_asc")
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
	ListDNSZonesRequestOrderByDomainAsc     = ListDNSZonesRequestOrderBy("domain_asc")
	ListDNSZonesRequestOrderByDomainDesc    = ListDNSZonesRequestOrderBy("domain_desc")
	ListDNSZonesRequestOrderBySubdomainAsc  = ListDNSZonesRequestOrderBy("subdomain_asc")
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
	ListDomainsRequestOrderByDomainAsc  = ListDomainsRequestOrderBy("domain_asc")
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
	ListRenewableDomainsRequestOrderByDomainAsc  = ListRenewableDomainsRequestOrderBy("domain_asc")
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

type ListTasksRequestOrderBy string

const (
	ListTasksRequestOrderByDomainDesc    = ListTasksRequestOrderBy("domain_desc")
	ListTasksRequestOrderByDomainAsc     = ListTasksRequestOrderBy("domain_asc")
	ListTasksRequestOrderByTypeAsc       = ListTasksRequestOrderBy("type_asc")
	ListTasksRequestOrderByTypeDesc      = ListTasksRequestOrderBy("type_desc")
	ListTasksRequestOrderByStatusAsc     = ListTasksRequestOrderBy("status_asc")
	ListTasksRequestOrderByStatusDesc    = ListTasksRequestOrderBy("status_desc")
	ListTasksRequestOrderByUpdatedAtAsc  = ListTasksRequestOrderBy("updated_at_asc")
	ListTasksRequestOrderByUpdatedAtDesc = ListTasksRequestOrderBy("updated_at_desc")
)

func (enum ListTasksRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "domain_desc"
	}
	return string(enum)
}

func (enum ListTasksRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListTasksRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListTasksRequestOrderBy(ListTasksRequestOrderBy(tmp).String())
	return nil
}

type RawFormat string

const (
	RawFormatUnknownRawFormat = RawFormat("unknown_raw_format")
	RawFormatBind             = RawFormat("bind")
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
	RecordHTTPServiceConfigStrategyRandom = RecordHTTPServiceConfigStrategy("random")
	RecordHTTPServiceConfigStrategyHashed = RecordHTTPServiceConfigStrategy("hashed")
	RecordHTTPServiceConfigStrategyAll    = RecordHTTPServiceConfigStrategy("all")
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
	RecordTypeUnknown = RecordType("unknown")
	RecordTypeA       = RecordType("A")
	RecordTypeAAAA    = RecordType("AAAA")
	RecordTypeCNAME   = RecordType("CNAME")
	RecordTypeTXT     = RecordType("TXT")
	RecordTypeSRV     = RecordType("SRV")
	RecordTypeTLSA    = RecordType("TLSA")
	RecordTypeMX      = RecordType("MX")
	RecordTypeNS      = RecordType("NS")
	RecordTypePTR     = RecordType("PTR")
	RecordTypeCAA     = RecordType("CAA")
	RecordTypeALIAS   = RecordType("ALIAS")
	RecordTypeLOC     = RecordType("LOC")
	RecordTypeSSHFP   = RecordType("SSHFP")
	RecordTypeHINFO   = RecordType("HINFO")
	RecordTypeRP      = RecordType("RP")
	RecordTypeURI     = RecordType("URI")
	RecordTypeDS      = RecordType("DS")
	RecordTypeNAPTR   = RecordType("NAPTR")
	RecordTypeDNAME   = RecordType("DNAME")
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
	RenewableDomainStatusUnknown        = RenewableDomainStatus("unknown")
	RenewableDomainStatusRenewable      = RenewableDomainStatus("renewable")
	RenewableDomainStatusLateReneweable = RenewableDomainStatus("late_reneweable")
	RenewableDomainStatusNotRenewable   = RenewableDomainStatus("not_renewable")
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
	SSLCertificateStatusUnknown = SSLCertificateStatus("unknown")
	SSLCertificateStatusNew     = SSLCertificateStatus("new")
	SSLCertificateStatusPending = SSLCertificateStatus("pending")
	SSLCertificateStatusSuccess = SSLCertificateStatus("success")
	SSLCertificateStatusError   = SSLCertificateStatus("error")
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
	TaskStatusUnavailable    = TaskStatus("unavailable")
	TaskStatusNew            = TaskStatus("new")
	TaskStatusWaitingPayment = TaskStatus("waiting_payment")
	TaskStatusPending        = TaskStatus("pending")
	TaskStatusSuccess        = TaskStatus("success")
	TaskStatusError          = TaskStatus("error")
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
	TaskTypeUnknown                = TaskType("unknown")
	TaskTypeCreateDomain           = TaskType("create_domain")
	TaskTypeCreateExternalDomain   = TaskType("create_external_domain")
	TaskTypeRenewDomain            = TaskType("renew_domain")
	TaskTypeTransferDomain         = TaskType("transfer_domain")
	TaskTypeTradeDomain            = TaskType("trade_domain")
	TaskTypeLockDomainTransfer     = TaskType("lock_domain_transfer")
	TaskTypeUnlockDomainTransfer   = TaskType("unlock_domain_transfer")
	TaskTypeEnableDnssec           = TaskType("enable_dnssec")
	TaskTypeDisableDnssec          = TaskType("disable_dnssec")
	TaskTypeUpdateDomain           = TaskType("update_domain")
	TaskTypeUpdateContact          = TaskType("update_contact")
	TaskTypeDeleteDomain           = TaskType("delete_domain")
	TaskTypeCancelTask             = TaskType("cancel_task")
	TaskTypeGenerateSslCertificate = TaskType("generate_ssl_certificate")
	TaskTypeRenewSslCertificate    = TaskType("renew_ssl_certificate")
	TaskTypeSendMessage            = TaskType("send_message")
	TaskTypeDeleteDomainExpired    = TaskType("delete_domain_expired")
	TaskTypeDeleteExternalDomain   = TaskType("delete_external_domain")
	TaskTypeCreateHost             = TaskType("create_host")
	TaskTypeUpdateHost             = TaskType("update_host")
	TaskTypeDeleteHost             = TaskType("delete_host")
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

// CheckContactsCompatibilityResponse: check contacts compatibility response.
type CheckContactsCompatibilityResponse struct {
	Compatible bool `json:"compatible"`

	OwnerCheckResult *CheckContactsCompatibilityResponseContactCheckResult `json:"owner_check_result"`

	AdministrativeCheckResult *CheckContactsCompatibilityResponseContactCheckResult `json:"administrative_check_result"`

	TechnicalCheckResult *CheckContactsCompatibilityResponseContactCheckResult `json:"technical_check_result"`
}

type CheckContactsCompatibilityResponseContactCheckResult struct {
	Compatible bool `json:"compatible"`

	ErrorMessage *string `json:"error_message"`
}

// ClearDNSZoneRecordsResponse: clear dns zone records response.
type ClearDNSZoneRecordsResponse struct {
}

// Contact: contact.
type Contact struct {
	ID string `json:"id"`
	// LegalForm: default value: legal_form_unknown
	LegalForm ContactLegalForm `json:"legal_form"`

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
	// Lang: default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale bool `json:"resale"`
	// Deprecated
	Questions *[]*ContactQuestion `json:"questions,omitempty"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn bool `json:"whois_opt_in"`
	// EmailStatus: default value: email_status_unknown
	EmailStatus ContactEmailStatus `json:"email_status"`

	State string `json:"state"`

	ExtensionNl *ContactExtensionNL `json:"extension_nl"`
}

type ContactExtensionEU struct {
	EuropeanCitizenship string `json:"european_citizenship"`
}

type ContactExtensionFR struct {
	// Mode: default value: mode_unknown
	Mode ContactExtensionFRMode `json:"mode"`

	// Precisely one of AssociationInfo, CodeAuthAfnicInfo, DunsInfo, IndividualInfo, TrademarkInfo must be set.
	IndividualInfo *ContactExtensionFRIndividualInfo `json:"individual_info,omitempty"`

	// Precisely one of AssociationInfo, CodeAuthAfnicInfo, DunsInfo, IndividualInfo, TrademarkInfo must be set.
	DunsInfo *ContactExtensionFRDunsInfo `json:"duns_info,omitempty"`

	// Precisely one of AssociationInfo, CodeAuthAfnicInfo, DunsInfo, IndividualInfo, TrademarkInfo must be set.
	AssociationInfo *ContactExtensionFRAssociationInfo `json:"association_info,omitempty"`

	// Precisely one of AssociationInfo, CodeAuthAfnicInfo, DunsInfo, IndividualInfo, TrademarkInfo must be set.
	TrademarkInfo *ContactExtensionFRTrademarkInfo `json:"trademark_info,omitempty"`

	// Precisely one of AssociationInfo, CodeAuthAfnicInfo, DunsInfo, IndividualInfo, TrademarkInfo must be set.
	CodeAuthAfnicInfo *ContactExtensionFRCodeAuthAfnicInfo `json:"code_auth_afnic_info,omitempty"`
}

type ContactExtensionFRAssociationInfo struct {
	PublicationJo *time.Time `json:"publication_jo"`

	PublicationJoPage uint32 `json:"publication_jo_page"`
}

type ContactExtensionFRCodeAuthAfnicInfo struct {
	CodeAuthAfnic string `json:"code_auth_afnic"`
}

type ContactExtensionFRDunsInfo struct {
	DunsID string `json:"duns_id"`

	LocalID string `json:"local_id"`
}

type ContactExtensionFRIndividualInfo struct {
	WhoisOptIn bool `json:"whois_opt_in"`
}

type ContactExtensionFRTrademarkInfo struct {
	TrademarkInpi string `json:"trademark_inpi"`
}

type ContactExtensionNL struct {
	// LegalForm: default value: legal_form_unknown
	LegalForm ContactExtensionNLLegalForm `json:"legal_form"`

	LegalFormRegistrationNumber string `json:"legal_form_registration_number"`
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
	// Status: default value: unknown
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
	// Algorithm: default value: rsamd5
	Algorithm DSRecordAlgorithm `json:"algorithm"`

	// Precisely one of Digest, PublicKey must be set.
	Digest *DSRecordDigest `json:"digest,omitempty"`

	// Precisely one of Digest, PublicKey must be set.
	PublicKey *DSRecordPublicKey `json:"public_key,omitempty"`
}

type DSRecordDigest struct {
	// Type: default value: sha_1
	Type DSRecordDigestType `json:"type"`

	Digest string `json:"digest"`

	PublicKey *DSRecordPublicKey `json:"public_key"`
}

type DSRecordPublicKey struct {
	Key string `json:"key"`
}

// DeleteDNSZoneResponse: delete dns zone response.
type DeleteDNSZoneResponse struct {
}

// DeleteExternalDomainResponse: delete external domain response.
type DeleteExternalDomainResponse struct {
}

// DeleteSSLCertificateResponse: delete ssl certificate response.
type DeleteSSLCertificateResponse struct {
}

// Domain: domain.
type Domain struct {
	Domain string `json:"domain"`

	OrganizationID string `json:"organization_id"`

	ProjectID string `json:"project_id"`
	// AutoRenewStatus: default value: feature_status_unknown
	AutoRenewStatus DomainFeatureStatus `json:"auto_renew_status"`

	Dnssec *DomainDNSSEC `json:"dnssec"`

	EppCode []string `json:"epp_code"`

	ExpiredAt *time.Time `json:"expired_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	Registrar string `json:"registrar"`

	IsExternal bool `json:"is_external"`
	// Status: default value: status_unknown
	Status DomainStatus `json:"status"`

	DNSZones []*DNSZone `json:"dns_zones"`

	OwnerContact *Contact `json:"owner_contact"`

	TechnicalContact *Contact `json:"technical_contact"`

	AdministrativeContact *Contact `json:"administrative_contact"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	ExternalDomainRegistrationStatus *DomainRegistrationStatusExternalDomain `json:"external_domain_registration_status,omitempty"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	TransferRegistrationStatus *DomainRegistrationStatusTransfer `json:"transfer_registration_status,omitempty"`

	Tld *Tld `json:"tld"`
}

type DomainDNSSEC struct {
	// Status: default value: feature_status_unknown
	Status DomainFeatureStatus `json:"status"`

	DsRecords []*DSRecord `json:"ds_records"`
}

type DomainRegistrationStatusExternalDomain struct {
	ValidationToken string `json:"validation_token"`
}

type DomainRegistrationStatusTransfer struct {
	// Status: default value: status_unknown
	Status DomainRegistrationStatusTransferStatus `json:"status"`

	VoteCurrentOwner bool `json:"vote_current_owner"`

	VoteNewOwner bool `json:"vote_new_owner"`
}

type DomainSummary struct {
	Domain string `json:"domain"`

	ProjectID string `json:"project_id"`
	// AutoRenewStatus: default value: feature_status_unknown
	AutoRenewStatus DomainFeatureStatus `json:"auto_renew_status"`
	// DnssecStatus: default value: feature_status_unknown
	DnssecStatus DomainFeatureStatus `json:"dnssec_status"`

	EppCode []string `json:"epp_code"`

	ExpiredAt *time.Time `json:"expired_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	Registrar string `json:"registrar"`

	IsExternal bool `json:"is_external"`
	// Status: default value: status_unknown
	Status DomainStatus `json:"status"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	ExternalDomainRegistrationStatus *DomainRegistrationStatusExternalDomain `json:"external_domain_registration_status,omitempty"`

	// Precisely one of ExternalDomainRegistrationStatus, TransferRegistrationStatus must be set.
	TransferRegistrationStatus *DomainRegistrationStatusTransfer `json:"transfer_registration_status,omitempty"`

	OrganizationID string `json:"organization_id"`
}

// GetDNSZoneTsigKeyResponse: get dns zone tsig key response.
type GetDNSZoneTsigKeyResponse struct {
	Name string `json:"name"`

	Key string `json:"key"`

	Algorithm string `json:"algorithm"`
}

// GetDNSZoneVersionDiffResponse: get dns zone version diff response.
type GetDNSZoneVersionDiffResponse struct {
	Changes []*RecordChange `json:"changes"`
}

// GetDomainAuthCodeResponse: get domain auth code response.
type GetDomainAuthCodeResponse struct {
	AuthCode string `json:"auth_code"`
}

type Host struct {
	Domain string `json:"domain"`

	Name string `json:"name"`

	IPs []net.IP `json:"ips"`
	// Status: default value: unknown_status
	Status HostStatus `json:"status"`
}

type ImportProviderDNSZoneRequestOnlineV1 struct {
	Token string `json:"token"`
}

// ImportProviderDNSZoneResponse: import provider dns zone response.
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

// ImportRawDNSZoneResponse: import raw dns zone response.
type ImportRawDNSZoneResponse struct {
	Records []*Record `json:"records"`
}

// ListContactsResponse: list contacts response.
type ListContactsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Contacts []*ContactRoles `json:"contacts"`
}

// ListDNSZoneNameserversResponse: list dns zone nameservers response.
type ListDNSZoneNameserversResponse struct {
	// Ns: DNS zone name servers returned.
	Ns []*Nameserver `json:"ns"`
}

// ListDNSZoneRecordsResponse: list dns zone records response.
type ListDNSZoneRecordsResponse struct {
	// TotalCount: total number of DNS zone records.
	TotalCount uint32 `json:"total_count"`
	// Records: paginated returned DNS zone records.
	Records []*Record `json:"records"`
}

// ListDNSZoneVersionRecordsResponse: list dns zone version records response.
type ListDNSZoneVersionRecordsResponse struct {
	// TotalCount: total number of DNS zones versions records.
	TotalCount uint32 `json:"total_count"`

	Records []*Record `json:"records"`
}

// ListDNSZoneVersionsResponse: list dns zone versions response.
type ListDNSZoneVersionsResponse struct {
	// TotalCount: total number of DNS zones versions.
	TotalCount uint32 `json:"total_count"`

	Versions []*DNSZoneVersion `json:"versions"`
}

// ListDNSZonesResponse: list dns zones response.
type ListDNSZonesResponse struct {
	// TotalCount: total number of DNS zones matching the requested criteria.
	TotalCount uint32 `json:"total_count"`
	// DNSZones: paginated returned DNS zones.
	DNSZones []*DNSZone `json:"dns_zones"`
}

// ListDomainHostsResponse: list domain hosts response.
type ListDomainHostsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Hosts []*Host `json:"hosts"`
}

// ListDomainsResponse: list domains response.
type ListDomainsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Domains []*DomainSummary `json:"domains"`
}

// ListRenewableDomainsResponse: list renewable domains response.
type ListRenewableDomainsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Domains []*RenewableDomain `json:"domains"`
}

// ListSSLCertificatesResponse: list ssl certificates response.
type ListSSLCertificatesResponse struct {
	TotalCount uint32 `json:"total_count"`

	Certificates []*SSLCertificate `json:"certificates"`
}

// ListTasksResponse: list tasks response.
type ListTasksResponse struct {
	TotalCount uint32 `json:"total_count"`

	Tasks []*Task `json:"tasks"`
}

type Nameserver struct {
	Name string `json:"name"`

	IP []string `json:"ip"`
}

type NewContact struct {
	// LegalForm: default value: legal_form_unknown
	LegalForm ContactLegalForm `json:"legal_form"`

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
	// Lang: default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale bool `json:"resale"`
	// Deprecated
	Questions *[]*ContactQuestion `json:"questions,omitempty"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn bool `json:"whois_opt_in"`

	State *string `json:"state"`

	ExtensionNl *ContactExtensionNL `json:"extension_nl"`
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
	// Type: default value: unknown
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
	// Strategy: default value: random
	Strategy RecordHTTPServiceConfigStrategy `json:"strategy"`
}

type RecordIdentifier struct {
	Name string `json:"name"`
	// Type: default value: unknown
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

// RefreshDNSZoneResponse: refresh dns zone response.
type RefreshDNSZoneResponse struct {
	// DNSZones: DNS zones returned.
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
	// Status: default value: unknown
	Status RenewableDomainStatus `json:"status"`

	RenewableDurationInYears *int32 `json:"renewable_duration_in_years"`

	ExpiredAt *time.Time `json:"expired_at"`

	LimitRenewAt *time.Time `json:"limit_renew_at"`

	LimitRedemptionAt *time.Time `json:"limit_redemption_at"`

	EstimatedDeleteAt *time.Time `json:"estimated_delete_at"`

	Tld *Tld `json:"tld"`
}

// RestoreDNSZoneVersionResponse: restore dns zone version response.
type RestoreDNSZoneVersionResponse struct {
}

type SSLCertificate struct {
	DNSZone string `json:"dns_zone"`

	AlternativeDNSZones []string `json:"alternative_dns_zones"`
	// Status: default value: unknown
	Status SSLCertificateStatus `json:"status"`

	PrivateKey string `json:"private_key"`

	CertificateChain string `json:"certificate_chain"`

	CreatedAt *time.Time `json:"created_at"`

	ExpiredAt *time.Time `json:"expired_at"`
}

// SearchAvailableDomainsResponse: search available domains response.
type SearchAvailableDomainsResponse struct {
	// AvailableDomains: array of available domains.
	AvailableDomains []*AvailableDomain `json:"available_domains"`
}

type Task struct {
	ID string `json:"id"`

	ProjectID string `json:"project_id"`

	OrganizationID string `json:"organization_id"`

	Domain *string `json:"domain"`
	// Type: default value: unknown
	Type TaskType `json:"type"`
	// Status: default value: unavailable
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

	Specifications map[string]string `json:"specifications"`
}

type TldOffer struct {
	Action string `json:"action"`

	OperationPath string `json:"operation_path"`

	Price *scw.Money `json:"price"`
}

type TransferInDomainRequestTransferRequest struct {
	Domain string `json:"domain"`

	AuthCode string `json:"auth_code"`
}

type UpdateContactRequestQuestion struct {
	Question *string `json:"question"`

	Answer *string `json:"answer"`
}

// UpdateDNSZoneNameserversResponse: update dns zone nameservers response.
type UpdateDNSZoneNameserversResponse struct {
	// Ns: DNS zone name servers returned.
	Ns []*Nameserver `json:"ns"`
}

// UpdateDNSZoneRecordsResponse: update dns zone records response.
type UpdateDNSZoneRecordsResponse struct {
	// Records: DNS zone records returned.
	Records []*Record `json:"records"`
}

// Service API

type ListDNSZonesRequest struct {
	// OrganizationID: organization ID on which to filter the returned DNS zones.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID on which to filter the returned DNS zones.
	ProjectID *string `json:"-"`
	// OrderBy: sort order of the returned DNS zones.
	// Default value: domain_asc
	OrderBy ListDNSZonesRequestOrderBy `json:"-"`
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of DNS zones to return per page.
	PageSize *uint32 `json:"-"`
	// Domain: domain on which to filter the returned DNS zones.
	Domain string `json:"-"`
	// DNSZone: DNS zone on which to filter the returned DNS zones.
	DNSZone string `json:"-"`
}

// ListDNSZones: list DNS zones.
// Retrieve the list of DNS zones you can manage and filter DNS zones associated with specific domain names.
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
	// Domain: domain in which to crreate the DNS zone.
	Domain string `json:"domain"`
	// Subdomain: subdomain of the DNS zone to create.
	Subdomain string `json:"subdomain"`
	// ProjectID: project ID in which to create the DNS zone.
	ProjectID string `json:"project_id"`
}

// CreateDNSZone: create a DNS zone.
// Create a new DNS zone specified by the domain name, the subdomain and the Project ID.
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
	// DNSZone: DNS zone to update.
	DNSZone string `json:"-"`
	// NewDNSZone: name of the new DNS zone to create.
	NewDNSZone *string `json:"new_dns_zone"`
	// ProjectID: project ID in which to create the new DNS zone.
	ProjectID string `json:"project_id"`
}

// UpdateDNSZone: update a DNS zone.
// Update the name and/or the Organizations for a DNS zone.
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
	// DNSZone: DNS zone to clone.
	DNSZone string `json:"-"`
	// DestDNSZone: destination DNS zone in which to clone the chosen DNS zone.
	DestDNSZone string `json:"dest_dns_zone"`
	// Overwrite: specifies whether or not the destination DNS zone will be overwritten.
	Overwrite bool `json:"overwrite"`
	// ProjectID: project ID of the destination DNS zone.
	ProjectID *string `json:"project_id"`
}

// CloneDNSZone: clone a DNS zone.
// Clone an existing DNS zone with all its records into a new DNS zone.
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
	// DNSZone: DNS zone to delete.
	DNSZone string `json:"-"`
	// ProjectID: project ID of the DNS zone to delete.
	ProjectID string `json:"-"`
}

// DeleteDNSZone: delete a DNS zone.
// Delete a DNS zone and all its records.
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
	// DNSZone: DNS zone on which to filter the returned DNS zone records.
	DNSZone string `json:"-"`
	// ProjectID: project ID on which to filter the returned DNS zone records.
	ProjectID *string `json:"-"`
	// OrderBy: sort order of the returned DNS zone records.
	// Default value: name_asc
	OrderBy ListDNSZoneRecordsRequestOrderBy `json:"-"`
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of DNS zone records per page.
	PageSize *uint32 `json:"-"`
	// Name: name on which to filter the returned DNS zone records.
	Name string `json:"-"`
	// Type: record type on which to filter the returned DNS zone records.
	// Default value: unknown
	Type RecordType `json:"-"`
	// ID: record ID on which to filter the returned DNS zone records.
	ID *string `json:"-"`
}

// ListDNSZoneRecords: list records within a DNS zone.
// Retrieve a list of DNS records within a DNS zone that has default name servers.
// You can filter records by type and name.
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
	parameter.AddToQuery(query, "id", req.ID)

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
	// DNSZone: DNS zone in which to update the DNS zone records.
	DNSZone string `json:"-"`
	// Changes: changes made to the records.
	Changes []*RecordChange `json:"changes"`
	// ReturnAllRecords: specifies whether or not to return all the records.
	ReturnAllRecords *bool `json:"return_all_records"`
	// DisallowNewZoneCreation: disable the creation of the target zone if it does not exist. Target zone creation is disabled by default.
	DisallowNewZoneCreation bool `json:"disallow_new_zone_creation"`
	// Serial: use the provided serial (0) instead of the auto-increment serial.
	Serial *uint64 `json:"serial"`
}

// UpdateDNSZoneRecords: update records within a DNS zone.
// Update records within a DNS zone that has default name servers and perform several actions on your records.
//
// Actions include:
//  - add: allows you to add a new record or add a new IP to an existing A record, for example
//  - set: allows you to edit a record or edit an IP from an existing A record, for example
//  - delete: allows you to delete a record or delete an IP from an existing A record, for example
//  - clear: allows you to delete all records from a DNS zone
//
// All edits will be versioned.
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
	// DNSZone: DNS zone on which to filter the returned DNS zone name servers.
	DNSZone string `json:"-"`
	// ProjectID: project ID on which to filter the returned DNS zone name servers.
	ProjectID *string `json:"-"`
}

// ListDNSZoneNameservers: list name servers within a DNS zone.
// Retrieve a list of name servers within a DNS zone and their optional glue records.
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
	// DNSZone: DNS zone in which to update the DNS zone name servers.
	DNSZone string `json:"-"`
	// Ns: new DNS zone name servers.
	Ns []*Nameserver `json:"ns"`
}

// UpdateDNSZoneNameservers: update name servers within a DNS zone.
// Update name servers within a DNS zone and set optional glue records.
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
	// DNSZone: DNS zone to clear.
	DNSZone string `json:"-"`
}

// ClearDNSZoneRecords: clear records within a DNS zone.
// Delete all records within a DNS zone that has default name servers.<br/>
// All edits will be versioned.
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
	// DNSZone: DNS zone to export.
	DNSZone string `json:"-"`
	// Format: DNS zone format.
	// Default value: bind
	Format RawFormat `json:"-"`
}

// ExportRawDNSZone: export a raw DNS zone.
// Export a DNS zone with default name servers, in a specific format.
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
	// DNSZone: DNS zone to import.
	DNSZone string `json:"-"`
	// Deprecated
	Content *string `json:"content,omitempty"`

	ProjectID string `json:"project_id"`
	// Deprecated: Format: default value: unknown_raw_format
	Format *RawFormat `json:"format,omitempty"`
	// BindSource: import a bind file format.
	// Precisely one of AxfrSource, BindSource must be set.
	BindSource *ImportRawDNSZoneRequestBindSource `json:"bind_source,omitempty"`
	// AxfrSource: import from the name server given with TSIG, to use or not.
	// Precisely one of AxfrSource, BindSource must be set.
	AxfrSource *ImportRawDNSZoneRequestAXFRSource `json:"axfr_source,omitempty"`
}

// ImportRawDNSZone: import a raw DNS zone.
// Import and replace the format of records from a given provider, with default name servers.
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

// ImportProviderDNSZone: import a DNS zone from another provider.
// Import and replace the format of records from a given provider, with default name servers.
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
	// DNSZone: DNS zone to refresh.
	DNSZone string `json:"-"`
	// RecreateDNSZone: specifies whether or not to recreate the DNS zone.
	RecreateDNSZone bool `json:"recreate_dns_zone"`
	// RecreateSubDNSZone: specifies whether or not to recreate the sub DNS zone.
	RecreateSubDNSZone bool `json:"recreate_sub_dns_zone"`
}

// RefreshDNSZone: refresh a DNS zone.
// Refresh an SOA DNS zone to reload the records in the DNS zone and update the SOA serial.
// You can recreate the given DNS zone and its sub DNS zone if needed.
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
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of DNS zones versions per page.
	PageSize *uint32 `json:"-"`
}

// ListDNSZoneVersions: list versions of a DNS zone.
// Retrieve a list of a DNS zone's versions.<br/>
// The maximum version count is 100. If the count reaches this limit, the oldest version will be deleted after each new modification.
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
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of DNS zones versions records per page.
	PageSize *uint32 `json:"-"`
}

// ListDNSZoneVersionRecords: list records from a given version of a specific DNS zone.
// Retrieve a list of records from a specific DNS zone version.
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

// GetDNSZoneVersionDiff: access differences from a specific DNS zone version.
// Access a previous DNS zone version to see the differences from another specific version.
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

// RestoreDNSZoneVersion: restore a DNS zone version.
// Restore and activate a version of a specific DNS zone.
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

// GetSSLCertificate: get a DNS zone's TLS certificate.
// Get the DNS zone's TLS certificate. If you do not have a certificate, the ouptut returns `no certificate found`.
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

// CreateSSLCertificate: create or get the DNS zone's TLS certificate.
// Create a new TLS certificate or retrieve information about an existing TLS certificate.
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

// ListSSLCertificates: list a user's TLS certificates.
// List all the TLS certificates a user has created, specified by the user's Project ID and the DNS zone.
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

// DeleteSSLCertificate: delete a TLS certificate.
// Delete an existing TLS certificate specified by its DNS zone. Deleting a TLS certificate is permanent and cannot be undone.
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

// GetDNSZoneTsigKey: get the DNS zone's TSIG key.
// Retrieve information about the TSIG key of a given DNS zone to allow AXFR requests.
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

// DeleteDNSZoneTsigKey: delete the DNS zone's TSIG key.
// Delete an existing TSIG key specified by its DNS zone. Deleting a TSIG key is permanent and cannot be undone.
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

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`

	Domain *string `json:"-"`

	Types []TaskType `json:"-"`

	Statuses []TaskStatus `json:"-"`
	// OrderBy: default value: domain_desc
	OrderBy ListTasksRequestOrderBy `json:"-"`
}

// ListTasks: list tasks.
// List all account tasks.
// You can filter the list by domain name.
func (s *RegistrarAPI) ListTasks(req *RegistrarAPIListTasksRequest, opts ...scw.RequestOption) (*ListTasksResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "domain", req.Domain)
	parameter.AddToQuery(query, "types", req.Types)
	parameter.AddToQuery(query, "statuses", req.Statuses)
	parameter.AddToQuery(query, "order_by", req.OrderBy)

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

// BuyDomains: buy one or more domains.
// Request the registration of domain names.
// You can provide an already existing domain's contact or a new contact.
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

// RenewDomains: renew one or more domains.
// Request the renewal of domain names.
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
	Domains []*TransferInDomainRequestTransferRequest `json:"domains"`

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

// TransferInDomain: transfer a domain.
// Request the transfer from another registrar domain to Scaleway.
func (s *RegistrarAPI) TransferInDomain(req *RegistrarAPITransferInDomainRequest, opts ...scw.RequestOption) (*OrderResponse, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/transfer-domains",
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

// TradeDomain: trade a domain contact.
// Request a trade for the contact owner.<br/>
// If an `organization_id` is given, the change is from the current Scaleway account to another Scaleway account.<br/>
// If no contact is given, the first contact of the other Scaleway account is taken.<br/>
// If the other Scaleway account has no contact. An error occurs.
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

// RegisterExternalDomain: register an external domain.
// Request the registration of an external domain name.
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

// DeleteExternalDomain: delete an external domain.
// Delete an external domain name.
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

type RegistrarAPICheckContactsCompatibilityRequest struct {
	Domains []string `json:"domains"`

	Tlds []string `json:"tlds"`

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

// CheckContactsCompatibility: check if contacts are compatible against a domain or a tld.
// Check if contacts are compatible against a domain or a tld.
// If not, it will return the information requiring a correction.
func (s *RegistrarAPI) CheckContactsCompatibility(req *RegistrarAPICheckContactsCompatibilityRequest, opts ...scw.RequestOption) (*CheckContactsCompatibilityResponse, error) {
	var err error

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/check-contacts-compatibility",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CheckContactsCompatibilityResponse

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

// ListContacts: list contacts.
// Return a list of contacts with their domains and roles.
// You can filter the list by domain name.
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

// GetContact: get a contact.
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
	// Lang: default value: unknown_language_code
	Lang LanguageCode `json:"lang"`

	Resale *bool `json:"resale"`
	// Deprecated
	Questions *[]*UpdateContactRequestQuestion `json:"questions,omitempty"`

	ExtensionFr *ContactExtensionFR `json:"extension_fr"`

	ExtensionEu *ContactExtensionEU `json:"extension_eu"`

	WhoisOptIn *bool `json:"whois_opt_in"`

	State *string `json:"state"`

	ExtensionNl *ContactExtensionNL `json:"extension_nl"`
}

// UpdateContact: update contact.
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
	// OrderBy: default value: domain_asc
	OrderBy ListDomainsRequestOrderBy `json:"-"`

	Registrar *string `json:"-"`
	// Status: default value: status_unknown
	Status DomainStatus `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`

	IsExternal *bool `json:"-"`

	Domain *string `json:"-"`
}

// ListDomains: list domains.
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
	parameter.AddToQuery(query, "domain", req.Domain)

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
	// OrderBy: default value: domain_asc
	OrderBy ListRenewableDomainsRequestOrderBy `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListRenewableDomains: list scaleway domains that can or not be renewed.
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

// GetDomain: get domain.
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

// UpdateDomain: update a domain.
// Update the domain contacts or create a new one.<br/>
// If you add the same contact for multiple roles. Only one ID will be created and used for all of them.
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

// LockDomainTransfer: lock domain transfer.
// Lock domain transfer. A locked domain transfer can't be transferred and the auth code can't be requested.
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

// UnlockDomainTransfer: unlock domain transfer.
// Unlock domain transfer. An unlocked domain can be transferred and the auth code can be requested for this.
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

// EnableDomainAutoRenew: enable domain auto renew.
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

// DisableDomainAutoRenew: disable domain auto renew.
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

// GetDomainAuthCode: return domain auth code.
// If possible, return the auth code for an unlocked domain transfer, or an error if the domain is locked.
// Some TLD may have a different procedure to retrieve the auth code, in that case, the information is given in the message field.
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

// EnableDomainDNSSEC: update domain DNSSEC.
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
//   - 4: SHA_384.
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

// DisableDomainDNSSEC: disable domain DNSSEC.
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
	// Domains: a list of domain to search, TLD is optional.
	Domains []string `json:"-"`
	// Tlds: array of tlds to search on.
	Tlds []string `json:"-"`
	// StrictSearch: search exact match.
	StrictSearch bool `json:"-"`
}

// SearchAvailableDomains: search available domains.
// Search a domain (or at maximum, 10 domains).
//
// If the TLD list is empty or not set the search returns the results from the most popular TLDs.
func (s *RegistrarAPI) SearchAvailableDomains(req *RegistrarAPISearchAvailableDomainsRequest, opts ...scw.RequestOption) (*SearchAvailableDomainsResponse, error) {
	var err error

	query := url.Values{}
	parameter.AddToQuery(query, "domains", req.Domains)
	parameter.AddToQuery(query, "tlds", req.Tlds)
	parameter.AddToQuery(query, "strict_search", req.StrictSearch)

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

type RegistrarAPICreateDomainHostRequest struct {
	Domain string `json:"-"`

	Name string `json:"name"`

	IPs []net.IP `json:"ips"`
}

// CreateDomainHost: create domain hostname with glue IPs.
func (s *RegistrarAPI) CreateDomainHost(req *RegistrarAPICreateDomainHostRequest, opts ...scw.RequestOption) (*Host, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/hosts",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Host

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIListDomainHostsRequest struct {
	Domain string `json:"-"`

	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`
}

// ListDomainHosts: list domain hostnames with they glue IPs.
func (s *RegistrarAPI) ListDomainHosts(req *RegistrarAPIListDomainHostsRequest, opts ...scw.RequestOption) (*ListDomainHostsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/hosts",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDomainHostsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIUpdateDomainHostRequest struct {
	Domain string `json:"-"`

	Name string `json:"-"`

	IPs *[]string `json:"ips"`
}

// UpdateDomainHost: update domain hostname with glue IPs.
func (s *RegistrarAPI) UpdateDomainHost(req *RegistrarAPIUpdateDomainHostRequest, opts ...scw.RequestOption) (*Host, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	if fmt.Sprint(req.Name) == "" {
		return nil, errors.New("field Name cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/hosts/" + fmt.Sprint(req.Name) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Host

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RegistrarAPIDeleteDomainHostRequest struct {
	Domain string `json:"-"`

	Name string `json:"-"`
}

// DeleteDomainHost: delete domain hostname.
func (s *RegistrarAPI) DeleteDomainHost(req *RegistrarAPIDeleteDomainHostRequest, opts ...scw.RequestOption) (*Host, error) {
	var err error

	if fmt.Sprint(req.Domain) == "" {
		return nil, errors.New("field Domain cannot be empty in request")
	}

	if fmt.Sprint(req.Name) == "" {
		return nil, errors.New("field Name cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/domain/v2beta1/domains/" + fmt.Sprint(req.Domain) + "/hosts/" + fmt.Sprint(req.Name) + "",
		Headers: http.Header{},
	}

	var resp Host

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

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDomainHostsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDomainHostsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDomainHostsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Hosts = append(r.Hosts, results.Hosts...)
	r.TotalCount += uint32(len(results.Hosts))
	return uint32(len(results.Hosts)), nil
}
