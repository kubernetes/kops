load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "fileutil.go",
        "lock.go",
        "preallocate.go",
        "purge.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "dir_unix.go",
            "lock_flock.go",
            "lock_unix.go",
            "preallocate_darwin.go",
            "sync_darwin.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "dir_unix.go",
            "lock_flock.go",
            "lock_linux.go",
            "preallocate_unix.go",
            "sync_linux.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "dir_windows.go",
            "lock_windows.go",
            "preallocate_unsupported.go",
            "sync.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "github.com/coreos/etcd/pkg/fileutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fileutil_test.go",
        "lock_test.go",
        "preallocate_test.go",
        "purge_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/fileutil",
    library = ":go_default_library",
)
