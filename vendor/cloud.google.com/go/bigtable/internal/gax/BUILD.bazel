load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "call_option.go",
        "invoke.go",
    ],
    importpath = "cloud.google.com/go/bigtable/internal/gax",
    visibility = ["//vendor/cloud.google.com/go/bigtable:__subpackages__"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["invoke_test.go"],
    importpath = "cloud.google.com/go/bigtable/internal/gax",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)
