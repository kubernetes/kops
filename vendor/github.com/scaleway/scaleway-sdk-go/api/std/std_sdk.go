// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package std provides methods and message types of the std  API.
package std

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

type LanguageCode string

const (
	LanguageCodeUnknownLanguageCode = LanguageCode("unknown_language_code")
	LanguageCodeEnUS                = LanguageCode("en_US")
	LanguageCodeFrFR                = LanguageCode("fr_FR")
	LanguageCodeDeDE                = LanguageCode("de_DE")
)

func (enum LanguageCode) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_language_code"
	}
	return string(enum)
}

func (enum LanguageCode) Values() []LanguageCode {
	return []LanguageCode{
		"unknown_language_code",
		"en_US",
		"fr_FR",
		"de_DE",
	}
}

func (enum LanguageCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LanguageCode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LanguageCode(LanguageCode(tmp).String())
	return nil
}
