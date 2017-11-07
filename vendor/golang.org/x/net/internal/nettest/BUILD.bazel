load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "interface.go",
        "rlimit.go",
        "stack.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "helper_bsd.go",
            "helper_posix.go",
            "helper_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "helper_nobsd.go",
            "helper_posix.go",
            "helper_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "helper_posix.go",
            "helper_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/net/internal/nettest",
    visibility = ["//vendor/golang.org/x/net:__subpackages__"],
)
