package gossip

type SeedProvider interface {
	GetSeeds() ([]string, error)
}

func NewStaticSeedProvider(seeds []string) *StaticSeedProvider {
	return &StaticSeedProvider{Seeds: seeds}
}

type StaticSeedProvider struct {
	Seeds []string
}

var _ SeedProvider = &StaticSeedProvider{}

func (s *StaticSeedProvider) GetSeeds() ([]string, error) {
	return s.Seeds, nil
}
