# volt

A CLI tool for sanity-checking the observability stack.
It can run mock services and emit synthetic logs, metrics, traces, and performance profiles.

## Quick Start

```bash
make
./volt
```

## Commands

### `service`

#### `echo`

  - An HTTP server that echoes incoming requests.
  - Listens on port `8080` by default (override with `-port`).
  - Exposes metrics at `/metrics` and pprof profiles at `/debug/pprof`.

#### `webhook`

  - An HTTP server that receives and logs Grafana Alertmanager alert payloads.
  - Listens on port `8080` by default (override with `-port`).
  - Handles the `/webhook` endpoint by default (override with `-endpoint`).
  - Use the `-ngrok` flag to expose the service to the internet through an ngrok tunnel.
  - Exposes metrics at `/metrics` and pprof profiles at `/debug/pprof`.

> To use the `-ngrok` flag, get an ngrok auth token from https://dashboard.ngrok.com
> and set it with the `NGROK_AUTHTOKEN` environment variable.

### `signal`

Sends a small batch of synthetic signals to a running service to verify that each pipeline is correctly wired end-to-end.

#### `forward`

  - Sends a sequence of log entries to a Forward endpoint.
  - Uses MessagePack encoding by default; pass `-json` to use JSON.
  - Each entry includes a `timestamp`, `level`, `message`, and key-value pairs.

#### `loki`

  - Pushes a log stream to a Loki endpoint via the HTTP Push API.
  - Uses JSON encoding by default; pass `-proto` to use protobuf.
  - Both encodings are Snappy-compressed.

#### `prometheus`

  - Writes a counter time-series to a Prometheus Remote Write endpoint.
  - Uses JSON encoding by default; pass `-proto` to use protobuf.
  - Both encodings are Snappy-compressed.

It uses Remote Write protocol v1.
The [v2 API](https://pkg.go.dev/github.com/prometheus/client_golang/exp/api/remote) is still experimental;
sending to a v2 endpoint returns an `unsupported proto version` error.

#### `opentelemetry`

  - Exports logs, metrics, and traces to an OpenTelemetry collector via OTLP.
  - All three signals are always sent.
  - Uses HTTP by default; pass `-grpc` to use gRPC.
  - All exporters use gzip compression and connect without TLS.

### `template`

  - Tests Grafana notification templates against Alertmanager payloads.
  - Validates template output against sample JSON before using templates in alerting workflows.

## Resources

  - **Protocols**
    - [Forward Protocol Specification v1](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1)
    - [Loki HTTP Push API](https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs)
    - [Prometheus Remote-Write 1.0 specification](https://prometheus.io/docs/specs/prw/remote_write_spec)
    - [Prometheus Remote-Write 2.0 specification](https://prometheus.io/docs/specs/prw/remote_write_spec_2_0)
    - [OpenTelemetry Protocol Specification](https://opentelemetry.io/docs/specs/otlp)
