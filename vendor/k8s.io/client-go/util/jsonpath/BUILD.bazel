load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "jsonpath.go",
        "node.go",
        "parser.go",
    ],
    importpath = "k8s.io/client-go/util/jsonpath",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/client-go/third_party/forked/golang/template:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "jsonpath_test.go",
        "parser_test.go",
    ],
    importpath = "k8s.io/client-go/util/jsonpath",
    library = ":go_default_library",
)
