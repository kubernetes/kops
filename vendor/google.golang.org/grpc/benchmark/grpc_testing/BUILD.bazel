load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "control.pb.go",
        "messages.pb.go",
        "payloads.pb.go",
        "services.pb.go",
        "stats.pb.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = [
        "control.proto",
        "messages.proto",
        "payloads.proto",
        "services.proto",
        "stats.proto",
    ],
    visibility = ["//visibility:public"],
)
