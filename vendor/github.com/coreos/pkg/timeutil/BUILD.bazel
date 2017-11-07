load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["backoff.go"],
    importpath = "github.com/coreos/pkg/timeutil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["backoff_test.go"],
    importpath = "github.com/coreos/pkg/timeutil",
    library = ":go_default_library",
)
