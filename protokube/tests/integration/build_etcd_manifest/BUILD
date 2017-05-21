load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_test")

go_binary(
    name = "build_etcd_manifest",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["integration_test.go"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/diff:go_default_library",
        "//protokube/pkg/protokube:go_default_library",
        "@io_k8s_client_go//pkg/api/install:go_default_library",
    ],
)
