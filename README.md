# Pantheon

<p>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://prometheus.io"><img src="https://img.shields.io/badge/Prometheus-Compatible-E6522C?style=flat-square&logo=prometheus" alt="Prometheus"></a>
  <a href="https://victoriametrics.com"><img src="https://img.shields.io/badge/VictoriaMetrics-Compatible-621773?style=flat-square" alt="VictoriaMetrics"></a>
  <a href="https://grafana.com"><img src="https://img.shields.io/badge/Grafana-Compatible-F46800?style=flat-square&logo=grafana" alt="Grafana"></a>
  <a href="#"><img src="https://img.shields.io/badge/Helm-Chart-0F1689?style=flat-square&logo=helm" alt="Helm"></a>
</p>

Pantheon is a Prometheus observability governance platform. It provides a unified control plane for managing scrape targets across both bare-metal and Kubernetes environments — with automatic target lifecycle management via a Kubernetes controller, declarative monitor rules, proxy support for authenticated exporters, and a universal exporter CLI.

## Architecture

Pantheon consists of three components:

| Component | Description |
|-----------|-------------|
| **pantheon-server** | Central API server. Manages scrape targets, selectors, and monitor rules. Exposes discovery endpoints compatible with Prometheus and VictoriaMetrics. |
| **pantheon-controller** | Kubernetes controller (operator). Watches Pod and Service objects and automatically manages the scrape target lifecycle on the server based on `MonitorRule` configuration. |
| **pantheonctl** | Universal CLI tool. Bundles multiple exporters and supports pushing custom metrics to a Pushgateway. |

## Quick Start

### 1. Start pantheon-server

Edit `config.toml` (SQLite by default, no extra setup needed):

```toml
appname = "pantheon-server"
port    = 8899
address = "0.0.0.0"
database_driver = "sqlite"

[sqlite]
file     = "pantheon.db"
database = "cmp"
```

Run the server:

```bash
./pantheon-server --config config.toml
# API available at http://localhost:8899
# Swagger UI at  http://localhost:8899/doc
```

### 2. Configure pantheonctl

Initialize the local config file and register the server:

```bash
# Create a blank config file
pantheonctl config init

# Register pantheon-server as a cluster
pantheonctl config add-cluster --name prod --server http://localhost:8899

# Set the active context
pantheonctl config set-context prod
```

### 3. Register a scrape target (manual)

```bash
# Add a node-exporter target with labels and a selector
pantheonctl target add \
  --address 10.0.0.1:9100 \
  --labels env=prod,team=infra \
  --selector job=node-exporter

# Add a blackbox target with params
pantheonctl target add \
  --address localhost:9115 \
  --labels env=prod \
  --params target=google.com,module=http_2xx \
  --selector job=blackbox

# Add a target and drop noisy metrics at proxy level
pantheonctl target add \
  --address 10.0.0.2:9090 \
  --drop-metrics "go_.*;process_.*" \
  --selector job=prometheus
```

### 4. Configure Prometheus / VictoriaMetrics

Point Prometheus at Pantheon's discovery endpoint using the selector `job=node-exporter`:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: node-exporter
    http_sd_configs:
      - url: http://localhost:8899/ph/v1/targets/selector/job/node-exporter
        refresh_interval: 30s
```

### 5. Use the proxy endpoint (optional, for authenticated exporters)

If your exporter requires credentials, configure Prometheus to scrape via the Pantheon proxy:

```yaml
scrape_configs:
  - job_name: authenticated-exporter
    http_sd_configs:
      - url: http://localhost:8899/ph/v1/targets/selector/job/my-exporter
    metrics_path: /ph/v1/proxy
    params:
      schema: [http]
      host:   [10.0.0.1]
      port:   [9100]
```

### 6. Kubernetes auto-discovery (controller)

Deploy the controller alongside the server, then create a monitor rule via CLI:

```bash
# Scrape all pods with app=my-app in the production namespace
pantheonctl monitor add \
  --name my-app-pods \
  --type pod \
  --namespace production \
  --selector app=my-app \
  --port metrics \
  --metric-path /metrics \
  --labels team=backend,env=prod \
  --drop-metrics "go_gc.*"
```

The controller will watch Kubernetes for matching pods and automatically register/deregister scrape targets. No further intervention needed.

## Features

- **Kubernetes Auto-Discovery** — controller watches Pods and Services, automatically managing the full target lifecycle (register/update/deregister) based on label selectors and monitor rules.
- **Monitor Rules** — declarative rules that define which Kubernetes resources (pods/services) to scrape, on which port, with which labels, and which metrics to drop.
- **Target Management** — CRUD API for managing scrape targets manually or programmatically, usable without Kubernetes.
- **Proxy Mode** — reverse proxy for exporters that require authentication, forwarding scrape requests transparently with credentials.
- **Selector Management** — group and filter targets using named selectors for fine-grained Prometheus job configuration.
- **Custom Labels** — inject extra labels into targets for enriched metric metadata.
- **Pushgateway Integration** — push custom or batch job metrics via `pantheonctl push`.
- **Multi-backend Support** — compatible with both Prometheus and VictoriaMetrics.
- **Blackbox Compatible** — supports blackbox exporter style targets.
- **Monitoring as Code** — manage all scrape configuration via API or Helm values.

## Pantheon Controller — Kubernetes Auto-Discovery

The controller is a Kubernetes operator that watches `Pod` and `Service` objects. When a resource matches a `MonitorRule` (by namespace and label selector), the controller automatically registers a scrape target on the pantheon-server. When the resource is deleted or no longer matches, the target is removed.

### Monitor Rule Example

```yaml
name: my-app-pods
type: pod                     # "pod" or "service"
namespace: production         # "*" matches all namespaces
selector: app=my-app          # label selector (comma-separated key=value)
port_name: metrics            # port name or number on the container
metric_path: /metrics
labels: team=backend,env=prod # extra labels injected into the target
drop_metrics: go_gc.*         # regex to drop unwanted metrics at proxy level
```

## Pantheonctl — Universal Exporter CLI

A lightweight CLI tool that bundles multiple exporters into a single binary and supports pushing custom metrics.

### Supported Exporters

- node-exporter
- redis-exporter
- mongodb-exporter
- nginx-exporter

### Usage

```bash
# Collect node metrics and push to Pushgateway
pantheonctl collect --type=node --push=http://pushgateway:9091

# Collect Redis metrics
pantheonctl collect --type=redis --redis-addr=redis://localhost:6379 --push=http://pushgateway:9091

# Push custom metrics to a Pushgateway
pantheonctl push my_job --address http://pushgateway:9091 --metric cpu_usage=0.8 --metric request_count=100:counter
```

## Development

### Building Binaries

```bash
# Build all modules
make all

# Build a specific module
make build module=pantheon-server
make build module=pantheon-controller
make build module=pantheonctl
```

### Testing

```bash
# Run all unit tests with coverage summary
make test

# Generate HTML coverage report (coverage.html)
make cover

# Run linter
make lint
```

### Swagger Documentation

```bash
make swagger
```

### Building Docker Images

```bash
# Build the Pantheon Server image
docker build --target server -t cylonchau/pantheon-server:latest .

# Build the Pantheon Controller image
docker build --target controller -t cylonchau/pantheon-controller:latest .
```

## Deployment

### Deploying via Helm

#### Prerequisites

* A Kubernetes cluster
* Helm 3.x installed

#### Deploy Server + Controller together

```bash
helm install pantheon ./helm --namespace monitoring --create-namespace
```

#### Deploy Server only

```bash
helm install pantheon-server ./helm \
  --namespace monitoring \
  --create-namespace \
  --set server.enabled=true \
  --set k8s.version=1.34
```

#### Deploy Controller only

```bash
helm install pantheon-controller ./helm \
  --namespace monitoring \
  --create-namespace \
  --set controller.enabled=true \
  --set server.enabled=false \
  --set k8s.version=1.34 \
  --set nodeSelector.enabled=false \
  --set controller.serverURL=10.0.0.1:8899 \
  --set controller.image.tag=latest
```

## Contribute

If you have any idea for an improvement or find a bug, do not hesitate to open an issue, fork the repository, and create a pull request.

## License

Apache License 2.0