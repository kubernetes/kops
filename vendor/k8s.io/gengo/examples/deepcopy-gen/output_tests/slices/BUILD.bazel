load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "zz_generated.go",
    ],
    importpath = "k8s.io/gengo/examples/deepcopy-gen/output_tests/slices",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/conversion:go_default_library"],
)
