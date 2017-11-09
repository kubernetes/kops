load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "app_yaml.pb.go",
        "appengine.pb.go",
        "application.pb.go",
        "deploy.pb.go",
        "instance.pb.go",
        "location.pb.go",
        "operation.pb.go",
        "service.pb.go",
        "version.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/appengine/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/iam/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/longrunning:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/latlng:go_default_library",
        "//vendor/google.golang.org/genproto/protobuf/field_mask:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
