load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["bridge.go"],
    importpath = "github.com/prometheus/client_golang/prometheus/graphite",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["bridge_test.go"],
    importpath = "github.com/prometheus/client_golang/prometheus/graphite",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
