/*
Copyright 2016 The Kubernetes Authors.

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

package test

import (
	"fmt"

	. "github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/types"
)

// Print a newline after the default Reporter due to issue
// https://github.com/jstemmer/go-junit-report/issues/31
type NewlineReporter struct{}

func (NewlineReporter) SpecSuiteWillBegin(config GinkgoConfigType, summary *SuiteSummary) {}

func (NewlineReporter) BeforeSuiteDidRun(setupSummary *SetupSummary) {}

func (NewlineReporter) AfterSuiteDidRun(setupSummary *SetupSummary) {}

func (NewlineReporter) SpecWillRun(specSummary *SpecSummary) {}

func (NewlineReporter) SpecDidComplete(specSummary *SpecSummary) {}

// SpecSuiteDidEnd Prints a newline between "35 Passed | 0 Failed | 0 Pending | 0 Skipped" and "--- PASS:"
func (NewlineReporter) SpecSuiteDidEnd(summary *SuiteSummary) { fmt.Printf("\n") }
