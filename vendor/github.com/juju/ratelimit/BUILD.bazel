load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "ratelimit.go",
        "reader.go",
    ],
    importpath = "github.com/juju/ratelimit",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["ratelimit_test.go"],
    importpath = "github.com/juju/ratelimit",
    library = ":go_default_library",
    deps = ["//vendor/gopkg.in/check.v1:go_default_library"],
)
