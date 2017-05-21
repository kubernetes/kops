load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_test")

go_binary(
    name = "conversion",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["integration_test.go"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/install:go_default_library",
        "//pkg/apis/kops/v1alpha1:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "//pkg/diff:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
    ],
)
