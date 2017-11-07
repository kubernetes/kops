load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "fuzzer.go",
        "valuefuzz.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/api/testing/fuzzer",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/google/gofuzz:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/serializer:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["valuefuzz_test.go"],
    importpath = "k8s.io/apimachinery/pkg/api/testing/fuzzer",
    library = ":go_default_library",
)
