# Running an apiserver and controller-manager locally

This document describes how to run an apiserver and controller-manager
locally.

## Build the executables

`apiserver-boot build executables`

This will build the apiserver and controller-manager using
go build and put the binaries under `bin/`.  The commands
used to build the binaries are printed to the terminal.

## Run the executables and etcd

`apiserver-boot run local`

This will start the apiserver and controller-manager binaries
under `bin/` and send the output of both to the terminal.
The commands used to start the binaries are printed
to the terminal.

**Note:** The location of the binaries can be controlled with `--apiserver` and `--controller-manager`.