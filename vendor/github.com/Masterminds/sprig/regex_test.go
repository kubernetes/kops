package sprig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexMatch(t *testing.T) {
	regex := "[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}"

	assert.True(t, regexMatch(regex, "test@acme.com"))
	assert.True(t, regexMatch(regex, "Test@Acme.Com"))
	assert.False(t, regexMatch(regex, "test"))
	assert.False(t, regexMatch(regex, "test.com"))
	assert.False(t, regexMatch(regex, "test@acme"))
}

func TestRegexFindAll(t *testing.T){
	regex := "a{2}"
	assert.Equal(t, 1, len(regexFindAll(regex, "aa", -1)))
	assert.Equal(t, 1, len(regexFindAll(regex, "aaaaaaaa", 1)))
	assert.Equal(t, 2, len(regexFindAll(regex, "aaaa", -1)))
	assert.Equal(t, 0, len(regexFindAll(regex, "none", -1)))
}

func TestRegexFindl(t *testing.T){
	regex := "fo.?"
	assert.Equal(t, "foo", regexFind(regex, "foorbar"))
	assert.Equal(t, "foo", regexFind(regex, "foo foe fome"))
	assert.Equal(t, "", regexFind(regex, "none"))
}

func TestRegexReplaceAll(t *testing.T){
	regex := "a(x*)b"
	assert.Equal(t, "-T-T-", regexReplaceAll(regex,"-ab-axxb-", "T"))
	assert.Equal(t, "--xx-", regexReplaceAll(regex,"-ab-axxb-", "$1"))
	assert.Equal(t, "---", regexReplaceAll(regex,"-ab-axxb-", "$1W"))
	assert.Equal(t, "-W-xxW-", regexReplaceAll(regex,"-ab-axxb-", "${1}W"))
}

func TestRegexReplaceAllLiteral(t *testing.T){
	regex := "a(x*)b"
	assert.Equal(t, "-T-T-", regexReplaceAllLiteral(regex,"-ab-axxb-", "T"))
	assert.Equal(t, "-$1-$1-", regexReplaceAllLiteral(regex,"-ab-axxb-", "$1"))
	assert.Equal(t, "-${1}-${1}-", regexReplaceAllLiteral(regex,"-ab-axxb-", "${1}"))
}

func TestRegexSplit(t *testing.T){
	regex := "a"
	assert.Equal(t, 4, len(regexSplit(regex,"banana", -1)))
	assert.Equal(t, 0, len(regexSplit(regex,"banana", 0)))
	assert.Equal(t, 1, len(regexSplit(regex,"banana", 1)))
	assert.Equal(t, 2, len(regexSplit(regex,"banana", 2)))

	regex = "z+"
	assert.Equal(t, 2, len(regexSplit(regex,"pizza", -1)))
	assert.Equal(t, 0, len(regexSplit(regex,"pizza", 0)))
	assert.Equal(t, 1, len(regexSplit(regex,"pizza", 1)))
	assert.Equal(t, 2, len(regexSplit(regex,"pizza", 2)))
}