load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/plugin:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/gengateway:go_default_library",
    ],
)

go_binary(
    name = "protoc-gen-grpc-gateway",
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
