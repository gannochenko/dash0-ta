# Dash0 Sergei Gannochenko test assignment

## Task

Design a high-throughput system that analyzes the incoming logs and dumps statistics to STDOUT every N seconds.

## Algorithm

From the descriptions I assume, that two log entries are different if their messages are different.
Thus, in order to calculate the number of different log entries for every distinct attribute value, we keep
a nested map of value => hash(message) => boolean and calculate the amount of keys when reporting to stdout.

Since the attributes present in all 3 sections, the value precedence takes place: the log-based attribute overrides the scope based and so on.

## Architecture

Since the high througput is requierd, any locking or critical sections is not an option. Instead, let's heavily rely on channels and workers. So:

1. Inside the controllers I avoid doing any computation, instead I immediately send the request to the channel to avoid traffic congestion. The write operation is non-blocking.
2. A worker pool (of adjustable size) grinds down jobs coming from the _jobs_ channel, finding `value => message` pairs for a chosen attribute. This is a heavy duty, since the complexity is O(N\*M\*P\*K) (4 nested cycles), so I can increase the amount of workers should the traffic increase. The found values are being sent to _result_ channel.
3. The _jobs_ channel is buffered, because the write operation is blocking with a 1 second timeout. This is required, otherwise if all workers are busy, write will be unsuccessful and I loose the data for the hit. If this becomes a choking point, either increase the amount of workers, or the buffer size.
4. Another worker reconciles the incoming results into a common data structure, dumps the content of the structure at given interval and cleans the structure up.
5. No critical section is required, because all operations with data structures are sequential.

## Components

The solution is a cloud native application suitable for running in production. Here are main components:

1. gRPC server exposing the opentelemetry.proto.collector.logs.v1.LogsService/Export method.
2. Proto definitions compiled into Golang are taken from the Otel package.
3. Code structure follows the Hexagonal architecture pattern: there domain logic is decoupled into services, and the services are then exposed via "adapters" or "controllers". Thus, gRPC networking is decoupled from the workers.
4. The code is organized using the dependency injection pattern, the dependencies are passed via interfaces to let the code be testable, orthogonal and isolated.
5. HTTP server exposes the _/metrics_ and _/health_ endpoints to be used by Prometheus and a container-orchestrator-of-your-choice respectively.
6. The application contains a shutdown sequence to support graceful shutdown.
7. Structured logging is used to produce some debug output.
8. Error handler and logger middlewares for the gRPC server are implemented.
9. Basic Otel is enabled (borrowed from the starter code) to measure CPU/Memory utilization, Goroutine amount.
10. The config is read from the `.env.local` file, and becomes available in the code through a dedicated Config service.

## How to run

Copy `.env.example` to `.env.local`, adjust the values when needed.

In order to run, just type

```sh
make run
```

It will run three containers: the application itself, [Prometheus](http://localhost:9090), [Grafana](http://localhost:3000/dashboards).

To run a 5m loadtest on the solution, type

```sh
make run_loadtest_long
```

You need to have [k6](https://k6.io/) installed.

On my M1 mac it gave me about 200 requests per second with average latency of 40ms.

You can go to Grafana, create a dashboard (see Appendix A) to track the operational metrics.

## Further improvements

Since I am running out of time, the following was left out of scope:

1. Additional operational metrics, such as amount of requests per second and error rate may be added.
2. Brush the code up, move some code to helpers.
3. Enable max message size and insecure credentials (be careful with that in prod) on the gRPC server.
4. Add factory methods for dependency creation.
5. Write unit tests, especially for the part that parses incoming log entries.
6. Write end-to-end tests to make sure the application functions correctly.
7. Integration tests may be also written.

## Links

- [Link to the task](https://github.com/dash0hq/take-home-assignments/tree/main/otlp-log-processor-backend-go)
- [Log structure example](https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/logs.json)

## Appendix A. Grafana dashboard

```json
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "thresholds"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "yellow",
                "value": 70
              },
              {
                "color": "red",
                "value": 90
              }
            ]
          },
          "unit": "percent",
          "min": 0,
          "max": 100
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "id": 1,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "values": false,
          "calcs": ["lastNotNull"],
          "fields": ""
        },
        "textMode": "auto"
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "rate(process_cpu_seconds_total[5m]) * 100",
          "legendFormat": "CPU Usage %",
          "refId": "A"
        }
      ],
      "title": "CPU Usage",
      "type": "stat"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "thresholds"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "yellow",
                "value": 1000000000
              },
              {
                "color": "red",
                "value": 2000000000
              }
            ]
          },
          "unit": "bytes",
          "decimals": 2
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 0
      },
      "id": 2,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "values": false,
          "calcs": ["lastNotNull"],
          "fields": ""
        },
        "textMode": "auto"
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "process_resident_memory_bytes",
          "legendFormat": "RSS Memory",
          "refId": "A"
        }
      ],
      "title": "Memory Usage",
      "type": "stat"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "vis": false
            },
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "never",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "percent",
          "min": 0,
          "max": 100
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 24,
        "x": 0,
        "y": 8
      },
      "id": 3,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom"
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "rate(process_cpu_seconds_total[5m]) * 100",
          "legendFormat": "CPU Usage %",
          "refId": "A"
        }
      ],
      "title": "CPU Usage Over Time",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "vis": false
            },
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "never",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "bytes"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 24,
        "x": 0,
        "y": 16
      },
      "id": 4,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom"
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "process_resident_memory_bytes",
          "legendFormat": "Resident Memory",
          "refId": "A"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "process_virtual_memory_bytes",
          "legendFormat": "Virtual Memory",
          "refId": "B"
        }
      ],
      "title": "Memory Usage Over Time",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "vis": false
            },
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "never",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "bytes"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 24
      },
      "id": 5,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom"
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_memstats_heap_alloc_bytes",
          "legendFormat": "Heap Allocated",
          "refId": "A"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_memstats_heap_inuse_bytes",
          "legendFormat": "Heap In Use",
          "refId": "B"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_memstats_heap_idle_bytes",
          "legendFormat": "Heap Idle",
          "refId": "C"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_memstats_stack_inuse_bytes",
          "legendFormat": "Stack In Use",
          "refId": "D"
        }
      ],
      "title": "Go Memory Statistics",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "vis": false
            },
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "never",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "short"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 24
      },
      "id": 6,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom"
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "pluginVersion": "9.0.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_goroutines",
          "legendFormat": "Goroutines",
          "refId": "A"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_threads",
          "legendFormat": "OS Threads",
          "refId": "B"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "expr": "go_sched_gomaxprocs_threads",
          "legendFormat": "GOMAXPROCS",
          "refId": "C"
        }
      ],
      "title": "Goroutines and Threads",
      "type": "timeseries"
    }
  ],
  "refresh": "5s",
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["go", "performance", "monitoring"],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "",
  "title": "Go Application - CPU and Memory Utilization",
  "uid": "go-cpu-memory-dashboard",
  "version": 1,
  "weekStart": ""
}
```
