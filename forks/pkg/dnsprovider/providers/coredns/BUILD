package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "coredns.go",
        "interface.go",
        "rrchangeset.go",
        "rrset.go",
        "rrsets.go",
        "zone.go",
        "zones.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//forks/pkg/dnsprovider:go_default_library",
        "//forks/pkg/dnsprovider/providers/coredns/stubs:go_default_library",
        "//forks/pkg/dnsprovider/rrstype:go_default_library",
        "@com_github_coreos_etcd//client:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_miekg_coredns//middleware/etcd/msg:go_default_library",
        "@in_gopkg_gcfg_v1//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["coredns_test.go"],
    library = ":go_default_library",
    tags = ["automanaged"],
    deps = [
        "//forks/pkg/dnsprovider:go_default_library",
        "//forks/pkg/dnsprovider/providers/coredns/stubs:go_default_library",
        "//forks/pkg/dnsprovider/rrstype:go_default_library",
        "//forks/pkg/dnsprovider/tests:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//federation/pkg/dnsprovider/providers/coredns/stubs:all-srcs",
    ],
    tags = ["automanaged"],
)
