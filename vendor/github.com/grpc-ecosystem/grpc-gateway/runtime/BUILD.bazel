load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "context.go",
        "convert.go",
        "doc.go",
        "errors.go",
        "handler.go",
        "marshal_json.go",
        "marshal_jsonpb.go",
        "marshaler.go",
        "marshaler_registry.go",
        "mux.go",
        "pattern.go",
        "proto2_convert.go",
        "query.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/runtime",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/jsonpb:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/runtime/internal:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/utilities:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["pattern_test.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/runtime",
    library = ":go_default_library",
    deps = ["//vendor/github.com/grpc-ecosystem/grpc-gateway/utilities:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "context_test.go",
        "errors_test.go",
        "marshal_json_test.go",
        "marshal_jsonpb_test.go",
        "marshaler_registry_test.go",
        "mux_test.go",
        "query_test.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/runtime_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/golang/protobuf/jsonpb:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/wrappers:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/examplepb:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/utilities:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)
