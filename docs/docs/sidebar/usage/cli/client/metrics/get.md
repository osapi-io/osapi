# Metrics

Fetch Prometheus metrics from the API server:

```bash
$ osapi client metrics
# HELP http_server_request_duration_seconds Duration of HTTP server requests.
# TYPE http_server_request_duration_seconds histogram
http_server_request_duration_seconds_bucket{...} 0
...
```

The output is raw Prometheus exposition text, which can be piped to other tools
for processing.
