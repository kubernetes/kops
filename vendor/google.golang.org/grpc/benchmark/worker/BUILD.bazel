load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "benchmark_client.go",
        "benchmark_server.go",
        "main.go",
        "util.go",
    ],
    importpath = "google.golang.org/grpc/benchmark/worker",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/benchmark:go_default_library",
        "//vendor/google.golang.org/grpc/benchmark/grpc_testing:go_default_library",
        "//vendor/google.golang.org/grpc/benchmark/stats:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/credentials:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
    ],
)

go_binary(
    name = "worker",
    importpath = "google.golang.org/grpc/benchmark/worker",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
