# Data Models

## Logs

Developers are usually used to structured logging
where each log entry is a JSON object with a timestamp, level, message, and contextual key-value pairs.
However, log collectors and storage backends each define their own data model,
so understanding these models and the conversions between them is essential when designing a logging pipeline.

Application loggers that provide an interface for leveled logging and structured key-value pairs should provide
abstractions to emit logs in the format expected by the wire protocols they target (Forward, OTLP, Loki, etc.).

Furthermore, when receiving logs in one format and forwarding them
to a storage backend or another collector in a different format, the data will be converted.
You may have full or partial control over these mappings depending on the collector you use.
For example, if you receive logs via the Forward protocol using the OpenTelemetry Collector,
the logs will be converted from Fluent format to the OpenTelemetry log data model.
Similarly, when logs are forwarded to Loki from Fluent Bit or the OpenTelemetry Collector,
they must be mapped to Loki's stream-based format.
Loki requires low-cardinality stream labels for indexing,
so understanding which fields become labels is critical for performance and query efficiency.

When designing a log shipping pipeline, consider the data model requirements of the final storage destination.
For example, ingesting logs through the OpenTelemetry Collector's Forward receiver
produces differently structured OpenTelemetry log records than ingesting them directly via the OpenTelemetry Protocol (OTLP).
Each conversion step can rename fields, change nesting, or lose information,
so it pays to trace the full path from application to storage backend before committing to a pipeline design.

### Fluent Bit

*Fluent Bit* represents events using the following format:

```
[TAG]: [[TIMESTAMP, METADATA], MESSAGE]
```

  - `TAG` is the Fluent tag used for routing and classifying log sources.
    Tags drive the `INPUTS → FILTERS → OUTPUTS` routing pipeline via tag matching rules that support wildcards and prefixes.
  - `TIMESTAMP` is the event timestamp, encoded as a floating-point value in seconds with nanosecond precision.
  - `METADATA` is an optional object containing event metadata (introduced in *v2.1.0*).
  - `MESSAGE` is the log record object containing the log body and any structured fields.

Prior to version *v2.1.0*, the format did not include the metadata field:

```
[TAG]: [TIMESTAMP, MESSAGE]
```

Example output from the `stdout` output plugin:

```
logs.test: [1776488604.028796000, {"timestamp":"2026-04-18T01:03:24.028796-04:00","level":"info","message":"Hello, World!"}]
```

### OpenTelemetry

OpenTelemetry defines its own data model for logs, metrics, and traces.
An OpenTelemetry log payload sent over OTLP has the following structure:

```json
{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [...]
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "...",
            "version": "...",
            "attributes": [...]
          },
          "logRecords": [
            {
              "timeUnixNano": "...",
              "observedTimeUnixNano": "...",
              "severityNumber": ...,
              "severityText": "...",
              "body": {
                "stringValue": "..."
              },
              "attributes": [...]
            }
          ]
        }
      ],
      "schemaUrl": "..."
    }
  ]
}
```

  - `resource` describes the origin of the logs (e.g., service name, host, deployment environment).
    All logs produced by the same process share the same resource. It is associated with the `LoggerProvider`.
  - `scopeLogs` group log records by instrumentation scope.
    - `scope` identifies the instrumentation library that produced the logs.
      It is associated with a `Logger` and a single process may have multiple loggers.
    - `logRecords` are the individual log entries, each containing timestamp, severity, body, and attributes.

Attributes can be set at three independent levels: `resource`, `scope`, and `logRecord`.
Keep high-cardinality fields (*user IDs*, *request IDs*) as `logRecord` attributes to keep `resource` and `scope` lean (low-cardinality).

### Loki

Loki stores logs as **streams**, where each stream is a sequence of log lines that share the same set of labels.
Labels are **low-cardinality** key-value pairs used for indexing and routing.
The log line itself is an opaque string — structured fields inside the line are only parsed at query time using *LogQL*.

Logs are pushed to Loki via its HTTP push API as JSON:

```json
{
  "streams": [
    {
      "stream": {
        "service": "api",
        "env": "production"
      },
      "values": [
        ["1745136204028796000", "{\"level\":\"info\",\"message\":\"Hello, World!\"}"],
        ["1745136205100000000", "{\"level\":\"error\",\"message\":\"Something failed\"}"]
      ]
    }
  ]
}
```

**Key implications:**

  - **Label cardinality matters.**
    - Each unique label set creates a new stream. Keep labels limited to a small, stable set of dimensions.
    - Using high-cardinality values as labels will cause label explosion and seriously degrade Loki's ingestion and query performance.
  - **Log lines are unstructured at ingest.**
    - Loki stores the raw log line without parsing it.
    - Structured fields are extracted only at query time with LogQL's parsers.
    - Schema changes in your application logs therefore do not require re-indexing.
  - **Collector label mapping.**
    - When forwarding logs to Loki, collectors extract specific fields to use as stream labels.
    - For example, Fluent Bit's Loki output plugin maps the Fluent tag and specified record fields to labels.
    - Understanding this mapping is critical to avoid high-cardinality label problems and to ensure logs land in the expected streams.
  - **OTLP ingestion.**
    - Loki also accepts logs directly via OTLP at `/otlp/v1/logs`.
    - When receiving OTLP logs, Loki maps OpenTelemetry resource attributes (e.g., `service.name`) to stream labels automatically.
    - The log body becomes the log line.

## Metrics

Most people are familiar with the text-based Prometheus exposition format,
where each sample is a time-series value identified by a metric name and a set of low-cardinality labels.

When collectors pull or receive metrics, they may convert them from one format to another.
For example, if you scrape a Prometheus `/metrics` endpoint using the OpenTelemetry Collector,
the metrics are converted from the Prometheus text format to the OpenTelemetry metrics data model.

### Prometheus

Prometheus uses a simple text-based exposition format.
Each time series is identified by a metric name and an optional set of key-value labels.

Metric names must start with a letter and may contain letters, digits, `_`, and `:`.
Characters like `.` and `-` are not allowed.
Collectors that convert from formats where these characters are valid will replace them.

```
<metric_name>{<label_name>="<label_value>", ...} <value> [<timestamp>]
```

Example:

```
http_requests_total{method="POST", status="200"} 1027 1395066363000
```

Labels starting with `__` are reserved for internal use.
For example, the metric name is internally represented as the `__name__` label.

### OpenTelemetry

OpenTelemetry metrics use their own data model with explicit metric types.
An OpenTelemetry metrics payload has the following structure:

```json
{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": [...]
      },
      "scopeMetrics": [
        {
          "scope": {
            "name": "...",
            "version": "...",
            "attributes": [...]
          },
          "metrics": [
            {
              "name": "...",
              "unit": "...",
              "...": {
                "dataPoints": [
                  {
                    "attributes": [...],
                    "startTimeUnixNano": "...",
                    "timeUnixNano": "...",
                  },
                ],
              }
            }
          ]
        }
      ]
    }
  ]
}
```

  - `resource` describes the origin of the metrics (e.g., service name, host, environment).
    All metrics from the same process share the same resource. It is associated with the `MeterProvider`.
  - `scopeMetrics` group metrics by instrumentation scope.
    - `scope` identifies the instrumentation library.
      It is associated with a `Meter` and a single process may have multiple meters.
    - `metrics` are the individual metric definitions, each containing data points with timestamps, attributes, and type-specific fields.

Attributes can be set at three independent levels: `resource`, `scope`, and `dataPoint` (equivalent to Prometheus labels).
Keep high-cardinality dimensions at the `dataPoint` level.

**Conversion implications:** When the OpenTelemetry Collector scrapes a Prometheus endpoint,
metric names undergo transformation: `.` is replaced with `_`.
When exporting back to Prometheus format, the reverse transformation applies.

## Traces

Unlike logs and metrics, tracing has largely converged on a single standard: OpenTelemetry.
Legacy formats like Zipkin and Jaeger are still present in many environments,
but both now provide OpenTelemetry-native ingestion paths and OpenTelemetry is the recommended choice for new instrumentation.

### OpenTelemetry

OpenTelemetry traces follow the same three-level structure as logs and metrics.
An OpenTelemetry traces payload has the following structure:

```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [...]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "...",
            "version": "...",
            "attributes": [...]
          },
          "spans": [
            {
              "traceId": "...",
              "spanId": "...",
              "parentSpanId": "...",
              "name": "...",
              "startTimeUnixNano": "...",
              "endTimeUnixNano": "...",
              "attributes": [...],
              "events": [
                {
                  "timeUnixNano": "...",
                  "name": "...",
                  "attributes": [...]
                }
              ]
            },
            {
              "traceId": "...",
              "spanId": "...",
              "parentSpanId": "...",
              "name": "...",
              "startTimeUnixNano": "...",
              "endTimeUnixNano": "...",
              "attributes": [...],
              "links": [
                {
                  "traceId": "...",
                  "spanId": "...",
                  "attributes": [...]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

  - `resource` describes the origin of the traces (e.g., service name, version, deployment environment).
    All spans from the same process share the same resource. It is associated with the `TracerProvider`.
  - `scopeSpans` group spans by instrumentation scope.
    - `scope` identifies the instrumentation library.
      It is associated with a `Tracer` and a single process may have multiple tracers.
    - `spans` are the individual trace segments, each representing a unit of work.

Attributes can be set at three independent levels: `resource`, `scope`, and `span`.
Span-level attributes are the most specific and directly queryable in Tempo.

## Resources

  - **Fluent Bit**
    - [Event Format](https://docs.fluentbit.io/manual/concepts/key-concepts#event-format)
  - **Prometheus**
    - [Data Model](https://prometheus.io/docs/concepts/data_model)
  - **Loki**
    - [Data Format](https://grafana.com/docs/loki/latest/get-started/architecture/#data-format)
    - [Labels](https://grafana.com/docs/loki/latest/get-started/labels)
    - [Structured Metadata](https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata)
  - **OpenTelemetry**
    - [Logs Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model)
    - [Metrics Data Model](https://opentelemetry.io/docs/specs/otel/metrics/data-model)
    - [Tracing API](https://opentelemetry.io/docs/specs/otel/trace/api)
