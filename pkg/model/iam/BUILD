load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["iam_builder.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/util/stringorslice:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/awstasks:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["iam_builder_test.go"],
    library = ":go_default_library",
    deps = ["//pkg/util/stringorslice:go_default_library"],
)
