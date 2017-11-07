load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "delegator_1_7.go",
        "delegator_1_8.go",
        "http.go",
        "instrument_client.go",
        "instrument_client_1_8.go",
        "instrument_server.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus/promhttp",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "http_test.go",
        "instrument_client_1_8_test.go",
        "instrument_server_test.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus/promhttp",
    library = ":go_default_library",
    deps = ["//vendor/github.com/prometheus/client_golang/prometheus:go_default_library"],
)
