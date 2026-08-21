---
name: update-container-images
description: >
  Update pinned container image tags in this repo's compose.yaml files and Dockerfiles to the latest version
  compatible with each image's current versioning scheme,
  using the source and registry URLs documented near each service or FROM line where available.
  Use when the user asks to bump, update, upgrade, refresh, or check Docker/container image versions
  in a compose file or Dockerfile, or asks "are our images up to date" for this repo's stacks.
---

# Update Docker Compose and Dockerfile image tags

Every service in this repo's compose files is documented with a comment block right above it, e.g.:

```yaml
# Loki – lightweight log storage
#
#   https://github.com/grafana/loki
#   https://hub.docker.com/r/grafana/loki
#
loki:
  image: grafana/loki:3.7.6
```

That comment block is the intended source of truth for "what is this image and where do I check its versions".
Use it instead of guessing at a registry from the image string alone,
since the same org can publish under different names in different registries.

Dockerfiles pin base images the same way, via `FROM` lines,
sometimes with a registry URL comment right above (as in `compose/observability/volt/Dockerfile`),
sometimes with no per-line comment at all (as in `compose/observability/slack/Dockerfile`).
The same update, same care about versioning scheme, and same report apply to both file types.
Treat every `FROM <image>:<tag>` line as equivalent to a compose service's `image:` line,
with the differences noted below.

## 1. Find the target(s)

  - If invoked with an argument that looks like a path or glob, scope to that file (or files) only.
  - Otherwise, find every `compose.yaml`/`docker-compose.yml`/`docker-compose.yaml` **and** every `Dockerfile`
    (including suffixed variants like `Dockerfile.builder`) in the current repo,
    and confirm the list with the user before touching more than one file.
  - If an argument contains `--dry-run`, do the full analysis and report but skip step 4 (no file edits).

## 2. Parse each service / FROM line

For every top-level entry under `services:` in a compose file, and every `FROM` instruction in a Dockerfile:

  - In compose files, skip entries with no `image:` key
    (e.g. a `build:` service like `alert-receiver`) — there's nothing to bump.
  - In Dockerfiles, skip a `FROM` line whose image name matches an earlier stage's `AS <name>`
    alias in the same file (e.g. a final `FROM builder AS final` or `COPY --from=builder`) — 
    that references a build stage, not a registry image.
    Multi-stage builds otherwise get every real `FROM` line checked independently, since stages
    commonly pin unrelated images (e.g. `golang` for the builder stage, `alpine` for the final stage).
  - Read the comment block immediately above the entry, up to the previous blank/non-comment line. Pull out:
    - The **source repo** URL (`github.com/<owner>/<repo>`) — used for release notes / changelog context.
    - The **registry** URL, one of `hub.docker.com/r/<ns>/<repo>`, `quay.io/repository/<ns>/<repo>`,
      or a GitHub Packages link (`github.com/<owner>/<repo>/pkgs/container/<repo>`, i.e. `ghcr.io/<owner>/<repo>`).
      Some services (e.g. `fluent-bit`) only document the source repo — that's fine, fall back to it.
    - If there's no per-line comment at all
      (common in Dockerfiles, which often only have generic stage markers like `# BUILD STAGE`)
      and the image name has no namespace (e.g. `golang`, `alpine`, `node`, `python`),
      treat it as a Docker Official Image: registry is `hub.docker.com/_/<repo>`, no source repo to fall back on.
      That's enough to look up tags in step 3 — don't skip the image just because it's undocumented.
  - Parse the current tag from the `image:` line or `FROM` line and infer its versioning scheme:
    the prefix (`v` or none) and the number of numeric segments (almost everything in this repo is `v?MAJOR.MINOR.PATCH`).
    An image that has always been tagged `v1.2.3` should stay `vX.Y.Z`, not jump to a `1.2.3-alpine` or `nightly` variant.
    A Dockerfile base image tagged `1.26.5-alpine` should stay on the `-alpine` variant, not jump to a bare tag.

## 3. Find the latest matching tag

Prefer querying the registry directly over scraping the human-facing web page.
Registry pages are JS-rendered and easy to misread, while the underlying APIs return exact, machine-readable tag lists.
The `image:` field already tells you the registry host and repo path, so derive the API call from that
(the comment URLs are your cross-check and your source for release notes, not a hard requirement to fetch as HTML).

Use `curl` via Bash. Registry-specific list-tags calls:

  - **Docker Hub** (`docker.io`, e.g. `grafana/loki`, `otel/opentelemetry-collector-contrib`):

    ```
    curl -s "https://hub.docker.com/v2/repositories/<ns>/<repo>/tags?page_size=100" | jq -r '.results[].name'
    ```
    Paginate via the `next` field in the response if you need more than 100.
    For a Docker Official Image (unnamespaced, e.g. `golang`, `alpine`), use `library` as `<ns>`.

  - **Quay.io** (e.g. `quay.io/prometheus/node-exporter`):

    ```
    curl -s "https://quay.io/api/v1/repository/<ns>/<repo>/tag/?limit=100&onlyActiveTags=true" | jq -r '.tags[].name'
    ```

  - **GHCR / GitHub Packages** (e.g. `ghcr.io/google/cadvisor`):
    GHCR's tag-list API needs auth even for public images, so instead use the OCI Distribution v2 API anonymously:

    ```
    TOKEN=$(curl -s "https://ghcr.io/token?scope=repository:<owner>/<repo>:pull" | jq -r .token)
    curl -s -H "Authorization: Bearer $TOKEN" "https://ghcr.io/v2/<owner>/<repo>/tags/list" | jq -r '.tags[]'
    ```

  - **Anything else with a documented registry** (e.g. `cr.fluentbit.io/fluent/fluent-bit`):
    try the same anonymous OCI v2 flow against that host before falling back to GitHub tags:

    ```
    TOKEN=$(curl -s "https://<registry-host>/token?scope=repository:<ns>/<repo>:pull" | jq -r .token)
    curl -s -H "Authorization: Bearer $TOKEN" "https://<registry-host>/v2/<ns>/<repo>/tags/list" | jq -r '.tags[]'
    ```

  - **Fallback for any service** (registry API unreachable, unauthenticated, or undocumented):
    use the GitHub repo's tags as a proxy, since these projects publish images matching their git tags:

    ```
    curl -s "https://api.github.com/repos/<owner>/<repo>/tags?per_page=100" | jq -r '.[].name'
    ```

    If `GITHUB_TOKEN` is set in the environment,
    pass it as `-H "Authorization: Bearer $GITHUB_TOKEN"` to avoid the unauthenticated rate limit.

Once you have a raw tag list:

  1. Filter to tags matching the current tag's shape from step 2 (same prefix, same number of numeric segments,
     no extra suffix like `-rc`, `-beta`, `-alpine`, no `latest`, `main`, `nightly`, date-stamped tags)
     Unless the current tag already uses that kind of suffix, in which case match it.
  2. Sort the survivors as semantic versions (numeric major.minor.patch comparison, not string sort) and take the highest.
  3. If nothing survives the filter, don't guess — leave that service's image untouched
     and note it in the report as needing manual review (the upstream project may have changed its tagging scheme).

## 4. Apply the change

Edit only the tag in the `image:` line or `FROM` line, in place, preserving indentation,
the rest of the file, and (in Dockerfiles) any trailing `AS <stage>` untouched.
Don't touch `build:` services, don't reorder services or stages, don't "clean up" unrelated lines.

If the new major version differs from the current major version, still apply it (that's genuinely the latest),
but flag it clearly in the report below — a major bump is more likely to need config changes
and is worth the user reading the release notes for.

## 5. Report

Finish with a table like:

| File | Service / stage | Old tag | New tag | Notes |
|----|----|----|----|----|
| observability/compose.yaml | loki | 3.7.6 | 3.8.0 | |
| observability/compose.yaml | mimir | 3.1.4 | 3.1.4 | already latest |
| observability/compose.yaml | grafana | 13.1.3 | 14.0.0 | **major bump** — check release notes |
| observability/compose.yaml | fluent-bit | 5.1.0 | 5.1.0 | couldn't confirm — registry API unreachable, left unchanged |
| observability/volt/Dockerfile | builder (golang) | 1.26.2 | 1.26.5 | |
| observability/volt/Dockerfile | final (alpine) | 3.23.4 | 3.24.1 | |

Then run `git diff` in the affected repo so the changes are visible, and stop.
Don't commit, push, or open a PR unless the user explicitly asks for that next.
