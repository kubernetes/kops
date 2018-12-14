// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
)

var _ sql.Scanner = &Date{}
var _ driver.Valuer = Date{}

func TestDate(t *testing.T) {
	pp := Date{}
	err := pp.UnmarshalText([]byte{})
	assert.NoError(t, err)
	err = pp.UnmarshalText([]byte("yada"))
	assert.Error(t, err)

	orig := "2014-12-15"
	b := []byte(orig)
	bj := []byte("\"" + orig + "\"")
	err = pp.UnmarshalText([]byte(orig))
	assert.NoError(t, err)

	txt, err := pp.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, orig, string(txt))

	err = pp.UnmarshalJSON(bj)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, pp.String())

	err = pp.UnmarshalJSON([]byte(`"1972/01/01"`))
	assert.Error(t, err)

	b, err = pp.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, bj, b)

	dateOriginal := Date(time.Date(2014, 10, 10, 0, 0, 0, 0, time.UTC))

	bsonData, err := bson.Marshal(&dateOriginal)
	assert.NoError(t, err)

	var dateCopy Date
	err = bson.Unmarshal(bsonData, &dateCopy)
	assert.NoError(t, err)
	assert.Equal(t, dateOriginal, dateCopy)

	var dateZero Date
	err = dateZero.UnmarshalJSON([]byte("null"))
	assert.NoError(t, err)
	assert.Equal(t, Date{}, dateZero)
}

func TestDate_Scan(t *testing.T) {
	ref := time.Now().Truncate(24 * time.Hour).UTC()
	date, str := Date(ref), ref.Format(RFC3339FullDate)

	values := []interface{}{str, []byte(str), ref}
	for _, value := range values {
		result := Date{}
		(&result).Scan(value)
		assert.Equal(t, date, result, "value: %#v", value)
	}

	dd := Date{}
	err := dd.Scan(nil)
	assert.NoError(t, err)
	assert.Equal(t, Date{}, dd)

	err = dd.Scan(19700101)
	assert.Error(t, err)
}

func TestDate_Value(t *testing.T) {
	ref := time.Now().Truncate(24 * time.Hour).UTC()
	date := Date(ref)
	dbv, err := date.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, dbv, ref.Format("2006-01-02"))
}

func TestDate_IsDate(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{"2017-12-22", true},
		{"2017-1-1", false},
		{"17-13-22", false},
		{"2017-02-29", false}, // not a valid date : 2017 is not a leap year
		{"1900-02-29", false}, // not a valid date : 1900 is not a leap year
		{"2100-02-29", false}, // not a valid date : 2100 is not a leap year
		{"2000-02-29", true},  // a valid date : 2000 is a leap year
		{"2400-02-29", true},  // a valid date : 2000 is a leap year
		{"2017-13-22", false},
		{"2017-12-32", false},
		{"20171-12-32", false},
		{"YYYY-MM-DD", false},
		{"20-17-2017", false},
		{"2017-12-22T01:02:03Z", false},
	}
	for _, test := range tests {
		assert.Equal(t, test.valid, IsDate(test.value), "value [%s] should be valid: [%t]", test.value, test.valid)
	}
}
