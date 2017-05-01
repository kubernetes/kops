# Issue 267
### Configuration file instead of command line switches
[Link to issue](https://github.com/kubernetes/kops/issues/267)


# Overloading flags 

The purpose of the new YAML configuration file is map command line flags to an optional YAML configuration file.

There will be a 1 to 1 relationship between kops flags and directives in the configuration file.

The configuration file will be unmarshalled onto the Cobra command struct for each supported command. (`create cluster`, `get cluster`, `delete cluster`, `update cluster`)

The unmarshalling will be dynamic, and lazy (only unmarshal if the directive in the YAML configuration file matches the struct member verbatim).

Each command that supports the new optional overload will have a unique implementation in the `Run()` function.



Ex:

    Cloud: aws # Will only be unmarshalled if the command struct as a "Cloud" member.
   
# Viper

The configuration parsing will use the viper dependency already present in kops. It will create a new viper parser, with a unique name as to not conflict with existing Viper implementation.



