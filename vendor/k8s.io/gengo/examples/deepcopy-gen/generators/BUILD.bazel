load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["deepcopy.go"],
    importpath = "k8s.io/gengo/examples/deepcopy-gen/generators",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/examples/set-gen/sets:go_default_library",
        "//vendor/k8s.io/gengo/generator:go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["deepcopy_test.go"],
    importpath = "k8s.io/gengo/examples/deepcopy-gen/generators",
    library = ":go_default_library",
    deps = ["//vendor/k8s.io/gengo/types:go_default_library"],
)
