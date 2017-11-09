load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "client.go",
        "client_reporter.go",
        "server.go",
        "server_reporter.go",
        "util.go",
    ],
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "server_test.go",
    ],
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/grpc-ecosystem/go-grpc-prometheus/examples/testproto:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
        "//vendor/github.com/stretchr/testify/require:go_default_library",
        "//vendor/github.com/stretchr/testify/suite:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)
