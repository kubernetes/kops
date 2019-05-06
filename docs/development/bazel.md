# Bazel builds

## Overview

Building with bazel offers a deterministic, faster build, including creating smaller docker images.

While bazel works well for small projects, building with kubernetes still has a few challenges.  We take the following approach:

* We don't yet generate files in bazel - we use external scripts (for now)
* We vendor our dependencies, rather than relying on gazelle (although actually gazelle works, the issue is when external dependencies like apimachinery include bazel files that confuse gazelle)
* We strip bazel files from external dependencies, so we don't confuse gazelle

## How to run

```
make bazel-build
make bazel-build-cli

# To build end to end and upload to s3
make bazel-upload
```

Test:
```
make bazel-test
```

Regenerate bazel files using gazelle:
```
make gazelle
```

## Other changes needed

* By default the `go_test` command doesn't allow tests to use data.  So we need to use `data = glob(["testdata/**"]),` or similar. We add `# keep` to stop gazelle from removing it.  `data` doesn't make it easy to access files in a parent directory, so we'll have to clean up some of the test / package structure.
