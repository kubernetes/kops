load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fields.go",
        "requirements.go",
        "selector.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/fields",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/selection:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fields_test.go",
        "selector_test.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/fields",
    library = ":go_default_library",
)
