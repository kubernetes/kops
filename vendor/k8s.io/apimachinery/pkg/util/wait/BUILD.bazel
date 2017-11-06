load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "wait.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/util/wait",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["wait_test.go"],
    importpath = "k8s.io/apimachinery/pkg/util/wait",
    library = ":go_default_library",
    deps = ["//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library"],
)
