load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["generate.go"],
    importpath = "k8s.io/apiserver/pkg/storage/names",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/util/rand:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["generate_test.go"],
    importpath = "k8s.io/apiserver/pkg/storage/names",
    library = ":go_default_library",
)
