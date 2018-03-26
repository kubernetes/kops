load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "helpers.go",
        "validate.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:android": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:darwin": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:dragonfly": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:freebsd": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:nacl": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:netbsd": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:openbsd": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:plan9": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:solaris": [
            "validate_disabled.go",
        ],
        "@io_bazel_rules_go//go/platform:windows": [
            "validate_disabled.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "k8s.io/kubernetes/pkg/security/apparmor",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/util/feature:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/features:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubelet/types:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/util/file:go_default_library",
    ],
)
