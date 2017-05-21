load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "node_api_adapter.go",
        "validate_cluster.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/util:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/util/wait:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "node_api_adapter_test.go",
        "validate_cluster_test.go",
    ],
    library = ":go_default_library",
    deps = [
        "@io_k8s_apimachinery//pkg/api/resource:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//kubernetes/fake:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
    ],
)
