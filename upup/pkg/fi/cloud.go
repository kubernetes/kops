package fi

type CloudProviderID string

const CloudProviderAWS CloudProviderID = "aws"
const CloudProviderGCE CloudProviderID = "gce"

type Cloud interface {
	ProviderID() CloudProviderID

	FindDNSHostedZone(dnsName string) (string, error)
}

// zonesToCloud allows us to infer from certain well-known zones to a cloud
// Note it is safe to "overmap" zones that don't exist: we'll check later if the zones actually exist
var zonesToCloud = map[string]CloudProviderID{
	"us-east-1a": CloudProviderAWS,
	"us-east-1b": CloudProviderAWS,
	"us-east-1c": CloudProviderAWS,
	"us-east-1d": CloudProviderAWS,
	"us-east-1e": CloudProviderAWS,

	"us-west-1a": CloudProviderAWS,
	"us-west-1b": CloudProviderAWS,
	"us-west-1c": CloudProviderAWS,
	"us-west-1d": CloudProviderAWS,
	"us-west-1e": CloudProviderAWS,

	"us-west-2a": CloudProviderAWS,
	"us-west-2b": CloudProviderAWS,
	"us-west-2c": CloudProviderAWS,
	"us-west-2d": CloudProviderAWS,
	"us-west-2e": CloudProviderAWS,

	"eu-west-1a": CloudProviderAWS,
	"eu-west-1b": CloudProviderAWS,
	"eu-west-1c": CloudProviderAWS,
	"eu-west-1d": CloudProviderAWS,
	"eu-west-1e": CloudProviderAWS,

	"eu-central-1a": CloudProviderAWS,
	"eu-central-1b": CloudProviderAWS,
	"eu-central-1c": CloudProviderAWS,
	"eu-central-1d": CloudProviderAWS,
	"eu-central-1e": CloudProviderAWS,

	"ap-south-1a": CloudProviderAWS,
	"ap-south-1b": CloudProviderAWS,
	"ap-south-1c": CloudProviderAWS,
	"ap-south-1d": CloudProviderAWS,
	"ap-south-1e": CloudProviderAWS,

	"ap-southeast-1a": CloudProviderAWS,
	"ap-southeast-1b": CloudProviderAWS,
	"ap-southeast-1c": CloudProviderAWS,
	"ap-southeast-1d": CloudProviderAWS,
	"ap-southeast-1e": CloudProviderAWS,

	"ap-southeast-2a": CloudProviderAWS,
	"ap-southeast-2b": CloudProviderAWS,
	"ap-southeast-2c": CloudProviderAWS,
	"ap-southeast-2d": CloudProviderAWS,
	"ap-southeast-2e": CloudProviderAWS,

	"ap-northeast-1a": CloudProviderAWS,
	"ap-northeast-1b": CloudProviderAWS,
	"ap-northeast-1c": CloudProviderAWS,
	"ap-northeast-1d": CloudProviderAWS,
	"ap-northeast-1e": CloudProviderAWS,

	"ap-northeast-2a": CloudProviderAWS,
	"ap-northeast-2b": CloudProviderAWS,
	"ap-northeast-2c": CloudProviderAWS,
	"ap-northeast-2d": CloudProviderAWS,
	"ap-northeast-2e": CloudProviderAWS,

	"sa-east-1a": CloudProviderAWS,
	"sa-east-1b": CloudProviderAWS,
	"sa-east-1c": CloudProviderAWS,
	"sa-east-1d": CloudProviderAWS,
	"sa-east-1e": CloudProviderAWS,
}

// GuessCloudForZone tries to infer the cloudprovider from the zone name
func GuessCloudForZone(zone string) (CloudProviderID, bool) {
	c, found := zonesToCloud[zone]
	return c, found
}
