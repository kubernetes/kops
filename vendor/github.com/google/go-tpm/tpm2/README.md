# TPM 2.0 client library

## Tests

This library contains unit tests in `github.com/google/go-tpm/tpm2`, which just
tests that various encoding and error checking functions work correctly. It also
contains more comprehensive integration tests in
`github.com/google/go-tpm/tpm2/test`, which run actual commands on a TPM.

By default, these integration tests are run against the
[`go-tpm-tools`](https://github.com/google/go-tpm-tools)
simulator, which is baesed on the
[Microsoft Reference TPM2 code](https://github.com/microsoft/ms-tpm-20-ref). To
run both the unit and integration tests, run (in this directory)
```bash
go test . ./test
```

These integration tests can also be run against a real TPM device. This is
slightly more complex as the tests often need to be built as a normal user and
then executed as root. For example,
```bash
# Build the test binary without running it
go test -c github.com/google/go-tpm/tpm2/test
# Execute the test binary as root
sudo ./test.test --tpm-path=/dev/tpmrm0
```
On Linux, The `--tpm-path` causes the integration tests to be run against a
real TPM located at that path (usually `/dev/tpmrm0` or `/dev/tpm0`). On Windows, the story is similar, execept that
the `--use-tbs` flag is used instead.

Tip: if your TPM host is remote and you don't want to install Go on it, this
same two-step process can be used. The test binary can be copied to a remote
host and run without extra installation (as the test binary has very few
*runtime* dependancies).
