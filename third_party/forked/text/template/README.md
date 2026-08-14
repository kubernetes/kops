# text/template

This is a copy of the Go standard library `text/template` package (from Go 1.26.5), with a single functional change: template execution does not resolve field names to methods on the data being rendered (`exec.go`, `evalField`).

## Why

The stdlib template executor calls `reflect.Value.MethodByName` with a non-constant argument. Linking that code disables the Go linker's method pruning for the entire binary: every exported method of every reachable type is kept, which costs roughly 80 MB per binary in `kops` and `kops-controller` (mostly unused AWS/GCP/Azure SDK and client-go method code and its metadata).

kOps templates (nodeup bootstrap script, addon manifests, `kops toolbox template`) render values from config structs, maps, and template functions. None of them invoke methods on the templated data, so this fork is a source-compatible replacement at those call sites.

## What changed

- `exec.go`: the method-resolution branch in `evalField` is removed.
- `funcs.go`: `FuncMap` aliases the stdlib type, preserving exact API type identity without making the stdlib executor reachable.
- `internal/fmtsort`: copy of the stdlib-internal package, unchanged; it preserves the stdlib's sorted map iteration in `range`.
- Import paths are updated accordingly. Parsing still uses the stdlib `text/template/parse` package; apart from the changes above and import paths, the copied code is byte-identical to its upstream version.

Do not use this package for new template rendering without checking that method resolution on template data is not needed, and do not reintroduce reachable stdlib `text/template` or `html/template` execution elsewhere: a single reachable copy of the stdlib executor restores the full method-pruning penalty. When updating Go, re-copy the sources (including `internal/fmtsort`), reapply the changes above, and run this package's tests. golang/go#72895 (`Template.ExecuteLite`) would make this fork unnecessary if accepted and the call sites migrated.
