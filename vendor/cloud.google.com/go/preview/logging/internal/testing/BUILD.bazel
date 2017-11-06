load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["fake.go"],
    importpath = "cloud.google.com/go/preview/logging/internal/testing",
    visibility = ["//vendor/cloud.google.com/go/preview/logging:__subpackages__"],
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/label:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/monitoredres:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["fake_test.go"],
    importpath = "cloud.google.com/go/preview/logging/internal/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/v2:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
