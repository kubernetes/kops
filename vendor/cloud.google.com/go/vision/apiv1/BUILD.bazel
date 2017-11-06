load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "image_annotator_client.go",
        "vision.go",
    ],
    importpath = "cloud.google.com/go/vision/apiv1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/cloud/vision/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["image_annotator_client_example_test.go"],
    importpath = "cloud.google.com/go/vision/apiv1_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/cloud/vision/v1:go_default_library",
    ],
)
