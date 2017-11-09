load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "default_generator.go",
        "default_package.go",
        "doc.go",
        "error_tracker.go",
        "execute.go",
        "generator.go",
        "import_tracker.go",
        "snippet_writer.go",
    ],
    importpath = "k8s.io/gengo/generator",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/parser:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["snippet_writer_test.go"],
    importpath = "k8s.io/gengo/generator_test",
    deps = [
        ":go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/parser:go_default_library",
    ],
)
