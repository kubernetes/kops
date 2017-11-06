load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "json.go",
        "meta.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/runtime/serializer/json",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/github.com/json-iterator/go:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/serializer/recognizer:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/framer:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/yaml:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["meta_test.go"],
    importpath = "k8s.io/apimachinery/pkg/runtime/serializer/json",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["json_test.go"],
    importpath = "k8s.io/apimachinery/pkg/runtime/serializer/json_test",
    deps = [
        ":go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/diff:go_default_library",
    ],
)
