package fi

type CloudProviderID string

const CloudProviderAWS CloudProviderID = "aws"
const CloudProviderGCE CloudProviderID = "gce"

type Cloud interface {
	ProviderID() CloudProviderID
}
