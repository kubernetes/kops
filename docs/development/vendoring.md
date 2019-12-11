# Vendoring Go dependencies

kops uses [dep](https://github.com/golang/dep) to manage vendored
dependencies in most versions leading up to kops 1.14.
kops uses [go mod](https://github.com/golang/go/wiki/Modules) to manage
vendored dependencies in versions 1.15 and newer.

## Prerequisites

The following software must be installed prior to running the
update commands:

* [bazel](https://github.com/bazelbuild/bazel)
* [dep](https://github.com/golang/dep) for kops 1.14 or older
* [go mod](https://github.com/golang/go/wiki/Modules) for kops 1.15 and newer branches (including master)
* [hg](https://www.mercurial-scm.org/wiki/Download)

## Adding a dependency to the vendor directory

Go modules will manage required dependencies based on the imports
found in the source code. Follow these steps to run the update process:

1. Add the desired import to a `.go` file.
2. Run `make gomod` to start the update process. If this step is
successful, the imported dependency will be added to the `vendor`
subdirectory.
3. Commit any changes, including changes to the `vendor` directory,
`go.mod`, and `go.sum`.
4. Open a pull request with these changes separately from other work
so that it is easier to review.

## Updating a dependency in the vendor directory (e.g. aws-sdk-go)

1. Update the locked version as specified in `go.mod`
2. Run `make gomod`.
3. Review the changes to ensure that they are as intended / trustworthy.
4. Commit any changes, including changes to the `vendor` directory,
`go.mod` and `go.sum`.
5. Open a pull request with these changes separately from other work so that it
is easier to review.  Please include any significant changes you observed.
