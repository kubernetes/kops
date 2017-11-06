load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "errors.go",
        "path.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/util/validation/field",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "errors_test.go",
        "path_test.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/util/validation/field",
    library = ":go_default_library",
)
