# Running Services

## Summary

A tiered port scheme makes it easy to identify which tier and service you are communicating with.

| **Container** | **Port** | **Protocol** | **Endpoint** | **Description** |
|----|----|----|----|----|
| **node-exporter** | `3100` | HTTP | `GET /` | *Web app* |
| | | | `GET /metrics` | *Prometheus metrics* |
| **cadvisor** | `3200` | HTTP | `GET /` | *Web app* |
| | | | `GET /metrics` | *Prometheus metrics* |
| **fluent-bit** | `4100` | HTTP | `/` | *Build info* |
| | | | `GET /api/v1/health` | *Health check* |
| | | | `GET /api/v1/metrics` | *Internal metrics* |
| | | | `GET /api/v1/metrics/prometheus` | *Internal Prometheus metrics* |
| | | | `GET/PUT/POST /api/v2/reload` | *Hot reload API* |
| | `4110` | TCP/UDP | | *Forward protocol* |
| | `4130` | HTTP | `POST /api/v1/write` | *Prometheus Remote Write* |
| | `4150` | HTTP/gRPC | `/` | *OpenTelemetry Protocol* |
| | `4170` | HTTP | `GET /metrics` | *Prometheus metrics* |
| **opentelemetry-collector** | `4200` | HTTP | `/health` | *Health check* |
| | `4201` | HTTP | `GET /metrics` | *Internal Prometheus metrics* |
| | `4202` | HTTP | `GET /debug/pprof` | *Go `net/http/pprof` endpoints* |
| | `4203` | HTTP | `GET /debug/servicez` | *zPages: ServiceZ* |
| | | | `GET /debug/tracez` | *zPages: TraceZ* |
| | `4210` | TCP | | *Forward protocol* |
| | `4230` | HTTP | `POST /api/v1/write` | *Prometheus Remote Write* |
| | `4250` | HTTP | `/` | *OpenTelemetry Protocol* |
| | `4260` | gRPC | | *OpenTelemetry Protocol* |
| | `4270` | HTTP | `GET /metrics` | *Prometheus metrics* |
| **alloy** | `4300` | HTTP | `GET /` | *Web app* |
| | | | `GET /-/ready` | *Readiness check* |
| | | | `GET /-/healthy` | *Health check* |
| | | | `GET /-/reload` | *Hot reload* |
| | | | `GET /metrics` | *Internal Prometheus metrics* |
| | `4310` | TCP | | *Forward protocol* |
| | `4320` | HTTP | `POST /loki/api/v1/push` | *Loki Push API* |
| | `4330` | HTTP | `POST /api/v1/metrics/write` | *Prometheus Remote Write* |
| | `4340` | Connect | `POST /push.v1.PusherService/Push` | *Pyroscope Ingestion API* |
| | | HTTP | `POST /ingest` | *Pyroscope Legacy Ingestion API* |
| | `4350` | HTTP | `/` | *OpenTelemetry Protocol* |
| | `4360` | gRPC | | *OpenTelemetry Protocol* |
| **loki** | `5100` | HTTP | `/` | *HTTP API* |
| | | | `GET /ready` | *Readiness check* |
| | | | `GET /metrics` | *Internal Prometheus metrics* |
| | | | `GET /config` | *Current configurations* |
| | | | `GET /services` | *Running services* |
| | | | `POST /loki/api/v1/push` | *Log ingestion* |
| | | | `POST /otlp/v1/logs` | *OpenTelemetry Protocol logs ingestion* |
| | | | `GET/POST/DELETE /loki/api/v1/rules` | *Rules API* |
| | `5110` | gRPC | | *gRPC API* |
| **mimir** | `5200` | HTTP | `/` | *HTTP API* |
| | | | `GET /ready` | *Readiness check* |
| | | | `GET /metrics` | *Internal Prometheus metrics* |
| | | | `GET /config` | *Current configurations* |
| | | | `GET /services` | *Running services* |
| | | | `GET /api/v1/rules` | *Rules* |
| | | | `GET /api/v1/alerts` | *Active Alerts* |
| | | | `POST /api/v1/push` | *Prometheus Remote Write* |
| | | | `POST /otlp/v1/metrics` | *OpenTelemetry Protocol metrics ingestion* |
| | | | `/prometheus` | *Prometheus-compatible API* |
| | `5210` | gRPC | | *gRPC API* |
| **tempo** | `5300` | HTTP | `/` | *HTTP API* |
| | | | `GET /ready` | *Readiness check* |
| | | | `GET /metrics` | *Internal Prometheus metrics* |
| | `5310` | gRPC | | *gRPC API* |
| | `5350` | HTTP | `/` | *OpenTelemetry Protocol* |
| | `5360` | gRPC | | *OpenTelemetry Protocol* |
| **pyroscope** | `5400` | HTTP | `/` | *HTTP API* |
| | | Connect | `POST /push.v1.PusherService/Push` | *Ingestion API* |
| | | | `POST /querier.v1.QuerierService/Diff` | *Querying API* |
| | | HTTP | `POST /ingest` | *Legacy Ingestion API* |
| | | | `GET /pyroscope/render` | *Legacy Querying API* |
| | `5410` | gRPC | | *gRPC API* |
| **alertmanager** | `6100` | HTTP | `GET /` | *Web app* |
| | | | `GET /-/ready` | *Readiness check* |
| | | | `GET /-/healthy` | *Health check* |
| | | | `GET /-/reload` | *Hot reload* |
| **mailpit** | `6200` | HTTP | `GET /` | *Web app* |
| | `6225` | SMTP | | *SMTP API* |
| **renderer** | `7100` | HTTP | `POST /` | *HTTP API* |
| **grafana** | `7200` | HTTP | `GET /` | *Web app* |
| | | | `GET /api/health` | *Health check* |
| | | | `GET /api/org` | *Organization* |
| | | | `GET /api/datasources` | *Data Sources* |
| | | | `GET /api/v1/provisioning/alert-rules/export` | *Alert Rules* |
| | | | `GET /api/v1/provisioning/contact-points/export` | *Contact Points* |
| | | | `GET /api/v1/provisioning/policies/export` | *Notification Polices* |

## Resources

  - **Fluent Bit**
    - [Hot reload](https://docs.fluentbit.io/manual/administration/hot-reload)
    - [Monitoring](https://docs.fluentbit.io/manual/administration/monitoring)
  - **Alloy**
    - [The Grafana Alloy HTTP Endpoints](https://grafana.com/docs/alloy/latest/reference/http)
  - **Loki**
    - [HTTP API](https://grafana.com/docs/loki/latest/reference/loki-http-api)
  - **Mimir**
    - [HTTP API](https://grafana.com/docs/mimir/latest/references/http-api)
  - **Tempo**
    - [HTTP API](https://grafana.com/docs/tempo/latest/api_docs)
  - **Pyroscope**
    - [Server API](https://grafana.com/docs/pyroscope/latest/reference-server-api)
  - **Alertmanager**
    - [Alerts API](https://prometheus.io/docs/alerting/latest/alerts_api)
    - [Management API](https://prometheus.io/docs/alerting/latest/management_api)
  - **Grafana**
    - [HTTP API](https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api)
  - **Mailpit**
    - [API](https://mailpit.axllent.org/docs/api-v1)
