package extra

import (
	"testing"

	"github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func init() {
	RegisterFuzzyDecoders()
}

func Test_any_to_string(t *testing.T) {
	should := require.New(t)
	var val string
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal("100", val)
	should.Nil(jsoniter.UnmarshalFromString("10", &val))
	should.Equal("10", val)
	should.Nil(jsoniter.UnmarshalFromString("10.1", &val))
	should.Equal("10.1", val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal("10.1", val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
}
func Test_any_to_int64(t *testing.T) {
	should := require.New(t)
	var val int64

	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(int64(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(int64(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(int64(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(int64(10), val)

	should.Nil(jsoniter.UnmarshalFromString(`-10`, &val))
	should.Equal(int64(-10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_int(t *testing.T) {
	should := require.New(t)
	var val int
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(100, val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(10, val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(10, val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(10, val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_int16(t *testing.T) {
	should := require.New(t)
	var val int16
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(int16(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(int16(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(int16(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(int16(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_int32(t *testing.T) {
	should := require.New(t)
	var val int32
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(int32(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(int32(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(int32(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(int32(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_int8(t *testing.T) {
	should := require.New(t)
	var val int8
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(int8(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(int8(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(int8(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(int8(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_uint8(t *testing.T) {
	should := require.New(t)
	var val uint8
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(uint8(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(uint8(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(uint8(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(uint8(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_uint64(t *testing.T) {
	should := require.New(t)
	var val uint64

	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(uint64(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(uint64(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(uint64(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(uint64(10), val)

	// TODO fix?
	should.NotNil(jsoniter.UnmarshalFromString(`-10`, &val))
	should.Equal(uint64(0), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}
func Test_any_to_uint32(t *testing.T) {
	should := require.New(t)
	var val uint32

	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(uint32(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(uint32(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(uint32(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(uint32(10), val)

	// TODO fix?
	should.NotNil(jsoniter.UnmarshalFromString(`-10`, &val))
	should.Equal(uint32(0), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}
func Test_any_to_uint16(t *testing.T) {
	should := require.New(t)
	var val uint16

	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(uint16(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(uint16(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(uint16(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(uint16(10), val)

	// TODO fix?
	should.NotNil(jsoniter.UnmarshalFromString(`-10`, &val))
	should.Equal(uint16(0), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}
func Test_any_to_uint(t *testing.T) {
	should := require.New(t)
	var val uint
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(uint(100), val)
	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(uint(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(uint(10), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(uint(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
	// large float to int
	should.NotNil(jsoniter.UnmarshalFromString(`1234512345123451234512345.0`, &val))
}

func Test_any_to_float32(t *testing.T) {
	should := require.New(t)
	var val float32
	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(float32(100), val)

	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(float32(10.1), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(float32(10.1), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(float32(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
}

func Test_any_to_float64(t *testing.T) {
	should := require.New(t)
	var val float64

	should.Nil(jsoniter.UnmarshalFromString(`"100"`, &val))
	should.Equal(float64(100), val)

	should.Nil(jsoniter.UnmarshalFromString(`"10.1"`, &val))
	should.Equal(float64(10.1), val)
	should.Nil(jsoniter.UnmarshalFromString(`10.1`, &val))
	should.Equal(float64(10.1), val)
	should.Nil(jsoniter.UnmarshalFromString(`10`, &val))
	should.Equal(float64(10), val)
	should.NotNil(jsoniter.UnmarshalFromString("{}", &val))
	should.NotNil(jsoniter.UnmarshalFromString("[]", &val))
}

func Test_empty_array_as_map(t *testing.T) {
	should := require.New(t)
	var val map[string]interface{}
	should.Nil(jsoniter.UnmarshalFromString(`[]`, &val))
	should.Equal(map[string]interface{}{}, val)
}

func Test_empty_array_as_object(t *testing.T) {
	should := require.New(t)
	var val struct{}
	should.Nil(jsoniter.UnmarshalFromString(`[]`, &val))
	should.Equal(struct{}{}, val)
}
