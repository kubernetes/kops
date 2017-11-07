load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "terminal.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "util.go",
            "util_bsd.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "util.go",
            "util_linux.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "util_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/crypto/ssh/terminal",
    visibility = ["//visibility:public"],
    deps = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "//vendor/golang.org/x/sys/unix:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "//vendor/golang.org/x/sys/unix:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "//vendor/golang.org/x/sys/windows:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_test(
    name = "go_default_test",
    srcs = ["terminal_test.go"],
    importpath = "golang.org/x/crypto/ssh/terminal",
    library = ":go_default_library",
)
