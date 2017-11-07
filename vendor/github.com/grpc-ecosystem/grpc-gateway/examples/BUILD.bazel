load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/examplepb:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/runtime:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_binary(
    name = "examples",
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "integration_test.go",
        "main_test.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/jsonpb:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/clients/abe:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/clients/echo:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/examplepb:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/server:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/sub:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/runtime:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)
