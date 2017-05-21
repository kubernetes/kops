load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "clientset.go",
        "cluster.go",
        "commonvfs.go",
        "federation.go",
        "import_known_versions.go",
        "instancegroup.go",
        "utils.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/install:go_default_library",
        "//pkg/apis/kops/registry:go_default_library",
        "//pkg/apis/kops/v1alpha1:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "//pkg/apis/kops/validation:go_default_library",
        "//pkg/client/simple:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
        "@io_k8s_kubernetes//pkg/api/errors:go_default_library",
    ],
)
