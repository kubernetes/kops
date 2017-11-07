load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["push.go"],
    importpath = "github.com/prometheus/client_golang/prometheus/push",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["push_test.go"],
    importpath = "github.com/prometheus/client_golang/prometheus/push",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_add_from_gatherer_test.go",
        "examples_test.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus/push_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
    ],
)
