load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "address.go",
            "binary.go",
            "interface.go",
            "interface_classic.go",
            "interface_multicast.go",
            "message.go",
            "route.go",
            "route_classic.go",
            "sys.go",
            "sys_darwin.go",
            "syscall.go",
            "zsys_darwin.go",
        ],
        "//conditions:default": [],
    }),
    cgo = True,
    importpath = "golang.org/x/net/route",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "address_darwin_test.go",
            "address_test.go",
            "message_darwin_test.go",
            "message_test.go",
            "route_test.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/net/route",
    library = ":go_default_library",
)
