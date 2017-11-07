load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["server.go"],
    importpath = "google.golang.org/grpc/examples/route_guide/server",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/credentials:go_default_library",
        "//vendor/google.golang.org/grpc/examples/route_guide/routeguide:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
    ],
)

go_binary(
    name = "server",
    importpath = "google.golang.org/grpc/examples/route_guide/server",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
