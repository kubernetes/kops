/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// LazyQuoteStringSliceVar is like flagset.StringSliceVar, but supports quotes in values.
func LazyQuoteStringSliceVar(f *pflag.FlagSet, p *[]string, name string, value []string, usage string) {
	f.VarP(newLazyQuoteStringSliceValue(value, p), name, "", usage)
}

// lazyQuoteStringSliceValue implements stringSlice, but allows quotes in the values.
// Inspired by https://github.com/spf13/pflag/pull/371
type lazyQuoteStringSliceValue struct {
	value      *[]string
	changed    bool
	lazyQuotes bool
}

func newLazyQuoteStringSliceValue(val []string, p *[]string) *lazyQuoteStringSliceValue {
	ssv := new(lazyQuoteStringSliceValue)
	ssv.value = p
	*ssv.value = val
	ssv.lazyQuotes = true
	return ssv
}

func readAsCSV(val string, lazyQuotes bool) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	csvReader.LazyQuotes = lazyQuotes
	return csvReader.Read()
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (s *lazyQuoteStringSliceValue) Set(val string) error {
	v, err := readAsCSV(val, s.lazyQuotes)
	if err != nil {
		return fmt.Errorf("stringSliceValue %q: %w", val, err)
	}
	if !s.changed {
		*s.value = v
	} else {
		*s.value = append(*s.value, v...)
	}
	s.changed = true
	return nil
}

func (s *lazyQuoteStringSliceValue) Type() string {
	return "stringSlice"
}

func (s *lazyQuoteStringSliceValue) String() string {
	str, _ := writeAsCSV(*s.value)
	return "[" + str + "]"
}

func (s *lazyQuoteStringSliceValue) Append(val string) error {
	*s.value = append(*s.value, val)
	return nil
}

func (s *lazyQuoteStringSliceValue) Replace(val []string) error {
	*s.value = val
	return nil
}

func (s *lazyQuoteStringSliceValue) GetSlice() []string {
	return *s.value
}
