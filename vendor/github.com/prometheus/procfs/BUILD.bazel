load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "buddyinfo.go",
        "doc.go",
        "fs.go",
        "ipvs.go",
        "mdstat.go",
        "mountstats.go",
        "proc.go",
        "proc_io.go",
        "proc_limits.go",
        "proc_stat.go",
        "stat.go",
        "xfrm.go",
    ],
    importpath = "github.com/prometheus/procfs",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/prometheus/procfs/xfs:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "buddyinfo_test.go",
        "fs_test.go",
        "ipvs_test.go",
        "mdstat_test.go",
        "mountstats_test.go",
        "proc_io_test.go",
        "proc_limits_test.go",
        "proc_stat_test.go",
        "proc_test.go",
        "stat_test.go",
        "xfrm_test.go",
    ],
    importpath = "github.com/prometheus/procfs",
    library = ":go_default_library",
)
