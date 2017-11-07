load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "job_service.pb.go",
        "model_service.pb.go",
        "operation_metadata.pb.go",
        "prediction_service.pb.go",
        "project_service.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/cloud/ml/v1beta1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/httpbody:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/serviceconfig:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/longrunning:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
