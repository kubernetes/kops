load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "check_error.pb.go",
        "distribution.pb.go",
        "log_entry.pb.go",
        "metric_value.pb.go",
        "operation.pb.go",
        "service_controller.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/api/servicecontrol/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/type:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/money:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
