load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "log.go",
            "service.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/svc/debug",
    visibility = ["//visibility:public"],
    deps = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "//vendor/golang.org/x/sys/windows/svc:go_default_library",
        ],
        "//conditions:default": [],
    }),
)
