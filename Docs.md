# Long Ping Documentation

## Overview

Long Ping is a network observability tool that monitors long-term network performance by sending one ICMP ping packet per second to each configured host. Rather than sending large bursts of packets, it maintains rolling windows over the last **15**, **100**, and **1000** packets to calculate statistics. This allows detection of subtle, long-term degradation that periodic burst-based tools would miss.

## How It Works

- One ping packet is sent per second per monitored IP address.
- DNS resolution is performed at startup; each resolved IP address is monitored independently.
- Rolling ring buffers store the results of the last 15, 100, and 1000 packets per IP.
- Statistics are computed continuously from those buffers.
- Metrics are exposed via a Prometheus-compatible HTTP endpoint and optionally written to InfluxDB on a configurable interval.

## HTTP Endpoints

The application listens on **port 3000**.

| Endpoint    | Method | Description |
|-------------|--------|-------------|
| `/healthz`  | GET    | Liveness probe. Always returns `200 OK`. |
| `/readyz`   | GET    | Readiness probe. Returns `503` if InfluxDB is enabled but not yet connected; otherwise `200 OK`. |
| `/metrics`  | GET    | Prometheus metrics endpoint. |

## Configuration

All configuration is done through **environment variables**. There are no configuration files.

### Core Settings

| Variable        | Required | Default | Description |
|-----------------|----------|---------|-------------|
| `HOSTS`         | **Yes**  | —       | Space-separated list of hostnames or IP addresses to monitor. Example: `"8.8.8.8 github.com 192.168.1.1"` |
| `LOG_LEVEL`     | No       | `info`  | Logging verbosity. Accepted values: `debug`, `info`, `warn`, `error`. |
| `INFLUX_ENABLED`| No       | `false` | Set to `true` to enable writing metrics to InfluxDB. |
| `PROBE_INTERVAL`| No       | `15`    | Interval in seconds at which rolling statistics are written to InfluxDB. Has no effect if `INFLUX_ENABLED` is not `true`. |
| `PROBE_TIMEOUT` | No       | `1`     | Timeout in seconds for each individual ping probe. Packets that do not receive a reply within this window are counted as lost. |

### InfluxDB Settings

These variables are **required** when `INFLUX_ENABLED=true`. The application will exit at startup if any are missing.

| Variable        | Required | Default | Description |
|-----------------|----------|---------|-------------|
| `INFLUX_SERVER` | **Yes**  | —       | Full URL of the InfluxDB server, including protocol and port. Example: `http://influxdb.example.com:8086/` |
| `INFLUX_TOKEN`  | **Yes**  | —       | InfluxDB authentication token. |
| `INFLUX_BUCKET` | **Yes**  | —       | InfluxDB bucket to write metrics into. |
| `INFLUX_ORG`    | **Yes**  | —       | InfluxDB organization name. |
| `DB_MAX_ERROR`  | **Yes**  | `10`    | Maximum number of consecutive InfluxDB write errors before the application exits. |

## Metrics

### Prometheus

All metrics are labeled with `hostname` and `ip_address` and are available at `GET /metrics`.

| Metric Name              | Description |
|--------------------------|-------------|
| `total_sent`             | Total ICMP packets sent to the host. |
| `total_received`         | Total ICMP packets received from the host. |
| `total_loss`             | Total ICMP packets lost. |
| `total_duplicates`       | Total duplicate ICMP packets received. |
| `packetloss_15`          | Packet loss ratio over the last 15 packets (0.0–1.0). |
| `packetloss_100`         | Packet loss ratio over the last 100 packets (0.0–1.0). |
| `packetloss_1000`        | Packet loss ratio over the last 1000 packets (0.0–1.0). |
| `avg_15_latency_ns`      | Average round-trip latency (ns) over the last 15 packets. |
| `avg_100_latency_ns`     | Average round-trip latency (ns) over the last 100 packets. |
| `avg_1000_latency_ns`    | Average round-trip latency (ns) over the last 1000 packets. |
| `min_15_latency_ns`      | Minimum round-trip latency (ns) over the last 15 packets. |
| `min_100_latency_ns`     | Minimum round-trip latency (ns) over the last 100 packets. |
| `min_1000_latency_ns`    | Minimum round-trip latency (ns) over the last 1000 packets. |
| `max_15_latency_ns`      | Maximum round-trip latency (ns) over the last 15 packets. |
| `max_100_latency_ns`     | Maximum round-trip latency (ns) over the last 100 packets. |
| `max_1000_latency_ns`    | Maximum round-trip latency (ns) over the last 1000 packets. |
| `jitter_15_ns`           | Jitter (ns) over the last 15 packets. |
| `jitter_100_ns`          | Jitter (ns) over the last 100 packets. |
| `jitter_1000_ns`         | Jitter (ns) over the last 1000 packets. |

### InfluxDB

When InfluxDB is enabled, the same statistics listed above are written to the configured bucket under the measurement name `longping`. Each data point is tagged with `Host` (hostname) and `Ip` (resolved IP address). Writes occur on the interval defined by `PROBE_INTERVAL` (default: every 15 seconds).

## Example Configuration

### Bare Metal / Local

```bash
export HOSTS="8.8.8.8 github.com 192.168.1.1"
export LOG_LEVEL="info"
export PROBE_TIMEOUT="1"
export PROBE_INTERVAL="15"
export INFLUX_ENABLED="false"
```

### With InfluxDB Enabled

```bash
export HOSTS="8.8.8.8 github.com"
export LOG_LEVEL="info"
export PROBE_TIMEOUT="1"
export PROBE_INTERVAL="15"
export INFLUX_ENABLED="true"
export INFLUX_SERVER="http://influxdb.example.com:8086/"
export INFLUX_TOKEN="your-influxdb-token"
export INFLUX_BUCKET="network-monitoring"
export INFLUX_ORG="your-org"
export DB_MAX_ERROR="10"
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: longping-config
  namespace: apps
data:
  HOSTS: "8.8.8.8 github.com"
  LOG_LEVEL: "info"
  INFLUX_ENABLED: "false"
  PROBE_INTERVAL: "15"
  PROBE_TIMEOUT: "1"
```

## Deployment

A `Dockerfile` is included in the repository for building a container image. Kubernetes manifests are available in the [`kubernetes/`](kubernetes/) directory, including a `ConfigMap` and `Deployment`.

### Running with Docker

```bash
docker build -t longping .
docker run -e HOSTS="8.8.8.8 github.com" longping
```

> **Note:** ICMP ping requires elevated privileges. When running in a container, the container may need `NET_RAW` capability or to run with `--privileged`.

## Signals

Long Ping handles `SIGINT` and `SIGTERM` for graceful shutdown. When either signal is received, the application closes the InfluxDB connection (if enabled) and exits cleanly.
