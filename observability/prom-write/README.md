# prom-write

`prom-write` is a tiny CLI utility for testing a *Prometheus Remote Write* endpoint.
It generates a single counter metric and pushes a few samples.

Each request payload is encoded as Prometheus `WriteRequest` protobuf and compressed with *Snappy*.

This tool uses the Prometheus Remote Write protocol v1.
The [API v2](https://pkg.go.dev/github.com/prometheus/client_golang/exp/api/remote) is still experimental.
If a v2 endpoint is used, you will receive an `unsupported proto version` error.

## Quick Start

```bash
go build .
./prom-write <remote_write_url>
```

Example:

```bash
./prom-write http://localhost:9090/api/v1/write
```

## Resources

  - [Snappy](https://github.com/google/snappy)
  - [Prometheus Remote-Write 1.0 specification](https://prometheus.io/docs/specs/prw/remote_write_spec)
  - [Prometheus Remote-Write 2.0 specification](https://prometheus.io/docs/specs/prw/remote_write_spec_2_0)
