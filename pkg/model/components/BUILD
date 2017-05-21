load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "apiserver.go",
        "context.go",
        "defaults.go",
        "docker.go",
        "kubecontrollermanager.go",
        "kubedns.go",
        "kubelet.go",
        "kubeproxy.go",
        "kubescheduler.go",
        "networking.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/util:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/gce:go_default_library",
        "//upup/pkg/fi/loader:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_blang_semver//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["kubecontrollermanager_test.go"],
    library = ":go_default_library",
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
    ],
)
