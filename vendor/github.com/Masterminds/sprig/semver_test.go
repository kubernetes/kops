package sprig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemverCompare(t *testing.T) {
	tests := map[string]string{
		`{{ semverCompare "1.2.3" "1.2.3" }}`:  `true`,
		`{{ semverCompare "^1.2.0" "1.2.3" }}`: `true`,
		`{{ semverCompare "^1.2.0" "2.2.3" }}`: `false`,
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestSemver(t *testing.T) {
	tests := map[string]string{
		`{{ $s := semver "1.2.3-beta.1+c0ff33" }}{{ $s.Prerelease }}`: "beta.1",
		`{{ $s := semver "1.2.3-beta.1+c0ff33" }}{{ $s.Major}}`:       "1",
		`{{ semver "1.2.3" | (semver "1.2.3").Compare }}`:             `0`,
		`{{ semver "1.2.3" | (semver "1.3.3").Compare }}`:             `1`,
		`{{ semver "1.4.3" | (semver "1.2.3").Compare }}`:             `-1`,
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
