# OpenTelemetry support

kOps is experimenting with initial support for OpenTelemetry, starting with tracing.

The support should be considered experimental; the trace file format and schema will likely change, and these initial experiments might be removed entirely.

kOps supports a "serverless" mode of operation, where it logs the OpenTracing output to a file.  We do this because our e2e test runner (prow) doesn't yet have a destination for OpenTelemetry data.

To try this out:

`OTEL_EXPORTER_OTLP_TRACES_FILE=/tmp/trace go run ./cmd/kops get cluster`

You should now see that the /tmp/trace file is created.

Then we have an experimental tool to serve the trace file to jaeger:

```
cd tools/otel/traceserver
go run . --src /tmp/trace --run jaeger
```

Not everything is instrumented yet, and not all the traces are fully joined up (we need to thread more contexts through more methods),
but you should be able to start to explore the operations that we run and their performance.
