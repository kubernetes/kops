load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cmd.go",
        "generator.go",
        "import_tracker.go",
        "namer.go",
        "package.go",
        "parser.go",
        "tags.go",
    ],
    importpath = "k8s.io/gengo/examples/go-to-protobuf/protobuf",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/generator:go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/parser:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
        "//vendor/k8s.io/kubernetes/third_party/forked/golang/reflect:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["namer_test.go"],
    importpath = "k8s.io/gengo/examples/go-to-protobuf/protobuf",
    library = ":go_default_library",
)
