load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "rawconn.go",
        "rawconn_stub.go",
        "reflect.go",
        "socket.go",
        "sys.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "cmsghdr.go",
            "cmsghdr_bsd.go",
            "error_unix.go",
            "iovec_64bit.go",
            "mmsghdr_stub.go",
            "msghdr_bsd.go",
            "msghdr_bsdvar.go",
            "rawconn_msg.go",
            "rawconn_nommsg.go",
            "sys_bsd.go",
            "sys_darwin.go",
            "sys_posix.go",
            "sys_unix.go",
            "zsys_darwin_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "cmsghdr.go",
            "cmsghdr_linux_64bit.go",
            "error_unix.go",
            "iovec_64bit.go",
            "mmsghdr_unix.go",
            "msghdr_linux.go",
            "msghdr_linux_64bit.go",
            "rawconn_mmsg.go",
            "rawconn_msg.go",
            "sys_linux.go",
            "sys_linux_amd64.go",
            "sys_posix.go",
            "sys_unix.go",
            "zsys_linux_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "cmsghdr_stub.go",
            "error_windows.go",
            "iovec_stub.go",
            "mmsghdr_stub.go",
            "msghdr_stub.go",
            "rawconn_msg.go",
            "rawconn_nommsg.go",
            "sys_posix.go",
            "sys_windows.go",
        ],
        "//conditions:default": [],
    }),
    cgo = True,
    importpath = "golang.org/x/net/internal/socket",
    visibility = ["//vendor/golang.org/x/net:__subpackages__"],
)

go_test(
    name = "go_default_xtest",
    srcs = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "socket_go1_9_test.go",
            "socket_test.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "socket_go1_9_test.go",
            "socket_test.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "socket_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/net/internal/socket_test",
    deps = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            ":go_default_library",
            "//vendor/golang.org/x/net/internal/nettest:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            ":go_default_library",
            "//vendor/golang.org/x/net/internal/nettest:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            ":go_default_library",
            "//vendor/golang.org/x/net/internal/nettest:go_default_library",
        ],
        "//conditions:default": [],
    }),
)
