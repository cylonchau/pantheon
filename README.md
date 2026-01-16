# Pantheon

<p>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://prometheus.io"><img src="https://img.shields.io/badge/Prometheus-Compatible-E6522C?style=flat-square&logo=prometheus" alt="Prometheus"></a>
  <a href="https://victoriametrics.com"><img src="https://img.shields.io/badge/VictoriaMetrics-Compatible-621773?style=flat-square" alt="VictoriaMetrics"></a>
  <a href="https://grafana.com"><img src="https://img.shields.io/badge/Grafana-Compatible-F46800?style=flat-square&logo=grafana" alt="Grafana"></a>
  <a href="#"><img src="https://img.shields.io/badge/Helm-Chart-0F1689?style=flat-square&logo=helm" alt="Helm"></a>
</p>

Pantheon - centralized management of scrape targets for Prometheus/VictoriaMetrics via HTTP SD.

Pantheonctl - universal exporter CLI tool.

## Features

- Prometheus/VictoriaMetrics targets discovery via HTTP SD.
- Multi Prometheus/VictoriaMetrics quick switch.
- Target Management.
- Proxy Mode (if exporter access with authentication).
- Custom labels management.
- Pushgateway integration.
- Exporter integration.
- Blackbox compatible.
- Monitoring as Code.



## Pantheonctl - Universal Exporter CLI

A lightweight CLI tool that bundles multiple exporters into a single binary.

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

# Push custom metrics
pantheonctl push --name=my_counter --type=counter --value=42 --labels="job=etl"
```


## Contribute
If you have any idea for an improvement or find a bug do not hesitate in opening an issue, just simply fork and create a pull-request to help improve the exporter.

## License

Apache License 2.0
