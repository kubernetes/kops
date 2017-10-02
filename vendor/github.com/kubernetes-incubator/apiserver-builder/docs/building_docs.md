# Creating reference documentation

This document describes how to build reference documentation for
a Kubernetes apiserver.

## Building default reference documentation


### If using the apiserver-builder framework for your apiserver

1. Build the apiserver binary
  - `apiserver-builder build executables`
2. Generate the docs from the swagger.json
  - `apiserver-build build docs`
  - This will automatically start a server and get the swagger.json from it
  - **Note:** to include docs for operations, use the flag `--operations=true`
3. Open `docs/build/index.html` in a browser

### If *not* using the apiserver-builder framework for your apiserver

1. Get the openapi json
  - Fetch a copy of the "swagger.json" file from your apiserver (located at url /swagger.json), and copy it to docs/ and specify
    the `--build-openapi=false` flag.
2. Generate the docs for your swagger
  - `apiserver-build build docs --build-openapi=false`

## Customizing group descriptions

To add custom descriptions and content to an API group, modify the docs/static_include/_<group>.md file
with your content.  These files are created once when the docs are first generated, but will not overwrite
your changes.

After adding your content, rerun `apiserver-boot build docs`.

## Adding examples

It is highly recommended to add examples of your types to the right-most column.

After adding your content, rerun `apiserver-boot build docs`.

### Type examples

Add an example for your type using in a file

`docs/examples/<type-name>/<type-name>.yaml`

```yaml
note: Description of your example.
sample: |
  apiVersion: <version>
  kind: <kind>
  metadata:
    name: <name>
  spec:
    <spec>
```

### Operation examples

**Note:** Building operations requires providing the `--operations=true` flag.

You can also provide example for operations.  Example operations
contain:

- resource instance name
- resource instance namespace
- yaml request body to the
- json response body from the apiserver

```yaml
name: <resource-name>
namespace: <resource-namespace>
request: |
  <request-yaml>
response: |
  <response-json>
```

The docs framework will automatically create per-language / tool example tabs
using this input.

Locations for operation examples:

- create: `docs/examples/<type-name>/create.yaml`
- delete: `docs/examples/<type-name>/delete.yaml`
- list: `docs/examples/<type-name>/list.yaml`
- patch: `docs/examples/<type-name>/patch.yaml`
- read: `docs/examples/<type-name>/read.yaml`
  - **Note:** read does not have a request section
- replace: `docs/examples/<type-name>/replace.yaml`
- watch: `docs/examples/<type-name>/watch.yaml`

## Customizing dynamic sections

Some of the dynamicly generated sections of the docs maybe statically configured
by provided a `config.yaml` in the docs directory and providing the flag
`--generate-toc=false` when running `build docs`.

Using a config.yaml supports:

- Statically defining the table of contents with customized groupings
- Defining which types are inlined into parent types (e.g. Spec, Status, List)
- Adding notes and warnings to resources
- Grouping subresources
- Redefining display names for operations

See an example [here](https://github.com/kubernetes-incubator/reference-docs/blob/master/gen-apidocs/generators/config.yaml)
