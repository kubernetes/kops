load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "keys.pb.go",
        "mutation.pb.go",
        "query_plan.pb.go",
        "result_set.pb.go",
        "spanner.pb.go",
        "transaction.pb.go",
        "type.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/spanner/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/serviceconfig:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
