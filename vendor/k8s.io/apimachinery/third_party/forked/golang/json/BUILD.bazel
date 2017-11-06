load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["fields.go"],
    importpath = "k8s.io/apimachinery/third_party/forked/golang/json",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["fields_test.go"],
    importpath = "k8s.io/apimachinery/third_party/forked/golang/json",
    library = ":go_default_library",
)
