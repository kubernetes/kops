load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "bastion.go",
        "channel.go",
        "cluster.go",
        "codecs.go",
        "componentconfig.go",
        "doc.go",
        "dockerconfig.go",
        "federation.go",
        "instancegroup.go",
        "labels.go",
        "networking.go",
        "parse.go",
        "register.go",
        "status.go",
        "topology.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops/util:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_blang_semver//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apimachinery/announced:go_default_library",
        "@io_k8s_apimachinery//pkg/apimachinery/registered:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/serializer:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "parse_test.go",
        "semver_test.go",
    ],
    library = ":go_default_library",
    deps = ["@com_github_blang_semver//:go_default_library"],
)
