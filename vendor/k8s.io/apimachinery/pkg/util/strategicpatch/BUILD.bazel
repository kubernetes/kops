load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["patch.go"],
    importpath = "k8s.io/apimachinery/pkg/util/strategicpatch",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/util/json:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/mergepatch:go_default_library",
        "//vendor/k8s.io/apimachinery/third_party/forked/golang/json:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["patch_test.go"],
    importpath = "k8s.io/apimachinery/pkg/util/strategicpatch",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/davecgh/go-spew/spew:go_default_library",
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/mergepatch:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
    ],
)
