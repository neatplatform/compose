# observability

The `compose.yml` file starts a local observability stack by provisioning and configuring observability services as containers.

Use it to quickly connect an application and verify its telemetry pipeline end to end, including logs, metrics, and traces.
This is especially useful for local development, integration testing, and troubleshooting instrumentation before deploying to higher environments.

## Quick Start

```bash
make up    # Start the observability stack
make down  # Stop the observability stack
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

### Log Aggregation

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
  - **Backends**
    - **Prometheus**
      - [Data model](https://prometheus.io/docs/concepts/data_model)
      - [Querying basics](https://prometheus.io/docs/prometheus/latest/querying/basics)
      - [Query functions](https://prometheus.io/docs/prometheus/latest/querying/functions)
      - [Metric and label naming](https://prometheus.io/docs/practices/naming)
      - [Management API](https://prometheus.io/docs/prometheus/latest/management_api)
    - **Alertmanager**
      - [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)
      - [High Availability](https://prometheus.io/docs/alerting/latest/high_availability)
      - [Management API](https://prometheus.io/docs/alerting/latest/management_api)
  - **Misc**
    - **Slack**
      - [Block Kit](https://docs.slack.dev/block-kit)
      - [Block Kit Builder](https://app.slack.com/block-kit-builder)
      - [`chat.postMessage` method](https://docs.slack.dev/reference/methods/chat.postMessage)
      - [Sending messages using incoming webhooks](https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks)
