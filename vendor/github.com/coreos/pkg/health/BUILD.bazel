load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["health.go"],
    importpath = "github.com/coreos/pkg/health",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/httputil:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["health_test.go"],
    importpath = "github.com/coreos/pkg/health",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/pkg/httputil:go_default_library"],
)
