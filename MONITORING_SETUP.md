# Monitoring Setup with Docker Compose

This setup provides Prometheus and Grafana monitoring for the log-processor application.

## What's included

1. **Log Processor Application** - Go application with Prometheus metrics exposed
2. **Prometheus** - Scrapes metrics every second from the `/metrics` endpoint
3. **Grafana** - Visualizes metrics with pre-configured dashboards

## Changes Made

### Application Changes

- Added HTTP server alongside gRPC server
- Configured OpenTelemetry with Prometheus exporter
- Exposed `/metrics` endpoint on port 8080
- Added `/health` endpoint for health checks

### Configuration Files

- `docker-compose.yml` - Main orchestration file
- `prometheus/prometheus.yml` - Prometheus configuration with 1-second scraping
- `grafana/provisioning/` - Auto-configured datasource and dashboards
- `apps/log-processor/Dockerfile` - Multi-stage build for the Go application

## How to Run

1. **Start the stack:**

   ```bash
   docker-compose up --build
   ```

2. **Access the services:**

   - **Prometheus**: http://localhost:9090
   - **Grafana**: http://localhost:3000 (admin/admin)
   - **Metrics endpoint**: http://localhost:8080/metrics
   - **Health check**: http://localhost:8080/health

3. **View metrics:**
   - Go to Grafana dashboard "Log Processor Metrics"
   - Check Prometheus targets at http://localhost:9090/targets

## Environment Variables

The application supports these environment variables:

- `LOG_PROCESSOR_HTTP_ADDR` - HTTP server address (default: `:8080`)
- `LOG_PROCESSOR_GRPC_ADDR` - gRPC server address (default: `:443`)
- `LOG_PROCESSOR_WORKER_COUNT` - Number of workers (default: `5`)
- `LOG_PROCESSOR_WINDOW_SIZE` - Window size in seconds (default: `1000`)

## Prometheus Metrics

The application exposes standard Go runtime metrics and OpenTelemetry metrics:

- `go_goroutines` - Number of goroutines
- `go_memstats_alloc_bytes` - Memory allocated
- `promhttp_metric_handler_requests_total` - HTTP metrics requests
- Custom OpenTelemetry metrics (when added to your application)

## Scraping Frequency

Prometheus is configured to scrape metrics every 1 second as requested:

```yaml
global:
  scrape_interval: 1s
scrape_configs:
  - job_name: "log-processor"
    scrape_interval: 1s
```

## Grafana Dashboard

The included dashboard shows:

- HTTP request rates
- Go runtime metrics (memory, goroutines)
- Refresh rate: 1 second

## Stopping the Stack

```bash
docker-compose down
```

To remove volumes as well:

```bash
docker-compose down -v
```
