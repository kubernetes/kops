package scw

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
)

// ServiceInfo contains API metadata
// These metadata are only here for debugging. Do not rely on these values
type ServiceInfo struct {
	// Name is the name of the API
	Name string `json:"name"`

	// Description is a human readable description for the API
	Description string `json:"description"`

	// Version is the version of the API
	Version string `json:"version"`

	// DocumentationURL is the a web url where the documentation of the API can be found
	DocumentationURL *string `json:"documentation_url"`
}

// File is the structure used to receive / send a file from / to the API
type File struct {
	// Name of the file
	Name string `json:"name"`

	// ContentType used in the HTTP header `Content-Type`
	ContentType string `json:"content_type"`

	// Content of the file
	Content io.Reader `json:"content"`
}

func (f *File) UnmarshalJSON(b []byte) error {
	type file File
	var tmpFile struct {
		file
		Content []byte `json:"content"`
	}

	err := json.Unmarshal(b, &tmpFile)
	if err != nil {
		return err
	}

	tmpFile.file.Content = bytes.NewReader(tmpFile.Content)

	*f = File(tmpFile.file)
	return nil
}

// Money represents an amount of money with its currency type.
type Money struct {
	// CurrencyCode is the 3-letter currency code defined in ISO 4217.
	CurrencyCode string `json:"currency_code"`

	// Units is the whole units of the amount.
	// For example if `currencyCode` is `"USD"`, then 1 unit is one US dollar.
	Units int64 `json:"units"`

	// Nanos is the number of nano (10^-9) units of the amount.
	// The value must be between -999,999,999 and +999,999,999 inclusive.
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int32 `json:"nanos"`
}

// NewMoneyFromFloat converts a float with currency to a Money.
//
// value:        The float value.
// currencyCode: The 3-letter currency code defined in ISO 4217.
// precision:    The number of digits after the decimal point used to parse the nanos part of the value.
//
// Examples:
// - (value = 1.3333, precision = 2) => Money{Units = 1, Nanos = 330000000}
// - (value = 1.123456789, precision = 9) => Money{Units = 1, Nanos = 123456789}
func NewMoneyFromFloat(value float64, currencyCode string, precision int) *Money {
	if precision > 9 {
		panic(fmt.Errorf("max precision is 9"))
	}

	strValue := strconv.FormatFloat(value, 'f', precision, 64)
	units, nanos, err := splitFloatString(strValue)
	if err != nil {
		panic(err)
	}

	return &Money{
		CurrencyCode: currencyCode,
		Units:        units,
		Nanos:        nanos,
	}
}

// String returns the string representation of Money.
func (m Money) String() string {
	currencySignsByCodes := map[string]string{
		"EUR": "€",
		"USD": "$",
	}

	currencySign, currencySignFound := currencySignsByCodes[m.CurrencyCode]
	if !currencySignFound {
		logger.Debugf("%s currency code is not supported", m.CurrencyCode)
		currencySign = m.CurrencyCode
	}

	cents := fmt.Sprintf("%09d", m.Nanos)
	cents = cents[:2] + strings.TrimRight(cents[2:], "0")

	return fmt.Sprintf("%s %d.%s", currencySign, m.Units, cents)
}

// ToFloat converts a Money object to a float.
func (m Money) ToFloat() float64 {
	return float64(m.Units) + float64(m.Nanos)/1e9
}

// Size represents a size in bytes.
type Size uint64

const (
	B  Size = 1
	KB      = 1000 * B
	MB      = 1000 * KB
	GB      = 1000 * MB
	TB      = 1000 * GB
	PB      = 1000 * TB
)

// String returns the string representation of a Size.
func (s Size) String() string {
	return fmt.Sprintf("%d", s)
}

// TimeSeries represents a time series that could be used for graph purposes.
type TimeSeries struct {
	// Name of the metric.
	Name string `json:"name"`

	// Points contains all the points that composed the series.
	Points []*TimeSeriesPoint `json:"points"`

	// Metadata contains some string metadata related to a metric.
	Metadata map[string]string `json:"metadata"`
}

// TimeSeriesPoint represents a point of a time series.
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float32
}

func (tsp TimeSeriesPoint) MarshalJSON() ([]byte, error) {
	timestamp := tsp.Timestamp.Format(time.RFC3339)
	value, err := json.Marshal(tsp.Value)
	if err != nil {
		return nil, err
	}

	return []byte(`["` + timestamp + `",` + string(value) + "]"), nil
}

func (tsp *TimeSeriesPoint) UnmarshalJSON(b []byte) error {
	point := [2]interface{}{}

	err := json.Unmarshal(b, &point)
	if err != nil {
		return err
	}

	if len(point) != 2 {
		return fmt.Errorf("invalid point array")
	}

	strTimestamp, isStrTimestamp := point[0].(string)
	if !isStrTimestamp {
		return fmt.Errorf("%s timestamp is not a string in RFC 3339 format", point[0])
	}
	timestamp, err := time.Parse(time.RFC3339, strTimestamp)
	if err != nil {
		return fmt.Errorf("%s timestamp is not in RFC 3339 format", point[0])
	}
	tsp.Timestamp = timestamp

	// By default, JSON unmarshal a float in float64 but the TimeSeriesPoint is a float32 value.
	value, isValue := point[1].(float64)
	if !isValue {
		return fmt.Errorf("%s is not a valid float32 value", point[1])
	}
	tsp.Value = float32(value)

	return nil
}

// IPNet inherits net.IPNet and represents an IP network.
type IPNet struct {
	net.IPNet
}

func (n IPNet) MarshalJSON() ([]byte, error) {
	value := n.String()
	if value == "<nil>" {
		value = ""
	}
	return []byte(`"` + value + `"`), nil
}

func (n *IPNet) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if str == "" {
		*n = IPNet{}
		return nil
	}

	switch ip := net.ParseIP(str); {
	case ip.To4() != nil:
		str += "/32"
	case ip.To16() != nil:
		str += "/128"
	}

	ip, value, err := net.ParseCIDR(str)
	if err != nil {
		return err
	}
	value.IP = ip
	n.IPNet = *value

	return nil
}

// Duration represents a signed, fixed-length span of time represented as a
// count of seconds and fractions of seconds at nanosecond resolution. It is
// independent of any calendar and concepts like "day" or "month". It is related
// to Timestamp in that the difference between two Timestamp values is a Duration
// and it can be added or subtracted from a Timestamp.
// Range is approximately +-10,000 years.
type Duration struct {
	Seconds int64
	Nanos   int32
}

func (d *Duration) ToTimeDuration() *time.Duration {
	if d == nil {
		return nil
	}
	timeDuration := time.Duration(d.Nanos) + time.Duration(d.Seconds/1e9)
	return &timeDuration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	nanos := d.Nanos
	if nanos < 0 {
		nanos = -nanos
	}

	return []byte(`"` + fmt.Sprintf("%d.%09d", d.Seconds, nanos) + `s"`), nil
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var str string

	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if str == "" {
		*d = Duration{}
		return nil
	}

	seconds, nanos, err := splitFloatString(strings.TrimRight(str, "s"))
	if err != nil {
		return err
	}

	*d = Duration{
		Seconds: seconds,
		Nanos:   nanos,
	}

	return nil
}

// splitFloatString splits a float represented in a string, and returns its units (left-coma part) and nanos (right-coma part).
// E.g.:
// "3"     ==> units = 3  | nanos = 0
// "3.14"  ==> units = 3  | nanos = 14*1e7
// "-3.14" ==> units = -3 | nanos = -14*1e7
func splitFloatString(input string) (units int64, nanos int32, err error) {
	parts := strings.SplitN(input, ".", 2)

	// parse units as int64
	units, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "invalid units")
	}

	// handle nanos
	if len(parts) == 2 {
		// add leading zeros
		strNanos := parts[1] + "000000000"[len(parts[1]):]

		// parse nanos as int32
		n, err := strconv.ParseUint(strNanos, 10, 32)
		if err != nil {
			return 0, 0, errors.Wrap(err, "invalid nanos")
		}

		nanos = int32(n)
	}

	if units < 0 {
		nanos = -nanos
	}

	return units, nanos, nil
}

// JSONObject represent any JSON object. See struct.proto.
// It will be marshaled into a json string.
// This type can be used just like any other map.
//
//	Example:
//
//	values := scw.JSONValue{
//		"Foo": "Bar",
//	}
//	values["Baz"] = "Qux"
type JSONObject map[string]interface{}

// EscapeMode is the mode that should be use for escaping a value
type EscapeMode uint

// The modes for escaping a value before it is marshaled, and unmarshalled.
const (
	NoEscape EscapeMode = iota
	Base64Escape
	QuotedEscape
)

// DecodeJSONObject will attempt to decode the string input as a JSONValue.
// Optionally decoding base64 the value first before JSON unmarshalling.
//
// Will panic if the escape mode is unknown.
func DecodeJSONObject(v string, escape EscapeMode) (JSONObject, error) {
	var b []byte
	var err error

	switch escape {
	case NoEscape:
		b = []byte(v)
	case Base64Escape:
		b, err = base64.StdEncoding.DecodeString(v)
	case QuotedEscape:
		var u string
		u, err = strconv.Unquote(v)
		b = []byte(u)
	default:
		panic(fmt.Sprintf("DecodeJSONObject called with unknown EscapeMode, %v", escape))
	}

	if err != nil {
		return nil, err
	}

	m := JSONObject{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// EncodeJSONObject marshals the value into a JSON string, and optionally base64
// encodes the string before returning it.
//
// Will panic if the escape mode is unknown.
func EncodeJSONObject(v JSONObject, escape EscapeMode) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	switch escape {
	case NoEscape:
		return string(b), nil
	case Base64Escape:
		return base64.StdEncoding.EncodeToString(b), nil
	case QuotedEscape:
		return strconv.Quote(string(b)), nil
	}

	panic(fmt.Sprintf("EncodeJSONObject called with unknown EscapeMode, %v", escape))
}
