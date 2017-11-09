load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["api.go"],
    importpath = "github.com/prometheus/client_golang/api/prometheus/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/api:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["api_test.go"],
    importpath = "github.com/prometheus/client_golang/api/prometheus/v1",
    library = ":go_default_library",
    deps = ["//vendor/github.com/prometheus/common/model:go_default_library"],
)
