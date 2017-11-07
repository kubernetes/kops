load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "install.go",
            "log.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/svc/eventlog",
    visibility = ["//visibility:public"],
    deps = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "//vendor/golang.org/x/sys/windows:go_default_library",
            "//vendor/golang.org/x/sys/windows/registry:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_test(
    name = "go_default_xtest",
    srcs = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "log_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/svc/eventlog_test",
    deps = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            ":go_default_library",
        ],
        "//conditions:default": [],
    }),
)
