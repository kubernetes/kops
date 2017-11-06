load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "k8s.io/apiextensions-apiserver",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/apiextensions-apiserver/pkg/cmd/server:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/server:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/util/logs:go_default_library",
    ],
)

go_binary(
    name = "apiextensions-apiserver",
    importpath = "k8s.io/apiextensions-apiserver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
