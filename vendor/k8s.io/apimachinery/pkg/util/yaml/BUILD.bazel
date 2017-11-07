load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["decoder.go"],
    importpath = "k8s.io/apimachinery/pkg/util/yaml",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["decoder_test.go"],
    importpath = "k8s.io/apimachinery/pkg/util/yaml",
    library = ":go_default_library",
)
