// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package bytequantity

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	/// Examples:          1mb, 1 gb, 1.0tb, 1mib, 2g, 2.001 t
	byteQuantityRegex = `^([0-9]+\.?[0-9]{0,3})[ ]?(mi?b?|gi?b?|ti?b?)?$`
	mib               = "MiB"
	gib               = "GiB"
	tib               = "TiB"
	gbConvert         = 1 << 10
	tbConvert         = gbConvert << 10
	maxGiB            = math.MaxUint64 / gbConvert
	maxTiB            = math.MaxUint64 / tbConvert
)

// ByteQuantity is a data type representing a byte quantity
type ByteQuantity struct {
	Quantity uint64
}

// ParseToByteQuantity parses a string representation of a byte quantity to a ByteQuantity type.
// A unit can be appended such as 16 GiB. If no unit is appended, GiB is assumed.
func ParseToByteQuantity(byteQuantityStr string) (ByteQuantity, error) {
	bqRegexp := regexp.MustCompile(byteQuantityRegex)
	matches := bqRegexp.FindStringSubmatch(strings.ToLower(byteQuantityStr))
	if len(matches) < 2 {
		return ByteQuantity{}, fmt.Errorf("%s is not a valid byte quantity", byteQuantityStr)
	}

	quantityStr := matches[1]
	unit := gib
	if len(matches) > 2 && matches[2] != "" {
		unit = matches[2]
	}
	quantity := uint64(0)
	switch strings.ToLower(string(unit[0])) {
	//mib
	case "m":
		inputDecSplit := strings.Split(quantityStr, ".")
		if len(inputDecSplit) == 2 {
			d, err := strconv.Atoi(inputDecSplit[1])
			if err != nil {
				return ByteQuantity{}, err
			}
			if d != 0 {
				return ByteQuantity{}, fmt.Errorf("cannot accept floating point MB value, only integers are accepted")
			}
		}
		// need error here so that this quantity doesn't bind in the local scope
		var err error
		quantity, err = strconv.ParseUint(inputDecSplit[0], 10, 64)
		if err != nil {
			return ByteQuantity{}, err
		}
	//gib
	case "g":
		quantityDec, err := strconv.ParseFloat(quantityStr, 10)
		if err != nil {
			return ByteQuantity{}, err
		}
		if quantityDec > maxGiB {
			return ByteQuantity{}, fmt.Errorf("error GiB value is too large")
		}
		quantity = uint64(quantityDec * gbConvert)
	//tib
	case "t":
		quantityDec, err := strconv.ParseFloat(quantityStr, 10)
		if err != nil {
			return ByteQuantity{}, err
		}
		if quantityDec > maxTiB {
			return ByteQuantity{}, fmt.Errorf("error TiB value is too large")
		}
		quantity = uint64(quantityDec * tbConvert)
	default:
		return ByteQuantity{}, fmt.Errorf("error unit %s is not supported", unit)
	}

	return ByteQuantity{
		Quantity: quantity,
	}, nil
}

// FromTiB returns a byte quantity of the passed in tebibytes quantity
func FromTiB(tib uint64) ByteQuantity {
	return ByteQuantity{
		Quantity: tib * tbConvert,
	}
}

// FromGiB returns a byte quantity of the passed in gibibytes quantity
func FromGiB(gib uint64) ByteQuantity {
	return ByteQuantity{
		Quantity: gib * gbConvert,
	}
}

// FromMiB returns a byte quantity of the passed in mebibytes quantity
func FromMiB(mib uint64) ByteQuantity {
	return ByteQuantity{
		Quantity: mib,
	}
}

// StringMiB returns a byte quantity in a mebibytes string representation
func (bq ByteQuantity) StringMiB() string {
	return fmt.Sprintf("%.0f %s", bq.MiB(), mib)
}

// StringGiB returns a byte quantity in a gibibytes string representation
func (bq ByteQuantity) StringGiB() string {
	return fmt.Sprintf("%.3f %s", bq.GiB(), gib)
}

// StringTiB returns a byte quantity in a tebibytes string representation
func (bq ByteQuantity) StringTiB() string {
	return fmt.Sprintf("%.3f %s", bq.TiB(), tib)
}

// MiB returns a byte quantity in mebibytes
func (bq ByteQuantity) MiB() float64 {
	return float64(bq.Quantity)
}

// GiB returns a byte quantity in gibibytes
func (bq ByteQuantity) GiB() float64 {
	return float64(bq.Quantity) * 1 / gbConvert
}

// TiB returns a byte quantity in tebibytes
func (bq ByteQuantity) TiB() float64 {
	return float64(bq.Quantity) * 1 / tbConvert
}
