load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "apply_federation.go",
        "auth_file.go",
        "federation_cluster.go",
        "federation_configuration.go",
        "federation_namespace.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//federation/model:go_default_library",
        "//federation/targets/kubernetestarget:go_default_library",
        "//federation/tasks:go_default_library",
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/registry:go_default_library",
        "//pkg/client/simple:go_default_library",
        "//pkg/kubeconfig:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/fitasks:go_default_library",
        "//upup/pkg/fi/k8sapi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/api/errors:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
        "@io_k8s_kubernetes//federation/apis/federation/v1beta1:go_default_library",
        "@io_k8s_kubernetes//federation/client/clientset_generated/federation_clientset:go_default_library",
        "@io_k8s_kubernetes//pkg/api/v1:go_default_library",
    ],
)
