package scw

import (
	"os"
	"strconv"

	"github.com/scaleway/scaleway-sdk-go/logger"
)

// Environment variables
const (
	// Up-to-date
	ScwCacheDirEnv              = "SCW_CACHE_DIR"
	ScwConfigPathEnv            = "SCW_CONFIG_PATH"
	ScwAccessKeyEnv             = "SCW_ACCESS_KEY"
	ScwSecretKeyEnv             = "SCW_SECRET_KEY" // #nosec G101
	ScwActiveProfileEnv         = "SCW_PROFILE"
	ScwAPIURLEnv                = "SCW_API_URL"
	ScwS3EndpointEnv            = "SCW_S3_ENDPOINT"
	ScwInsecureEnv              = "SCW_INSECURE"
	ScwDefaultOrganizationIDEnv = "SCW_DEFAULT_ORGANIZATION_ID"
	ScwDefaultProjectIDEnv      = "SCW_DEFAULT_PROJECT_ID"
	ScwDefaultRegionEnv         = "SCW_DEFAULT_REGION"
	ScwDefaultZoneEnv           = "SCW_DEFAULT_ZONE"
	ScwEnableBeta               = "SCW_ENABLE_BETA"
	DebugEnv                    = logger.DebugEnv

	// AWS
	AwsEndpointURL   = "AWS_ENDPOINT_URL"
	AwsEndpointURLS3 = "AWS_ENDPOINT_URL_S3"

	// All deprecated (cli&terraform)
	terraformAccessKeyEnv    = "SCALEWAY_ACCESS_KEY" // used both as access key and secret key
	terraformSecretKeyEnv    = "SCALEWAY_TOKEN"
	terraformOrganizationEnv = "SCALEWAY_ORGANIZATION"
	terraformRegionEnv       = "SCALEWAY_REGION"
	cliTLSVerifyEnv          = "SCW_TLSVERIFY"
	cliOrganizationEnv       = "SCW_ORGANIZATION"
	cliRegionEnv             = "SCW_REGION"
	cliSecretKeyEnv          = "SCW_TOKEN"

	// TBD
	// cliVerboseEnv         = "SCW_VERBOSE_API"
	// cliDebugEnv           = "DEBUG"
	// cliNoCheckVersionEnv  = "SCW_NOCHECKVERSION"
	// cliTestWithRealAPIEnv = "TEST_WITH_REAL_API"
	// cliSecureExecEnv      = "SCW_SECURE_EXEC"
	// cliGatewayEnv         = "SCW_GATEWAY"
	// cliSensitiveEnv       = "SCW_SENSITIVE"
	// cliAccountAPIEnv      = "SCW_ACCOUNT_API"
	// cliMetadataAPIEnv     = "SCW_METADATA_API"
	// cliMarketPlaceAPIEnv  = "SCW_MARKETPLACE_API"
	// cliComputePar1APIEnv  = "SCW_COMPUTE_PAR1_API"
	// cliComputeAms1APIEnv  = "SCW_COMPUTE_AMS1_API"
	// cliCommercialTypeEnv  = "SCW_COMMERCIAL_TYPE"
	// cliTargetArchEnv      = "SCW_TARGET_ARCH"
)

const (
	v1RegionFrPar = "par1"
	v1RegionNlAms = "ams1"
)

func LoadEnvProfile() *Profile {
	p := &Profile{}

	accessKey, _, envExist := getEnv(ScwAccessKeyEnv, terraformAccessKeyEnv)
	if envExist {
		p.AccessKey = &accessKey
	}

	secretKey, _, envExist := getEnv(ScwSecretKeyEnv, cliSecretKeyEnv, terraformSecretKeyEnv, terraformAccessKeyEnv)
	if envExist {
		p.SecretKey = &secretKey
	}

	apiURL, _, envExist := getEnv(ScwAPIURLEnv)
	if envExist {
		p.APIURL = &apiURL
	}

	s3Endpoint, _, envExist := getEnv(ScwS3EndpointEnv)
	if envExist {
		p.S3Endpoint = &s3Endpoint
	}

	insecureValue, envKey, envExist := getEnv(ScwInsecureEnv, cliTLSVerifyEnv)
	if envExist {
		insecure, err := strconv.ParseBool(insecureValue)
		if err != nil {
			logger.Warningf("env variable %s cannot be parsed: %s is invalid boolean", envKey, insecureValue)
		}

		if envKey == cliTLSVerifyEnv {
			insecure = !insecure // TLSVerify is the inverse of Insecure
		}

		p.Insecure = &insecure
	}

	organizationID, _, envExist := getEnv(ScwDefaultOrganizationIDEnv, cliOrganizationEnv, terraformOrganizationEnv)
	if envExist {
		p.DefaultOrganizationID = &organizationID
	}

	projectID, _, envExist := getEnv(ScwDefaultProjectIDEnv)
	if envExist {
		p.DefaultProjectID = &projectID
	}

	region, _, envExist := getEnv(ScwDefaultRegionEnv, cliRegionEnv, terraformRegionEnv)
	if envExist {
		region = v1RegionToV2(region)
		p.DefaultRegion = &region
	}

	zone, _, envExist := getEnv(ScwDefaultZoneEnv)
	if envExist {
		p.DefaultZone = &zone
	}

	return p
}

// GetS3EndpointFromAWSConf retrieves the set value of AWS_ENDPOINT_URL_S3
// or, if not set, AWS_ENDPOINT_URL.
// This function can be called from any client side code which intends to
// be AWS compatible, thus check the environment variables.
// In case AWS changes the key of these variable, this function should be the
// single point to update.
func GetS3EndpointFromAWSConf() string {
	if ep := os.Getenv(AwsEndpointURLS3); ep != "" {
		return ep
	}

	return os.Getenv(AwsEndpointURL)
}

func getEnv(upToDateKey string, deprecatedKeys ...string) (string, string, bool) {
	value, exist := os.LookupEnv(upToDateKey)
	if exist {
		logger.Debugf("reading value from %s\n", upToDateKey)
		return value, upToDateKey, true
	}

	for _, key := range deprecatedKeys {
		value, exist := os.LookupEnv(key)
		if exist {
			logger.Debugf("reading value from %s\n", key)
			logger.Warningf("%s is deprecated, please use %s instead\n", key, upToDateKey)
			return value, key, true
		}
	}

	return "", "", false
}

func v1RegionToV2(region string) string {
	switch region {
	case v1RegionFrPar:
		logger.Warningf("par1 is a deprecated name for region, use fr-par instead")
		return "fr-par"
	case v1RegionNlAms:
		logger.Warningf("ams1 is a deprecated name for region, use nl-ams instead")
		return "nl-ams"
	default:
		return region
	}
}
