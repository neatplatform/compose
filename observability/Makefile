SHELL := /bin/bash


.PHONY: prep
prep:
	@envsubst < alertmanager/config.yaml.tmpl > alertmanager/config.yaml

.PHONY: up
up: prep
	@podman compose -f compose.yaml up -d

.PHONY: list
list:
	@podman compose -f compose.yaml ps --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}"

.PHONY: down
down:
	@podman compose -f compose.yaml down

.PHONY: clean
clean:
	@podman volume ls -q --filter "label=io.podman.compose.project=observability" | xargs podman volume rm -f
	@rm -rf ./logs
	@rm alertmanager/config.yaml

.PHONY: logs
logs:
	@containers=$$(podman compose -f compose.yaml ps --format "{{.Names}}"); \
	 printf "🐳 Containers:\n\n"; \
	 PS3=$$'\n👉 Select a container by number: '; \
	 select container in $$containers; do \
	   if [ -n "$$container" ]; then \
	     echo; \
	     podman container logs --follow $$container; \
	     break; \
	   else \
	     printf "❌ Invalid selection. Try again.\n"; \
	   fi; \
	 done

.PHONY: loki-rules
loki-rules:
	@curl -s http://localhost:5100/loki/api/v1/rules | yq

.PHONY: mimir-rules
mimir-rules:
	@curl -s http://localhost:5200/prometheus/api/v1/rules | jq

.PHONY: grafana-rules
grafana-rules:
	@curl -s http://localhost:7200/api/v1/provisioning/alert-rules/export?format=json | jq

.PHONY: grafana-contactpoints
grafana-contactpoints:
	@curl -s http://localhost:7200/api/v1/provisioning/contact-points/export?format=json | jq

.PHONY: grafana-policies
grafana-policies:
	@curl -s http://localhost:7200/api/v1/provisioning/policies/export?format=json | jq
