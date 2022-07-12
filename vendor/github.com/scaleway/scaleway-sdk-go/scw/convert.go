package scw

import (
	"net"
	"time"
)

// StringPtr returns a pointer to the string value passed in.
func StringPtr(v string) *string {
	return &v
}

// StringSlicePtr converts a slice of string values into a slice of
// string pointers
func StringSlicePtr(src []string) []*string {
	dst := make([]*string, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// StringsPtr returns a pointer to the []string value passed in.
func StringsPtr(v []string) *[]string {
	return &v
}

// StringsSlicePtr converts a slice of []string values into a slice of
// []string pointers
func StringsSlicePtr(src [][]string) []*[]string {
	dst := make([]*[]string, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// BytesPtr returns a pointer to the []byte value passed in.
func BytesPtr(v []byte) *[]byte {
	return &v
}

// BytesSlicePtr converts a slice of []byte values into a slice of
// []byte pointers
func BytesSlicePtr(src [][]byte) []*[]byte {
	dst := make([]*[]byte, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// BoolPtr returns a pointer to the bool value passed in.
func BoolPtr(v bool) *bool {
	return &v
}

// BoolSlicePtr converts a slice of bool values into a slice of
// bool pointers
func BoolSlicePtr(src []bool) []*bool {
	dst := make([]*bool, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Int32Ptr returns a pointer to the int32 value passed in.
func Int32Ptr(v int32) *int32 {
	return &v
}

// Int32SlicePtr converts a slice of int32 values into a slice of
// int32 pointers
func Int32SlicePtr(src []int32) []*int32 {
	dst := make([]*int32, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Int64Ptr returns a pointer to the int64 value passed in.
func Int64Ptr(v int64) *int64 {
	return &v
}

// Int64SlicePtr converts a slice of int64 values into a slice of
// int64 pointers
func Int64SlicePtr(src []int64) []*int64 {
	dst := make([]*int64, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Uint32Ptr returns a pointer to the uint32 value passed in.
func Uint32Ptr(v uint32) *uint32 {
	return &v
}

// Uint32SlicePtr converts a slice of uint32 values into a slice of
// uint32 pointers
func Uint32SlicePtr(src []uint32) []*uint32 {
	dst := make([]*uint32, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Uint64Ptr returns a pointer to the uint64 value passed in.
func Uint64Ptr(v uint64) *uint64 {
	return &v
}

// Uint64SlicePtr converts a slice of uint64 values into a slice of
// uint64 pointers
func Uint64SlicePtr(src []uint64) []*uint64 {
	dst := make([]*uint64, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Float32Ptr returns a pointer to the float32 value passed in.
func Float32Ptr(v float32) *float32 {
	return &v
}

// Float32SlicePtr converts a slice of float32 values into a slice of
// float32 pointers
func Float32SlicePtr(src []float32) []*float32 {
	dst := make([]*float32, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Float64Ptr returns a pointer to the float64 value passed in.
func Float64Ptr(v float64) *float64 {
	return &v
}

// Float64SlicePtr converts a slice of float64 values into a slice of
// float64 pointers
func Float64SlicePtr(src []float64) []*float64 {
	dst := make([]*float64, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// TimeDurationPtr returns a pointer to the Duration value passed in.
func TimeDurationPtr(v time.Duration) *time.Duration {
	return &v
}

// TimePtr returns a pointer to the Time value passed in.
func TimePtr(v time.Time) *time.Time {
	return &v
}

// SizePtr returns a pointer to the Size value passed in.
func SizePtr(v Size) *Size {
	return &v
}

// IPPtr returns a pointer to the net.IP value passed in.
func IPPtr(v net.IP) *net.IP {
	return &v
}
