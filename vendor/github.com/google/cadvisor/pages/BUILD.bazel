load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "containers.go",
        "docker.go",
        "pages.go",
        "templates.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/abbot/go-http-auth:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/google/cadvisor/container/docker:go_default_library",
        "//vendor/github.com/google/cadvisor/http/mux:go_default_library",
        "//vendor/github.com/google/cadvisor/info/v1:go_default_library",
        "//vendor/github.com/google/cadvisor/manager:go_default_library",
    ],
)
