# compose

This repository contains *Docker Compose* files and service-specific configuration
for running local infrastructure stacks used in development and testing,
including observability and monitoring, databases, message queues, secret management, and security platforms.

## Tools

Install [Podman](https://podman.io):

```bash
# CLI
brew install podman
brew install podman-compose

# Desktop app (optional)
brew install --cask podman-desktop

# Start the Podman machine in rootful mode
podman machine init
podman machine set --rootful
podman machine start
```

## Resources

  - **Compose**
    - [Docker Compose](https://docs.docker.com/compose)
    - [Podman Compose](https://podman-desktop.io/docs/compose)
