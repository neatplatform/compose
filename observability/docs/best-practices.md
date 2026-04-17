# Best Practices

When designing telemetry collection, aggregation, storage, and presentation pipelines in a container-based architecture (e.g., Kubernetes),
the right approach depends on the number of services, application scale, geographic distribution, and organizational structure.
Solutions range from fully centralized to highly distributed, but a common goal is a
**single pane of glass**: one place to query logs, metrics, and traces with automatic correlation across signal types.

There are three layers to consider:

  - **Collectors:** Fluent Bit, Alloy, OpenTelemetry Collector
  - **Storage Backends:** Prometheus, Loki, Mimir, Tempo, Pyroscope
  - **Querying and Dashboarding:** Grafana

## Collector Choice

The choice of collector affects what protocols are natively supported and how much operational complexity you take on:

  - **Fluent Bit** is lightweight and excels at log collection and routing.
    It supports Forward protocol, file tailing, and multiple outputs including Loki, OTLP, and Prometheus Remote Write.
    Best for log-focused pipelines or resource-constrained environments.
  - **OpenTelemetry Collector** is the most protocol-agnostic option and handles all three signal types (logs, metrics, traces) uniformly.
    It is the best choice when standardizing on OTLP end-to-end or when routing data across heterogeneous backends.
  - **Alloy** provides a powerful pipeline DSL and is tightly integrated with the Grafana stack.
    Best when you are fully committed to the Grafana ecosystem and want native integration with Loki, Mimir, Tempo, and Pyroscope.

## Deployment Modes

When deploying collectors in Kubernetes, two common patterns exist:

**DaemonSet (node-level collector)**

A single collector pod runs on each Kubernetes node and collects telemetry from all pods on that node, typically by tailing log files under `/var/log/pods`.

  - Lower resource overhead — one collector per node regardless of the number of pods.
  - Log tailing requires the collector to parse the log format and attach the right metadata
    (pod name, namespace, container name) from the file path or Kubernetes API.
  - Limited per-application configuration flexibility — all pods on a node share a collector.
  - Best for organizations with homogeneous application stacks and standardized log formats.

**Sidecar (pod-level collector)**

A collector container is injected into each application pod and collects telemetry directly from the application process via localhost.

  - Higher resource usage — a collector process per pod adds CPU and memory overhead.
  - The application and collector share a network namespace, so the application connects to `localhost` on a well-known port.
    No TLS or authentication is required between the application and the collector.
  - The sidecar handles secure forwarding upstream, keeping security concerns out of application code.
  - Custom collector images that encode organizational best practices (TLS, authentication, enrichment, sampling)
    can be built once and reused across all services, removing the burden from individual developers.
  - Kubernetes **admission webhooks** can automatically inject the sidecar into application deployments,
    so developers do not need to configure it explicitly.
  - Easiest to debug locally — developers can run the same sidecar image against a local observability stack to reproduce and troubleshoot issues.
  - Maximum isolation and flexibility — each pod can have its own collector configuration tailored to the application's signal types and formats.
    Best for polyglot environments or teams requiring per-service telemetry customization.

## Pipeline Architecture

Once telemetry passes the initial collection layer, it can be routed directly to storage backends or through intermediate aggregation collectors.
At scale, a two-tier approach is common:

  1. **Edge Collectors** (DaemonSet or sidecar) — lightweight, close to the application. Handle batching, local buffering, and basic enrichment.
  2. **Aggregator Collectors** — receive from many edge collectors. Perform additional enrichment, tail-based trace sampling, and fan-out to multiple backends.

Most modern storage backends (Loki, Mimir, Tempo) use a microservices architecture
where ingestion, storage, and querying components can be scaled independently to match workload characteristics.

## Alerting

In practice, especially in green-field setups, centralizing all recording and alerting rules in Grafana is recommended.
The Grafana ruler can query any configured datasource, cross-query multiple sources in a single rule, and is easier to manage in one place.
Grafana Alertmanager supports a wide range of notification destinations and its Slack integration supports the *Block Kit* API.

## Multi-Tenancy

Multi-tenancy allows a single observability stack to serve multiple environments or organizations while keeping their data isolated.
Each tenant operates within its own logical namespace: logs, metrics, and traces are stored and queried independently,
preventing one team from accidentally accessing or updating another team's data.

Multi-tenancy enables per-tenant access control, quotas, retention policies, and other configurations.
As usage grows, tenants can be scaled or migrated independently.

Loki, Mimir, and Tempo all support multi-tenancy natively through an `X-Scope-OrgID` header.
Grafana supports multi-tenancy through Organizations.
Each organization is a separate workspace within a single Grafana instance.

## Resources

  - **Fluent Bit**
    - [TLS](https://docs.fluentbit.io/manual/administration/transport-security)
    - [Backpressure](https://docs.fluentbit.io/manual/administration/backpressure)
  - **OpenTelemetry**
    - [Scaling the Collector](https://opentelemetry.io/docs/collector/scaling)
  - **Alloy**
    - [Deploy Grafana Alloy](https://grafana.com/docs/alloy/latest/set-up/deploy)
  - **Prometheus**
    - [Metric and label naming](https://prometheus.io/docs/practices/naming)
  - **Loki**
    - [Multi-Tenancy](https://grafana.com/docs/loki/latest/operations/multi-tenancy)
    - [Configuration Best Practices](https://grafana.com/docs/loki/latest/configure/bp-configure)
  - **Mimir**
    - [Deployment Modes](https://grafana.com/docs/mimir/latest/references/architecture/deployment-modes)
    - [Scaling Out](https://grafana.com/docs/mimir/latest/manage/run-production-environment/scaling-out)
  - **Tempo**
