load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["versioning.go"],
    importpath = "k8s.io/apimachinery/pkg/runtime/serializer/versioning",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["versioning_test.go"],
    importpath = "k8s.io/apimachinery/pkg/runtime/serializer/versioning",
    library = ":go_default_library",
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/diff:go_default_library",
    ],
)
