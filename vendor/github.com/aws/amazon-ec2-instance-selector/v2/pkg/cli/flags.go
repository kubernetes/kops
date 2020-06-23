package cli

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/spf13/pflag"
)

const (
	maxInt    = int(^uint(0) >> 1)
	maxUint64 = math.MaxUint64
)

// RatioFlag creates and registers a flag accepting a Ratio
func (cl *CommandLineInterface) RatioFlag(name string, shorthand *string, defaultValue *string, description string) error {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.StringMe("")
	}
	if shorthand != nil {
		cl.Flags[name] = cl.Command.Flags().StringP(name, string(*shorthand), *defaultValue, description)
	} else {
		cl.Flags[name] = cl.Command.Flags().String(name, *defaultValue, description)
	}

	cl.validators[name] = func(val interface{}) error {
		if val == nil {
			return nil
		}
		vcpuToMemRatioVal := *val.(*string)
		valid, err := regexp.Match(`^[0-9]+:[0-9]+$`, []byte(vcpuToMemRatioVal))
		if err != nil || !valid {
			return fmt.Errorf("Invalid input for --%s. A valid example is 1:2", name)
		}
		vals := strings.Split(vcpuToMemRatioVal, ":")
		vcpusRatioVal, err1 := strconv.Atoi(vals[0])
		memRatioVal, err2 := strconv.Atoi(vals[1])
		if err1 != nil || err2 != nil {
			return fmt.Errorf("Invalid input for --%s. Ratio values must be integers. A valid example is 1:2", name)
		}
		cl.Flags[name] = cl.Float64Me(float64(memRatioVal) / float64(vcpusRatioVal))
		return nil
	}
	return nil
}

// IntMinMaxRangeFlags creates and registers a min, max, and helper flag each accepting an Integer
func (cl *CommandLineInterface) IntMinMaxRangeFlags(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntMinMaxRangeFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// ByteQuantityMinMaxRangeFlags creates and registers a min, max, and helper flag each accepting a byte quantity like 512mb
func (cl *CommandLineInterface) ByteQuantityMinMaxRangeFlags(name string, shorthand *string, defaultValue *bytequantity.ByteQuantity, description string) {
	cl.ByteQuantityMinMaxRangeFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// ByteQuantityFlag creates and registers a flag accepting a byte quantity like 512mb
func (cl *CommandLineInterface) ByteQuantityFlag(name string, shorthand *string, defaultValue *bytequantity.ByteQuantity, description string) {
	cl.ByteQuantityFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// IntFlag creates and registers a flag accepting an Integer
func (cl *CommandLineInterface) IntFlag(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// StringFlag creates and registers a flag accepting a String and a validator function.
// The validator function is provided so that more complex flags can be created from a string input.
func (cl *CommandLineInterface) StringFlag(name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	cl.StringFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description, nil, validationFn)
}

// StringSliceFlag creates and registers a flag accepting a list of strings.
func (cl *CommandLineInterface) StringSliceFlag(name string, shorthand *string, defaultValue []string, description string) {
	cl.StringSliceFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// RegexFlag creates and registers a flag accepting a string and validates that it is a valid regex.
func (cl *CommandLineInterface) RegexFlag(name string, shorthand *string, defaultValue *string, description string) {
	cl.RegexFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// StringOptionsFlag creates and registers a flag accepting a string and valid options for use in validation.
func (cl *CommandLineInterface) StringOptionsFlag(name string, shorthand *string, defaultValue *string, description string, validOpts []string) {
	cl.StringOptionsFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description, validOpts)
}

// BoolFlag creates and registers a flag accepting a boolean
func (cl *CommandLineInterface) BoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.Command.Flags(), name, shorthand, defaultValue, description)
}

// ConfigStringFlag creates and registers a flag accepting a String for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigStringFlag(name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	cl.StringFlagOnFlagSet(cl.Command.PersistentFlags(), name, shorthand, defaultValue, description, nil, validationFn)
}

// ConfigStringSliceFlag creates and registers a flag accepting a list of strings.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) ConfigStringSliceFlag(name string, shorthand *string, defaultValue []string, description string) {
	cl.StringSliceFlagOnFlagSet(cl.Command.PersistentFlags(), name, shorthand, defaultValue, description)
}

// ConfigIntFlag creates and registers a flag accepting an Integer for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigIntFlag(name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(cl.Command.PersistentFlags(), name, shorthand, defaultValue, description)
}

// ConfigBoolFlag creates and registers a flag accepting a boolean for configuration purposes.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigBoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.Command.PersistentFlags(), name, shorthand, defaultValue, description)
}

// ConfigStringOptionsFlag creates and registers a flag accepting a string and valid options for use in validation.
// Config flags will be grouped at the bottom in the output of --help
func (cl *CommandLineInterface) ConfigStringOptionsFlag(name string, shorthand *string, defaultValue *string, description string, validOpts []string) {
	cl.StringOptionsFlagOnFlagSet(cl.Command.PersistentFlags(), name, shorthand, defaultValue, description, validOpts)
}

// SuiteBoolFlag creates and registers a flag accepting a boolean for aggregate filters.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) SuiteBoolFlag(name string, shorthand *string, defaultValue *bool, description string) {
	cl.BoolFlagOnFlagSet(cl.suiteFlags, name, shorthand, defaultValue, description)
}

// SuiteStringFlag creates and registers a flag accepting a string for aggreagate filters.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) SuiteStringFlag(name string, shorthand *string, defaultValue *string, description string, validationFn validator) {
	cl.StringFlagOnFlagSet(cl.suiteFlags, name, shorthand, defaultValue, description, nil, validationFn)
}

// SuiteStringOptionsFlag creates and registers a flag accepting a string and valid options for use in validation.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) SuiteStringOptionsFlag(name string, shorthand *string, defaultValue *string, description string, validOpts []string) {
	cl.StringOptionsFlagOnFlagSet(cl.suiteFlags, name, shorthand, defaultValue, description, validOpts)
}

// SuiteStringSliceFlag creates and registers a flag accepting a list of strings.
// Suite flags will be grouped in the middle of the output --help
func (cl *CommandLineInterface) SuiteStringSliceFlag(name string, shorthand *string, defaultValue []string, description string) {
	cl.StringSliceFlagOnFlagSet(cl.suiteFlags, name, shorthand, defaultValue, description)
}

// BoolFlagOnFlagSet creates and registers a flag accepting a boolean for configuration purposes.
func (cl *CommandLineInterface) BoolFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *bool, description string) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.BoolMe(false)
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.BoolP(name, string(*shorthand), *defaultValue, description)
		return
	}
	cl.Flags[name] = flagSet.Bool(name, *defaultValue, description)
}

// IntMinMaxRangeFlagOnFlagSet creates and registers a min, max, and helper flag each accepting an Integer
func (cl *CommandLineInterface) IntMinMaxRangeFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *int, description string) {
	cl.IntFlagOnFlagSet(flagSet, name, shorthand, defaultValue, fmt.Sprintf("%s (sets --%s-min and -max to the same value)", description, name))
	cl.IntFlagOnFlagSet(flagSet, name+"-min", nil, nil, fmt.Sprintf("Minimum %s If --%s-max is not specified, the upper bound will be infinity", description, name))
	cl.IntFlagOnFlagSet(flagSet, name+"-max", nil, nil, fmt.Sprintf("Maximum %s If --%s-min is not specified, the lower bound will be 0", description, name))
	cl.validators[name] = func(val interface{}) error {
		if cl.Flags[name+"-min"] == nil || cl.Flags[name+"-max"] == nil {
			return nil
		}
		minArg := name + "-min"
		maxArg := name + "-max"
		minVal := cl.Flags[minArg].(*int)
		maxVal := cl.Flags[maxArg].(*int)
		if *minVal > *maxVal {
			return fmt.Errorf("Invalid input for --%s and --%s. %s must be less than or equal to %s", minArg, maxArg, minArg, maxArg)
		}
		return nil
	}
	cl.rangeFlags[name] = true
}

// ByteQuantityMinMaxRangeFlagOnFlagSet creates and registers a min, max, and helper flag each accepting a ByteQuantity like 5mb or 12gb
func (cl *CommandLineInterface) ByteQuantityMinMaxRangeFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *bytequantity.ByteQuantity, description string) {
	cl.ByteQuantityFlagOnFlagSet(flagSet, name, shorthand, defaultValue, fmt.Sprintf("%s (sets --%s-min and -max to the same value)", description, name))
	cl.ByteQuantityFlagOnFlagSet(flagSet, name+"-min", nil, nil, fmt.Sprintf("Minimum %s If --%s-max is not specified, the upper bound will be infinity", description, name))
	cl.ByteQuantityFlagOnFlagSet(flagSet, name+"-max", nil, nil, fmt.Sprintf("Maximum %s If --%s-min is not specified, the lower bound will be 0", description, name))
	cl.validators[name] = func(val interface{}) error {
		if cl.Flags[name+"-min"] == nil || cl.Flags[name+"-max"] == nil {
			return nil
		}
		minArg := name + "-min"
		maxArg := name + "-max"
		minVal := cl.Flags[name+"-min"].(*bytequantity.ByteQuantity).MiB()
		maxVal := cl.Flags[name+"-max"].(*bytequantity.ByteQuantity).MiB()
		if minVal > maxVal {
			return fmt.Errorf("Invalid input for --%s and --%s. %s must be less than or equal to %s", minArg, maxArg, minArg, maxArg)
		}
		return nil
	}
	cl.rangeFlags[name] = true
}

// ByteQuantityFlagOnFlagSet creates and registers a flag accepting a ByteQuantity
func (cl *CommandLineInterface) ByteQuantityFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *bytequantity.ByteQuantity, description string) {
	invalidInputMsg := fmt.Sprintf("Invalid input for --%s. A valid example is 16gb. ", name)
	byteQuantityProcessor := func(val interface{}) error {
		if val == nil {
			return nil
		}
		switch byteQuantityInput := val.(type) {
		case *string:
			bq, err := bytequantity.ParseToByteQuantity(*byteQuantityInput)
			if err != nil {
				return fmt.Errorf(invalidInputMsg+"Can't parse byte quantity %s.", *byteQuantityInput)
			}
			cl.Flags[name] = &bq
		case *bytequantity.ByteQuantity:
			return nil
		default:
			return fmt.Errorf(invalidInputMsg + "Input type is unsupported.")
		}
		return nil
	}
	byteQuantityValidator := func(val interface{}) error {
		if val == nil {
			return nil
		}
		switch val.(type) {
		case *bytequantity.ByteQuantity:
			return nil
		default:
			return fmt.Errorf(invalidInputMsg + "Processing failed.")
		}
	}
	var stringDefaultValue *string
	if defaultValue != nil {
		stringDefaultValue = cl.StringMe(defaultValue.StringGiB())
	} else {
		stringDefaultValue = nil
	}
	cl.StringFlagOnFlagSet(flagSet, name, shorthand, stringDefaultValue, description, byteQuantityProcessor, byteQuantityValidator)
}

// IntFlagOnFlagSet creates and registers a flag accepting an Integer
func (cl *CommandLineInterface) IntFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *int, description string) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.IntMe(0)
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.IntP(name, string(*shorthand), *defaultValue, description)
		return
	}
	cl.Flags[name] = flagSet.Int(name, *defaultValue, description)
}

// StringFlagOnFlagSet creates and registers a flag accepting a String and a validator function.
// The validator function is provided so that more complex flags can be created from a string input.
func (cl *CommandLineInterface) StringFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *string, description string, processorFn processor, validationFn validator) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = cl.StringMe("")
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.StringP(name, string(*shorthand), *defaultValue, description)
	} else {
		cl.Flags[name] = flagSet.String(name, *defaultValue, description)
	}
	cl.processors[name] = processorFn
	cl.validators[name] = validationFn
}

// StringOptionsFlagOnFlagSet creates and registers a flag accepting a String with valid options.
// The validOpts slice of strings will be used to perform validation
func (cl *CommandLineInterface) StringOptionsFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *string, description string, validOpts []string) {
	validationFn := func(val interface{}) error {
		if val == nil {
			return nil
		}
		for _, v := range validOpts {
			if v == *val.(*string) {
				return nil
			}
		}
		return fmt.Errorf("error %s must be one of: %s", name, strings.Join(validOpts, ", "))
	}
	cl.StringFlagOnFlagSet(flagSet, name, shorthand, defaultValue, description, nil, validationFn)
}

// StringSliceFlagOnFlagSet creates and registers a flag accepting a String Slice.
func (cl *CommandLineInterface) StringSliceFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue []string, description string) {
	if defaultValue == nil {
		cl.nilDefaults[name] = true
		defaultValue = []string{}
	}
	if shorthand != nil {
		cl.Flags[name] = flagSet.StringSliceP(name, string(*shorthand), defaultValue, description)
		return
	}
	cl.Flags[name] = flagSet.StringSlice(name, defaultValue, description)
}

// RegexFlagOnFlagSet creates and registers a flag accepting a string slice of regular expressions.
func (cl *CommandLineInterface) RegexFlagOnFlagSet(flagSet *pflag.FlagSet, name string, shorthand *string, defaultValue *string, description string) {
	invalidInputMsg := fmt.Sprintf("Invalid regex input for --%s. ", name)
	regexProcessor := func(val interface{}) error {
		if val == nil {
			return nil
		}
		switch v := val.(type) {
		case *string:
			regexVal, err := regexp.Compile(*v)
			if err != nil {
				return fmt.Errorf(invalidInputMsg + "Unable to compile the regex.")
			}
			cl.Flags[name] = regexVal
		case *regexp.Regexp:
			return nil
		default:
			return fmt.Errorf(invalidInputMsg + "Input type is unsupported.")
		}

		return nil
	}
	regexValidator := func(val interface{}) error {
		if val == nil {
			return nil
		}
		switch val.(type) {
		case *regexp.Regexp:
			return nil
		default:
			return fmt.Errorf(invalidInputMsg + "Processing failed.")
		}
	}
	cl.StringFlagOnFlagSet(flagSet, name, shorthand, defaultValue, description, regexProcessor, regexValidator)
}
