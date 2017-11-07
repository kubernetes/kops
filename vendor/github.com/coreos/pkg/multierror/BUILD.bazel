load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["multierror.go"],
    importpath = "github.com/coreos/pkg/multierror",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["multierror_test.go"],
    importpath = "github.com/coreos/pkg/multierror",
    library = ":go_default_library",
)
