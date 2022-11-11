package scw

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// localityPartsSeparator is the separator used in Zone and Region
const localityPartsSeparator = "-"

// Zone is an availability zone
type Zone string

const (
	// ZoneFrPar1 represents the fr-par-1 zone
	ZoneFrPar1 = Zone("fr-par-1")
	// ZoneFrPar2 represents the fr-par-2 zone
	ZoneFrPar2 = Zone("fr-par-2")
	// ZoneFrPar3 represents the fr-par-3 zone
	ZoneFrPar3 = Zone("fr-par-3")
	// ZoneNlAms1 represents the nl-ams-1 zone
	ZoneNlAms1 = Zone("nl-ams-1")
	// ZoneNlAms2 represents the nl-ams-2 zone
	ZoneNlAms2 = Zone("nl-ams-2")
	// ZonePlWaw1 represents the pl-waw-1 zone
	ZonePlWaw1 = Zone("pl-waw-1")
	// ZonePlWaw2 represents the pl-waw-2 zone
	ZonePlWaw2 = Zone("pl-waw-2")
)

var (
	// AllZones is an array that list all zones
	AllZones = []Zone{
		ZoneFrPar1,
		ZoneFrPar2,
		ZoneFrPar3,
		ZoneNlAms1,
		ZoneNlAms2,
		ZonePlWaw1,
		ZonePlWaw2,
	}
)

// Exists checks whether a zone exists
func (zone Zone) Exists() bool {
	for _, z := range AllZones {
		if z == zone {
			return true
		}
	}
	return false
}

// String returns a Zone as a string
func (zone Zone) String() string {
	return string(zone)
}

// Region returns the parent Region for the Zone.
// Manipulates the string directly to allow unlisted zones formatted as xx-yyy-z.
func (zone Zone) Region() (Region, error) {
	zoneStr := zone.String()
	if !validation.IsZone(zoneStr) {
		return "", fmt.Errorf("invalid zone '%v'", zoneStr)
	}
	zoneParts := strings.Split(zoneStr, localityPartsSeparator)
	return Region(strings.Join(zoneParts[:2], localityPartsSeparator)), nil
}

// Region is a geographical location
type Region string

const (
	// RegionFrPar represents the fr-par region
	RegionFrPar = Region("fr-par")
	// RegionNlAms represents the nl-ams region
	RegionNlAms = Region("nl-ams")
	// RegionPlWaw represents the pl-waw region
	RegionPlWaw = Region("pl-waw")
)

var (
	// AllRegions is an array that list all regions
	AllRegions = []Region{
		RegionFrPar,
		RegionNlAms,
		RegionPlWaw,
	}
)

// Exists checks whether a region exists
func (region Region) Exists() bool {
	for _, r := range AllRegions {
		if r == region {
			return true
		}
	}
	return false
}

// GetZones is a function that returns the zones for the specified region
func (region Region) GetZones() []Zone {
	switch region {
	case RegionFrPar:
		return []Zone{ZoneFrPar1, ZoneFrPar2, ZoneFrPar3}
	case RegionNlAms:
		return []Zone{ZoneNlAms1, ZoneNlAms2}
	case RegionPlWaw:
		return []Zone{ZonePlWaw1, ZonePlWaw2}
	default:
		return []Zone{}
	}
}

// ParseZone parses a string value into a Zone and returns an error if it has a bad format.
func ParseZone(zone string) (Zone, error) {
	switch zone {
	case "par1":
		// would be triggered by API market place
		// logger.Warningf("par1 is a deprecated name for zone, use fr-par-1 instead")
		return ZoneFrPar1, nil
	case "ams1":
		// would be triggered by API market place
		// logger.Warningf("ams1 is a deprecated name for zone, use nl-ams-1 instead")
		return ZoneNlAms1, nil
	default:
		if !validation.IsZone(zone) {
			zones := []string(nil)
			for _, z := range AllZones {
				zones = append(zones, string(z))
			}
			return "", errors.New("bad zone format, available zones are: %s", strings.Join(zones, ", "))
		}

		newZone := Zone(zone)
		if !newZone.Exists() {
			logger.Infof("%s is an unknown zone\n", newZone)
		}
		return newZone, nil
	}
}

// UnmarshalJSON implements the Unmarshaler interface for a Zone.
// this to call ParseZone on the string input and return the correct Zone object.
func (zone *Zone) UnmarshalJSON(input []byte) error {
	// parse input value as string
	var stringValue string
	err := json.Unmarshal(input, &stringValue)
	if err != nil {
		return err
	}

	// parse string as Zone
	*zone, err = ParseZone(stringValue)
	if err != nil {
		return err
	}
	return nil
}

// ParseRegion parses a string value into a Region and returns an error if it has a bad format.
func ParseRegion(region string) (Region, error) {
	switch region {
	case "par1":
		// would be triggered by API market place
		// logger.Warningf("par1 is a deprecated name for region, use fr-par instead")
		return RegionFrPar, nil
	case "ams1":
		// would be triggered by API market place
		// logger.Warningf("ams1 is a deprecated name for region, use nl-ams instead")
		return RegionNlAms, nil
	default:
		if !validation.IsRegion(region) {
			regions := []string(nil)
			for _, r := range AllRegions {
				regions = append(regions, string(r))
			}
			return "", errors.New("bad region format, available regions are: %s", strings.Join(regions, ", "))
		}

		newRegion := Region(region)
		if !newRegion.Exists() {
			logger.Infof("%s is an unknown region\n", newRegion)
		}
		return newRegion, nil
	}
}

// UnmarshalJSON implements the Unmarshaler interface for a Region.
// this to call ParseRegion on the string input and return the correct Region object.
func (region *Region) UnmarshalJSON(input []byte) error {
	// parse input value as string
	var stringValue string
	err := json.Unmarshal(input, &stringValue)
	if err != nil {
		return err
	}

	// parse string as Region
	*region, err = ParseRegion(stringValue)
	if err != nil {
		return err
	}
	return nil
}

// String returns a Region as a string
func (region Region) String() string {
	return string(region)
}
