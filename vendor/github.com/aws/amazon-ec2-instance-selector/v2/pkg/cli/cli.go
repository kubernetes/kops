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

// Package cli provides functions to build the selector command line interface
package cli

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type runFunc = func(cmd *cobra.Command, args []string)

// New creates an instance of CommandLineInterface
func New(binaryName string, shortUsage string, longUsage, examples string, run runFunc) CommandLineInterface {
	cmd := &cobra.Command{
		Use:     binaryName,
		Short:   shortUsage,
		Long:    longUsage,
		Example: examples,
		Run:     run,
	}
	return CommandLineInterface{
		Command:     cmd,
		Flags:       map[string]interface{}{},
		nilDefaults: map[string]bool{},
		rangeFlags:  map[string]bool{},
		validators:  map[string]validator{},
		processors:  map[string]processor{},
		suiteFlags:  pflag.NewFlagSet("suite", pflag.ExitOnError),
	}
}

// ParseFlags will parse flags registered in this instance of CLI from os.Args
func (cl *CommandLineInterface) ParseFlags() (map[string]interface{}, error) {
	cl.setUsageTemplate()
	// Remove Suite Flags so that args only include Config and Filter Flags
	cl.Command.SetArgs(removeIntersectingArgs(cl.suiteFlags))
	// This parses Config and Filter flags only
	if err := cl.Command.Execute(); err != nil {
		return nil, err
	}

	// Remove Config and Filter flags so that only suite flags are parsed
	if err := cl.suiteFlags.Parse(removeIntersectingArgs(cl.Command.Flags())); err != nil {
		return nil, err
	}

	// Add suite flags to Command flagset so that other processing can occur
	// This has to be done after usage is printed so that the flagsets can be grouped properly when printed
	cl.Command.Flags().AddFlagSet(cl.suiteFlags)
	if err := cl.SetUntouchedFlagValuesToNil(); err != nil {
		return nil, err
	}

	if err := cl.ProcessFlags(); err != nil {
		return nil, err
	}
	return cl.Flags, nil
}

// ParseAndValidateFlags will parse flags registered in this instance of CLI from os.Args
// and then perform validation
func (cl *CommandLineInterface) ParseAndValidateFlags() (map[string]interface{}, error) {
	flags, err := cl.ParseFlags()
	if err != nil {
		return nil, err
	}
	if err := cl.ValidateFlags(); err != nil {
		return nil, err
	}
	return flags, nil
}

// ProcessFlags iterates through any registered processors and executes them
// Processors are executed before validators
func (cl *CommandLineInterface) ProcessFlags() error {
	for flagName, processorFn := range cl.processors {
		if processorFn == nil {
			continue
		}
		if err := processorFn(cl.Flags[flagName]); err != nil {
			return err
		}
	}
	if err := cl.ProcessRangeFilterFlags(); err != nil {
		return err
	}
	return nil
}

// ValidateFlags iterates through any registered validators and executes them
func (cl *CommandLineInterface) ValidateFlags() error {
	for flagName, validationFn := range cl.validators {
		if validationFn == nil {
			continue
		}
		err := validationFn(cl.Flags[flagName])
		if err != nil {
			return err
		}
	}
	return nil
}

func removeIntersectingArgs(flagSet *pflag.FlagSet) []string {
	newArgs := []string{}
	skipNext := false
	for i, arg := range os.Args {
		if skipNext {
			skipNext = false
			continue
		}
		arg = strings.Split(arg, "=")[0]
		longFlag := strings.Replace(arg, "--", "", 1)
		if flagSet.Lookup(longFlag) != nil || shorthandLookup(flagSet, arg) != nil {
			if len(os.Args) > i+1 && os.Args[i+1][0] != '-' {
				skipNext = true
			}
			continue
		}
		newArgs = append(newArgs, os.Args[i])
	}
	return newArgs
}

func shorthandLookup(flagSet *pflag.FlagSet, arg string) *pflag.Flag {
	if len(arg) == 2 && arg[0] == '-' && arg[1] != '-' {
		return flagSet.ShorthandLookup(strings.Replace(arg, "-", "", 1))
	}
	return nil
}

func (cl *CommandLineInterface) setUsageTemplate() {
	transformedUsage := usageTemplate
	suiteFlagCount := 0
	cl.suiteFlags.VisitAll(func(*pflag.Flag) {
		suiteFlagCount++
	})
	if suiteFlagCount > 0 {
		transformedUsage = fmt.Sprintf(transformedUsage, "\n\nSuite Flags:\n"+cl.suiteFlags.FlagUsages()+"\n")
	} else {
		transformedUsage = fmt.Sprintf(transformedUsage, "")
	}
	cl.Command.SetUsageTemplate(transformedUsage)
	cl.suiteFlags.Usage = func() {}
	cl.Command.Flags().Usage = func() {}
}

// SetUntouchedFlagValuesToNil iterates through all flags and sets their value to nil if they were not specifically set by the user
// This allows for a specified value, a negative value (like false or empty string), or an unspecified (nil) entry.
func (cl *CommandLineInterface) SetUntouchedFlagValuesToNil() error {
	defaultHandlerErrMsg := "Unable to find a default value handler for %v, marking as no default value. This could be an error"
	defaultHandlerFlags := []string{}

	cl.Command.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			// If nilDefaults entry for flag is set to false, do not change default
			if val, _ := cl.nilDefaults[f.Name]; !val {
				return
			}
			switch v := cl.Flags[f.Name].(type) {
			case *int:
				if reflect.ValueOf(*v).IsZero() {
					cl.Flags[f.Name] = nil
				}
			case *bytequantity.ByteQuantity:
				if v.Quantity == 0 {
					cl.Flags[f.Name] = nil
				}
			case *string:
				if reflect.ValueOf(*v).IsZero() {
					cl.Flags[f.Name] = nil
				}
			case *bool:
				if reflect.ValueOf(*v).IsZero() {
					cl.Flags[f.Name] = nil
				}
			case *[]string:
				if reflect.ValueOf(v).IsZero() {
					cl.Flags[f.Name] = nil
				}
			default:
				defaultHandlerFlags = append(defaultHandlerFlags, f.Name)
				cl.Flags[f.Name] = nil
			}
		}
	})
	if len(defaultHandlerFlags) != 0 {
		return fmt.Errorf(defaultHandlerErrMsg, defaultHandlerFlags)
	}
	return nil
}

// ProcessRangeFilterFlags sets min and max to the appropriate 0 or max bounds based on the 3-tuple that a user specifies for base flag, min, and/or max
func (cl *CommandLineInterface) ProcessRangeFilterFlags() error {
	for flagName := range cl.rangeFlags {
		rangeHelperMin := fmt.Sprintf("%s-%s", flagName, "min")
		rangeHelperMax := fmt.Sprintf("%s-%s", flagName, "max")
		if cl.Flags[flagName] != nil {
			if cl.Flags[rangeHelperMin] != nil || cl.Flags[rangeHelperMax] != nil {
				return fmt.Errorf("error: --%s and --%s cannot be set when using --%s", rangeHelperMin, rangeHelperMax, flagName)
			}
			cl.Flags[rangeHelperMin] = cl.Flags[flagName]
			cl.Flags[rangeHelperMax] = cl.Flags[flagName]
		}
		if cl.Flags[rangeHelperMin] == nil && cl.Flags[rangeHelperMax] == nil {
			continue
		}

		if cl.Flags[rangeHelperMin] == nil {
			switch cl.Flags[rangeHelperMax].(type) {
			case *int:
				cl.Flags[rangeHelperMin] = cl.IntMe(0)
			case *bytequantity.ByteQuantity:
				cl.Flags[rangeHelperMin] = cl.ByteQuantityMe(bytequantity.ByteQuantity{Quantity: 0})
			default:
				return fmt.Errorf("Unable to set %s", rangeHelperMax)
			}
		} else if cl.Flags[rangeHelperMax] == nil {
			switch cl.Flags[rangeHelperMin].(type) {
			case *int:
				cl.Flags[rangeHelperMax] = cl.IntMe(maxInt)
			case *bytequantity.ByteQuantity:
				cl.Flags[rangeHelperMax] = cl.ByteQuantityMe(bytequantity.ByteQuantity{Quantity: maxUint64})
			default:
				return fmt.Errorf("Unable to set %s", rangeHelperMin)
			}
		}

		switch cl.Flags[rangeHelperMin].(type) {
		case *int:
			cl.Flags[flagName] = &selector.IntRangeFilter{
				LowerBound: *cl.IntMe(cl.Flags[rangeHelperMin]),
				UpperBound: *cl.IntMe(cl.Flags[rangeHelperMax]),
			}
		case *bytequantity.ByteQuantity:
			cl.Flags[flagName] = &selector.ByteQuantityRangeFilter{
				LowerBound: *cl.ByteQuantityMe(cl.Flags[rangeHelperMin]),
				UpperBound: *cl.ByteQuantityMe(cl.Flags[rangeHelperMax]),
			}
		}
	}
	return nil
}
