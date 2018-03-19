load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "backoff.go",
        "throttle.go",
    ],
    importpath = "k8s.io/client-go/util/flowcontrol",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/juju/ratelimit:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/clock:go_default_library",
        "//vendor/k8s.io/client-go/util/integer:go_default_library",
    ],
)
