package fi

type ProviderID string

const ProviderAWS ProviderID = "aws"
const ProviderGCE ProviderID = "gce"

type Cloud interface {
	ProviderID() ProviderID
}
