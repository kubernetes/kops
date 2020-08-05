package featureflag

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// All registered feature flags.
var (
	flagsMutex sync.Mutex
	flags      = make(map[string]FeatureFlag)
)

// FeatureFlag indicates whether a given feature is enabled or not.
type FeatureFlag interface {
	fmt.Stringer

	// Name returns the name of the feature flag.
	Name() string

	// Enabled returns true if the feature is enabled.
	Enabled() bool
}

// featureFlag represents a feature being gated.
type featureFlag struct {
	name    string
	enabled bool
}

// New returns a new feature flag.
func New(name string, enabled bool) FeatureFlag {
	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	ff, ok := flags[name]
	if !ok {
		ff = &featureFlag{name: name}
		flags[name] = ff
	}

	ff.(*featureFlag).enabled = enabled
	return ff
}

// Name returns the name of the feature flag.
func (f *featureFlag) Name() string { return f.name }

// Enabled returns true if the feature is enabled.
func (f *featureFlag) Enabled() bool { return f.enabled }

// String returns the string representation of the feature flag.
func (f *featureFlag) String() string { return fmt.Sprintf("%s=%t", f.name, f.enabled) }

// Set parses and stores features from a string like "feature1=true,feature2=false".
func Set(features string) {
	for _, s := range strings.Split(strings.TrimSpace(features), ",") {
		if len(s) == 0 {
			continue
		}

		segments := strings.SplitN(s, "=", 2)
		name := strings.TrimSpace(segments[0])

		enabled := true
		if len(segments) > 1 {
			value := strings.TrimSpace(segments[1])
			enabled, _ = strconv.ParseBool(value) // ignore errors and fallback to `false`
		}

		New(name, enabled)
	}
}

// Get returns a specific feature flag by name.
func Get(name string) FeatureFlag {
	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	f, ok := flags[name]
	if !ok {
		f = new(featureFlag)
	}

	return &featureFlag{
		name:    name,
		enabled: f.Enabled(),
	}
}

// All returns a list of all known feature flags.
func All() FeatureFlags {
	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	features := make(FeatureFlags, 0, len(flags))
	for name, flag := range flags {
		features = append(features, &featureFlag{
			name:    name,
			enabled: flag.Enabled(),
		})
	}

	return features
}

// FeatureFlags defines a list of feature flags.
type FeatureFlags []FeatureFlag

// String returns the string representation of a list of feature flags.
func (f FeatureFlags) String() string {
	features := make([]string, len(f))
	for i, ff := range f {
		features[i] = ff.String()
	}
	return strings.Join(features, ",")
}
