load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "a_bit_of_everything.go",
        "echo.go",
        "flow_combination.go",
        "main.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/server",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/examplepb:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/sub:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/sub2:go_default_library",
        "//vendor/github.com/rogpeppe/fastuuid:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)
