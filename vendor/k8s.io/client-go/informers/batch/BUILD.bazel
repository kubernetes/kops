load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "k8s.io/client-go/informers/batch",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/client-go/informers/batch/v1:go_default_library",
        "//vendor/k8s.io/client-go/informers/batch/v1beta1:go_default_library",
        "//vendor/k8s.io/client-go/informers/batch/v2alpha1:go_default_library",
        "//vendor/k8s.io/client-go/informers/internalinterfaces:go_default_library",
    ],
)
