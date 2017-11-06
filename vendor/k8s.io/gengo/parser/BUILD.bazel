load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "parse.go",
    ],
    importpath = "k8s.io/gengo/parser",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["local_parse_test.go"],
    importpath = "k8s.io/gengo/parser",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["parse_test.go"],
    importpath = "k8s.io/gengo/parser_test",
    deps = [
        ":go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
    ],
)
