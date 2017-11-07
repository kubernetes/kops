load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "k8s.io/gengo/examples/set-gen",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/examples/set-gen/generators:go_default_library",
    ],
)

go_binary(
    name = "set-gen",
    importpath = "k8s.io/gengo/examples/set-gen",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
