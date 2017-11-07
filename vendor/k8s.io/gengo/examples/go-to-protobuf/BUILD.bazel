load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "k8s.io/gengo/examples/go-to-protobuf",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/gengo/examples/go-to-protobuf/protobuf:go_default_library",
    ],
)

go_binary(
    name = "go-to-protobuf",
    importpath = "k8s.io/gengo/examples/go-to-protobuf",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
