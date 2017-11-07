load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "handlers.go",
        "stats.go",
    ],
    importpath = "google.golang.org/grpc/stats",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["stats_test.go"],
    importpath = "google.golang.org/grpc/stats_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
        "//vendor/google.golang.org/grpc/stats/grpc_testing:go_default_library",
    ],
)
