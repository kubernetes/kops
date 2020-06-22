def _impl(ctx):
    in_file = ctx.file.src

    basename = ctx.attr.src.label.name
    out_sha256 = ctx.actions.declare_file("%s.sha256" % basename)
    ctx.actions.run(
        executable = ctx.executable._cmd_sha256,
        outputs = [out_sha256],
        inputs = [in_file],
        arguments = [in_file.path, out_sha256.path],
    )

    return DefaultInfo(
        files = depset([out_sha256]),
    )

def _get_outputs(src):
    return {
        "sha256": src.name + ".sha256",
    }

hashes = rule(
    implementation = _impl,
    attrs = {
        "src": attr.label(mandatory = True, allow_single_file = True),
        "_cmd_sha256": attr.label(
            default = Label("//tools:sha256"),
            allow_single_file = True,
            executable = True,
            cfg = "host",
        ),
    },
    # We have to do this so that we can reference these outputs in other files
    # https://stackoverflow.com/a/50667861
    outputs = _get_outputs,
)
