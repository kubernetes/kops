load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "mksyscall.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "key.go",
            "syscall.go",
            "value.go",
            "zsyscall_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/registry",
    visibility = ["//visibility:public"],
    deps = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "//vendor/golang.org/x/sys/windows:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_test(
    name = "go_default_test",
    srcs = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "export_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/registry",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "registry_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/windows/registry_test",
    deps = select({
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            ":go_default_library",
        ],
        "//conditions:default": [],
    }),
)
