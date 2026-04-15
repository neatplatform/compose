# observability

The `compose.yml` file starts a local observability stack by provisioning and configuring observability services as containers.

Use it to quickly connect an application and verify its telemetry pipeline end to end, including logs, metrics, and traces.
This is especially useful for local development, integration testing, and troubleshooting instrumentation before deploying to higher environments.

## Quick Start

```bash
make up    # Start the observability stack
make down  # Stop the observability stack

make list  # Show all running containers
make logs  # Show a container logs
```

## Sanity Checks

```bash
# Inspecting container volumes
podman container run --rm -v observability_fluent_data:/data alpine ls /data
podman container run --rm -v observability_collector_data:/data alpine ls /data
podman container run --rm -v observability_alloy_data:/data alpine ls /data

# Send a test log to fluent-bit over the Fluent Forward protocol.
echo '{"timestamp":"2026-04-06T16:59:00.666666-04:00","level":"info","message":"Hello, World!"}' | \
  podman container run -i --rm --network observability_observability fluent/fluentd:latest \
    fluent-cat --host fluent-bit --port 24224 app.test

# Send a test log to opentelemetry-collector over the Fluent Forward protocol.
echo '{"timestamp":"2026-04-06T16:59:00.666666-04:00","level":"info","message":"Hello, World!"}' | \
  podman container run -i --rm --network observability_observability fluent/fluentd:latest \
    fluent-cat --host opentelemetry-collector --port 8006 app.test

# Send a test log to alloy over the Fluent Forward protocol.
echo '{"timestamp":"2026-04-06T16:59:00.666666-04:00","level":"info","message":"Hello, World!"}' | \
  podman container run -i --rm --network observability_observability fluent/fluentd:latest \
    fluent-cat --host alloy --port 8006 app.test
```

## Tools

### Collectors

#### Ingress

| **Service** | **Log Files** | **Forward Protocol** | **Prometheus** (Pull) | **Prometheus** (Push) | **OTLP HTTP** | **OTLP gRPC** |
|---|:----:|:----:|:----:|:----:|:----:|:----:|
| Fluent Bit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OTEL Collector | ✅ | ✅ | ✅ <sup>1</sup> | ✅ | ✅ | ✅ |
| Alloy | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

  1. The OpenTelemetry Collector supports the experimental [Prometheus Remote Write v2](https://prometheus.io/docs/specs/prw/remote_write_spec_2_0).
     Prometheus Remote Write requests sent to the OpenTelemetry Collector using the v1 API will not be accepted.

#### Egress

| **Service** | **Stdout** | **File** | **Prometheus** (Pull) | **Prometheus** (Push) | **OTLP HTTP** | **OTLP gRPC** | **Loki** | **Mimir** | **Tempo** | **Pyroscope** |
|----|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|
| Fluent Bit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | TBD | TBD | TBD | TBD |
| OTEL Collector | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | TBD | TBD | TBD | TBD |
| Alloy | ✅ | ✅ | ❌ <sup>1</sup> | ✅ | ✅ | ✅ | TBD | TBD | TBD | TBD |

  1. Alloy does not support exposing collected metrics at the `/metrics` HTTP endpoint.
     The default `/metrics` endpoint available on the server port (`12345`) only exposes Alloy's internal Prometheus metrics.
     Collected Prometheus metrics can be forwarded to a backend storage (such as Prometheus or Mimir) via Prometheus Remote Write.

## Services

| **Container** | **Port** | **Protocol** | **Endpoint** | **Description** |
|----|----|----|----|----|
| **node-exporter** | `2100` | HTTP | `/` | *Web app* |
| | | | `/metrics` | *Prometheus metrics* |
| **cadvisor** | `2200` | HTTP | `/` | *Web app* |
| | | | `/metrics` | *Prometheus metrics* |
| **fluent-bit** | `3100` | HTTP | `/` | *Build info* |
| | | | `/api/v1/uptime` | *Uptime* |
| | | | `/api/v1/plugins` | *Plugins* |
| | | | `/api/v1/health` | *Health check* |
| | | | `/api/v1/metrics` | *Internal metrics* |
| | | | `/api/v1/metrics/prometheus` | *Internal Prometheus metrics* |
| | | | `/api/v2/reload` | *Hot reload (`GET`/`PUT`/`POST`)* |
| | `3110` | TCP/UDP | | *Forward protocol* |
| | `3120` | HTTP | `/api/v1/write` | *Prometheus Remote Write* |
| | `3130` | HTTP/gRPC | `/` | *OpenTelemetry protocol* |
| | `3180` | HTTP | `/metrics` | *Prometheus metrics* |
| **opentelemetry-collector** | `3200` | HTTP | `/health` | *Health check* |
| | `3201` | HTTP | `/metrics` | *Internal Prometheus metrics* |
| | `3202` | HTTP | `/debug/pprof` | *Go `net/http/pprof` endpoints* |
| | `3203` | HTTP | `/debug/servicez` | *zPages: ServiceZ* |
| | | | `/debug/tracez` | *zPages: TraceZ* |
| | `3210` | TCP | | *Forward protocol* |
| | `3220` | HTTP | `/api/v1/write` | *Prometheus Remote Write* |
| | `3230` | HTTP | `/` | *OpenTelemetry protocol* |
| | `3232` | gRPC | `/` | *OpenTelemetry protocol* |
| | `3280` | HTTP | `/metrics` | *Prometheus metrics* |
| **alloy** | `3300` | HTTP | `/` | *Web app* |
| | | | `/metrics` | *Internal Prometheus metrics* |
| | | | `/-/ready` | *Readiness check* |
| | | | `/-/healthy` | *Health check* |
| | | | `/-/reload` | *Hot reload* |
| | | | `/debug/pprof` | *Go `net/http/pprof` endpoints* |
| | `3310` | TCP | | *Forward protocol* |
| | `3320` | HTTP | `/api/v1/metrics/write` | *Prometheus Remote Write* |
| | `3330` | HTTP | `/` | *OpenTelemetry protocol* |
| | `3332` | gRPC | `/` | *OpenTelemetry protocol* |
| **prometheus** | `4100` | HTTP | `/` | *Web app* |
| | | | `/metrics` | *Internal Prometheus metrics* |
| | | | `/-/healthy` | *Health check* |
| | | | `/-/ready` | *Readiness check* |
| | | | `/-/reload` | *Hot reload* |
| **alertmanager** | `4110` | HTTP | `/` | *Web app* |
| | | | `/-/healthy` | *Health check* |
| | | | `/-/ready` | *Readiness check* |
| | | | `/-/reload` | *Hot reload* |

## Pipelines

```
(File, Fluent Forward         )  →  [ INPUTS → FILTERS → OUTPUTS         ]  →  (Stdout, File                  )
(Prometheus Scrape/RemoteWrite)  →  [ INPUTS → FILTERS → OUTPUTS         ]  →  (Prometheus metrics/RemoteWrite)
(OpenTelemetry HTTP/gRPC      )  →  [ INPUTS → FILTERS → OUTPUTS         ]  →  (OpenTelemetry HTTP/gRPC       )

(File, Fluent Forward         )  →  [ RECEIVERS → PROCESSORS → EXPORTERS ]  →  (Stdout, File                  )
(Prometheus Scrape/RemoteWrite)  →  [ RECEIVERS → PROCESSORS → EXPORTERS ]  →  (Prometheus metrics/RemoteWrite)
(OpenTelemetry HTTP/gRPC      )  →  [ RECEIVERS → PROCESSORS → EXPORTERS ]  →  (OpenTelemetry HTTP/gRPC       )

(File, Fluent Forward         )  →  [ COMPONENT → COMPONENT → COMPONENT  ]  →  (Stdout, File                  )
(Prometheus Scrape/RemoteWrite)  →  [ COMPONENT → COMPONENT → COMPONENT  ]  →  (Prometheus RemoteWrite        )
(OpenTelemetry HTTP/gRPC      )  →  [ COMPONENT → COMPONENT → COMPONENT  ]  →  (OpenTelemetry HTTP/gRPC       )
```

## Considerations

### cAdvisor

[cAdvisor](https://github.com/google/cadvisor) is closely tied to Docker internals,
and support for Podman – especially on macOS – is quite limited.
Many required host volumes that cAdvisor relies on in a Docker or Linux environment are not available to mount.

  - `/dev/disk/:/dev/disk:ro`
  - `/var/run:/var/run:rw`
  - `/var/lib/docker:/var/lib/docker:ro`
  - `/var/lib/containers:/var/lib/containers:ro`

### Forward Protocol

Unlike Docker, Podman does not support the [Fluentd logging driver](https://docs.docker.com/engine/logging/drivers/fluentd/)
for shipping container logs over the [Fluentd Forward Protocol](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1).
To use *Forward* in this setup, applications and services should send logs directly to a *Fluentd Forward* endpoint.
This can be done with a *Fluentd* client library or by implementing the *Forward* protocol directly
and sending records in [JSON](https://www.json.org) or [MessagePack](https://msgpack.org) format.

### Slack Alerts

Alertmanager's `slack_configs` integration is built on Slack's legacy *Incoming Webhooks* and attachments format, which predates *Block Kit*.
Because `slack_configs` does not expose *Block Kit* fields, you cannot send rich messages through this integration.

If you need *Block Kit* formatting, use `webhook_config` to forward alert notifications to an intermediary service.
That service can receive the webhook payload, construct a *Block Kit* message, and send it to Slack through the Web API using `chat.postMessage`.
This requires a proper Slack app with the `chat:write` scope instead of an *Incoming Webhook*.
If you are already running Grafana, its built-in alerting supports this flow natively and can remove Alertmanager from the notification path.

## Resources

  - **Collectors**
    - **Fluent Bit**
      - [Key concepts](https://docs.fluentbit.io/manual/concepts/key-concepts)
      - [Data pipeline](https://docs.fluentbit.io/manual/concepts/data-pipeline)
      - [Backpressure](https://docs.fluentbit.io/manual/administration/backpressure)
      - [Hot reload](https://docs.fluentbit.io/manual/administration/hot-reload)
      - [Monitoring](https://docs.fluentbit.io/manual/administration/monitoring)
      - [TLS](https://docs.fluentbit.io/manual/administration/transport-security)
    - **OpenTelemetry**
      - [Collector](https://opentelemetry.io/docs/collector)
      - [Architecture](https://opentelemetry.io/docs/collector/architecture)
      - [Management](https://opentelemetry.io/docs/collector/management)
      - [Scaling the Collector](https://opentelemetry.io/docs/collector/scaling)
    - **Alloy**
      - [Grafana Alloy](https://grafana.com/docs/alloy/latest)
      - [Components](https://grafana.com/docs/alloy/latest/get-started/components)
      - [OpenTelemetry in Alloy](https://grafana.com/docs/alloy/latest/introduction/otel_alloy)
      - [The Alloy OpenTelemetry Engine](https://grafana.com/docs/alloy/latest/set-up/otel_engine)
      - [Deploy Grafana Alloy](https://grafana.com/docs/alloy/latest/set-up/deploy)
      - [Choose a Grafana Alloy Component](https://grafana.com/docs/alloy/latest/collect/choose-component)
      - [The Grafana Alloy HTTP Endpoints](https://grafana.com/docs/alloy/latest/reference/http)
  - **Backends**
    - **Prometheus**
      - [Data model](https://prometheus.io/docs/concepts/data_model)
      - [Metric and label naming](https://prometheus.io/docs/practices/naming)
      - [Querying basics](https://prometheus.io/docs/prometheus/latest/querying/basics)
      - [Query functions](https://prometheus.io/docs/prometheus/latest/querying/functions)
      - [HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api)
      - [Management API](https://prometheus.io/docs/prometheus/latest/management_api)
    - **Alertmanager**
      - [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)
      - [High Availability](https://prometheus.io/docs/alerting/latest/high_availability)
      - [Alerts API](https://prometheus.io/docs/alerting/latest/alerts_api)
      - [Management API](https://prometheus.io/docs/alerting/latest/management_api)
  - **Misc**
    - **Slack**
      - [Block Kit](https://docs.slack.dev/block-kit)
      - [Block Kit Builder](https://app.slack.com/block-kit-builder)
      - [`chat.postMessage` method](https://docs.slack.dev/reference/methods/chat.postMessage)
      - [Sending messages using incoming webhooks](https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks)
