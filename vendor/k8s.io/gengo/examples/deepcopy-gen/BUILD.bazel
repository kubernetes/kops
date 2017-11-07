load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "k8s.io/gengo/examples/deepcopy-gen",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/examples/deepcopy-gen/generators:go_default_library",
    ],
)

go_binary(
    name = "deepcopy-gen",
    importpath = "k8s.io/gengo/examples/deepcopy-gen",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
