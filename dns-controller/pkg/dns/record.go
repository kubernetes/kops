package dns

type RecordType string

const (
	// RecordTypeAlias is unusual: the controller will try to resolve the target locally
	RecordTypeAlias = "_alias"

	RecordTypeA     = "A"
	RecordTypeCNAME = "CNAME"
)

type Record struct {
	RecordType RecordType
	FQDN       string
	Value      string

	// If AliasTarget is set, this entry will not actually be set in DNS,
	// but will be used as an expansion for Records with type=RecordTypeAlias,
	// where the referring record has Value = our FQDN
	AliasTarget bool
}
