load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "env_unset.go",
        "file_unix.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "asm_darwin_amd64.s",
            "constants.go",
            "dev_darwin.go",
            "dirent.go",
            "endian_little.go",
            "env_unix.go",
            "flock.go",
            "race0.go",
            "sockcmsg_unix.go",
            "str.go",
            "syscall.go",
            "syscall_bsd.go",
            "syscall_darwin.go",
            "syscall_darwin_amd64.go",
            "syscall_unix.go",
            "syscall_unix_gc.go",
            "zerrors_darwin_amd64.go",
            "zsyscall_darwin_amd64.go",
            "zsysnum_darwin_amd64.go",
            "ztypes_darwin_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "asm_linux_amd64.s",
            "bluetooth_linux.go",
            "constants.go",
            "dev_linux.go",
            "dirent.go",
            "endian_little.go",
            "env_unix.go",
            "flock.go",
            "race0.go",
            "sockcmsg_linux.go",
            "sockcmsg_unix.go",
            "str.go",
            "syscall.go",
            "syscall_linux.go",
            "syscall_linux_amd64.go",
            "syscall_linux_amd64_gc.go",
            "syscall_unix.go",
            "syscall_unix_gc.go",
            "zerrors_linux_amd64.go",
            "zsyscall_linux_amd64.go",
            "zsysnum_linux_amd64.go",
            "ztypes_linux_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "endian_little.go",
        ],
        "//conditions:default": [],
    }),
    cgo = True,
    importpath = "golang.org/x/sys/unix",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "export_test.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "export_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/unix",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "dev_darwin_test.go",
            "mmap_unix_test.go",
            "syscall_bsd_test.go",
            "syscall_test.go",
            "syscall_unix_test.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "creds_test.go",
            "dev_linux_test.go",
            "mmap_unix_test.go",
            "syscall_linux_test.go",
            "syscall_test.go",
            "syscall_unix_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/sys/unix_test",
    deps = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            ":go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            ":go_default_library",
        ],
        "//conditions:default": [],
    }),
)
