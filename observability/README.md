# observability

The `compose.yml` file starts a local observability stack by provisioning and configuring observability services as containers.

Use it to quickly connect an application and verify its telemetry pipeline end to end, including logs, metrics, and traces.
This is especially useful for local development, integration testing, and troubleshooting instrumentation before deploying to higher environments.

## Quick Start

To post alerts to Slack, create a Slack app and export its Bot User OAuth Token as shown below.
You can create a new Slack app from a manifest by following the instructions [here](./slack/README.md#create-a-new-slack-app).

```bash
export SLACK_APP_AUTH_TOKEN="..."

make up    # Start the observability stack
make down  # Stop the observability stack

make list  # Show all running containers
make logs  # Show a container logs

open http://localhost:7200
```

The Grafana admin credentials are `admin`/`coffee`. A non-default password is required because
Grafana prompts for a password change on first login when the default `admin` password is used.

### Sanity Checks

```bash
# Send synthetic logs, metrics, traces, and pprof
cd volt
make
./volt

# Inspecting container volumes
podman container run --rm -v <volume_name>:/data alpine ls /data
```

## Services

### Data Collection

**Incoming**

| **Service** | **Log Files** | **Forward Protocol** | **Prometheus** (Pull) | **Prometheus** (Push) | **OTLP HTTP** | **OTLP gRPC** |
|----|:----:|:----:|:----:|:----:|:----:|:----:|
| Fluent Bit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OTEL Collector | ✅ | ✅ | ✅ <sup>1</sup> | ✅ | ✅ | ✅ |
| Alloy | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

  1. The OpenTelemetry Collector supports the experimental [Prometheus Remote Write v2](https://prometheus.io/docs/specs/prw/remote_write_spec_2_0).
     Prometheus Remote Write requests sent to the OpenTelemetry Collector using the v1 API will not be accepted.

**Outgoing**

| **Service** | **Stdout** | **File** | **Prometheus** (Pull) | **Prometheus** (Push) | **OTLP HTTP** | **OTLP gRPC** |
|----|:----:|:----:|:----:|:----:|:----:|:----:|
| Fluent Bit | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OTEL Collector | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Alloy | ✅ | ✅ | ❌ <sup>1</sup> | ✅ | ✅ | ✅ |

  1. Alloy does not support exposing collected metrics at the `/metrics` HTTP endpoint.
     The default `/metrics` endpoint available on the server port (`12345`) only exposes Alloy's internal Prometheus metrics.
     Collected Prometheus metrics can be forwarded to a backend storage (such as Prometheus or Mimir) via Prometheus Remote Write.

### Data Storage

| **Service** | **Loki Push API** | **Prometheus Remote Write** | **OpenTelemetry Protocol** | **Connect Protocol** |
|----|----|----|----|----|
| Loki | ✅ | | ✅ | |
| Mimir | | ✅ | ✅ | |
| Tempo | | | ✅ | |
| Pyroscope | | | | ✅ |

## Signal Correlation

Grafana connects your observability signals — logs, metrics, traces, and profiles — into a unified experience,
letting you move seamlessly between data sources to investigate issues faster.

  - **Traces → Logs**: Navigate from a span in Tempo directly to its matching logs in Loki.
  - **Traces → Metrics**: Jump from a span in a trace to related metrics in any configured metrics data source.
  - **Traces → Profiles**: Link from a span in a trace to the corresponding profiling data in Grafana Pyroscope.
  - **Logs → Traces**: Use *Derived Fields* to extract values from a log line and build links that open the matching trace in Tempo.

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

### Alertmanager Configuration

Alertmanager does not allow the direct use of environment variables inside the config file.
To inject configuration values and secrets through environment variables,
we preprocess a template file and substitute variable placeholders before startup.
For this reason, the `Makefile` includes a `prep` rule that resolves environment variable references in the Alertmanager config file.

## Read More

  - [Data Models](./docs/data-models.md)
  - [Data Pipelines](./docs/data-pipelines.md)
  - [Alerting](./docs/alerting.md)
  - [Running Services](./docs/running-services.md)
  - [Best Practices](./docs/best-practices.md)

## Resources

  - **Collectors**
    - **Fluent Bit**
      - [Key concepts](https://docs.fluentbit.io/manual/concepts/key-concepts)
      - [Data pipeline](https://docs.fluentbit.io/manual/concepts/data-pipeline)
    - **OpenTelemetry**
      - [Collector](https://opentelemetry.io/docs/collector)
      - [Architecture](https://opentelemetry.io/docs/collector/architecture)
    - **Alloy**
      - [Grafana Alloy](https://grafana.com/docs/alloy/latest)
      - [Components](https://grafana.com/docs/alloy/latest/get-started/components)
  - **Backends**
    - **Prometheus**
      - [Data model](https://prometheus.io/docs/concepts/data_model)
      - [Querying basics](https://prometheus.io/docs/prometheus/latest/querying/basics)
      - [Query functions](https://prometheus.io/docs/prometheus/latest/querying/functions)
    - **Loki**
      - [Grafana Loki](https://grafana.com/docs/loki/latest)
      - [Architecture](https://grafana.com/docs/loki/latest/get-started/architecture)
      - [Components](https://grafana.com/docs/loki/latest/get-started/components)
      - [Consistent Hash Rings](https://grafana.com/docs/loki/latest/get-started/hash-rings)
      - [Query Loki](https://grafana.com/docs/loki/latest/query)
    - **Mimir**
      - [Grafana Mimir](https://grafana.com/docs/mimir/latest)
      - [Architecture](https://grafana.com/docs/mimir/latest/get-started/about-grafana-mimir-architecture)
    - **Tempo**
      - [Grafana Tempo](https://grafana.com/docs/tempo/latest)
      - [Architecture](https://grafana.com/docs/tempo/latest/introduction/architecture)
    - **Pyroscope**
      - [Grafana Pyroscope](https://grafana.com/docs/pyroscope/latest)
      - [Architecture](https://grafana.com/docs/pyroscope/latest/reference-pyroscope-architecture/about-grafana-pyroscope-architecture)
  - **Frontends**
    - **Grafana**
      - [Grafana](https://grafana.com/docs/grafana/latest)
      - [Provision Grafana](https://grafana.com/docs/grafana/latest/administration/provisioning)
      - [Exemplars](https://grafana.com/docs/grafana/latest/fundamentals/exemplars)
