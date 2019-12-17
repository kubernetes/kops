def _impl(ctx):
    in_file = ctx.file.src

    out_file = ctx.actions.declare_file("%s.gz" % in_file.path)

    cmd = "gzip -c '%s' > '%s'" % (in_file.path, out_file.path)

    ctx.actions.run_shell(
        outputs = [out_file],
        inputs = [in_file],
        command = cmd,
    )

    return [DefaultInfo(files = depset([out_file]))]

gzip = rule(
    implementation = _impl,
    attrs = {
        "src": attr.label(mandatory = True, allow_single_file = True),
    },
)
