load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "k8s.io/client-go/informers/storage",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/client-go/informers/internalinterfaces:go_default_library",
        "//vendor/k8s.io/client-go/informers/storage/v1:go_default_library",
        "//vendor/k8s.io/client-go/informers/storage/v1alpha1:go_default_library",
        "//vendor/k8s.io/client-go/informers/storage/v1beta1:go_default_library",
    ],
)
