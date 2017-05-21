load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "bastion.go",
        "bootstrapscript.go",
        "context.go",
        "convenience.go",
        "dns.go",
        "external_access.go",
        "firewall.go",
        "iam.go",
        "master_volumes.go",
        "names.go",
        "network.go",
        "pki.go",
        "sshkey.go",
        "template_resource.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/util:go_default_library",
        "//pkg/apis/nodeup:go_default_library",
        "//pkg/dns:go_default_library",
        "//pkg/featureflag:go_default_library",
        "//pkg/model/components:go_default_library",
        "//pkg/model/iam:go_default_library",
        "//pkg/model/resources:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/awstasks:go_default_library",
        "//upup/pkg/fi/cloudup/awsup:go_default_library",
        "//upup/pkg/fi/cloudup/gce:go_default_library",
        "//upup/pkg/fi/cloudup/gcetasks:go_default_library",
        "//upup/pkg/fi/fitasks:go_default_library",
        "@com_github_blang_semver//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
        "@io_k8s_apiserver//pkg/authentication/user:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["context_test.go"],
    library = ":go_default_library",
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
    ],
)
