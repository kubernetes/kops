load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["util.go"],
    importpath = "k8s.io/apiserver/pkg/authentication/serviceaccount",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/api/validation:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["util_test.go"],
    importpath = "k8s.io/apiserver/pkg/authentication/serviceaccount",
    library = ":go_default_library",
)
