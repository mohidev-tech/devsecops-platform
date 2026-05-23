# Grafana dashboards

`slo-api.json` — request rate, queue depth, drain rate, SLO compliance, error-budget burn.

## Import

The Grafana sidecar (configured in `observability/prometheus/values.yaml`) auto-imports any ConfigMap labeled `grafana_dashboard=1`. Quick load:

```bash
kubectl -n monitoring create configmap api-slo-dashboard \
  --from-file=slo-api.json=observability/grafana/dashboards/slo-api.json
kubectl -n monitoring label configmap api-slo-dashboard grafana_dashboard=1
```

## Open Grafana

```bash
kubectl -n monitoring port-forward svc/kps-grafana 3000:80
# http://localhost:3000  — admin / admin
```

## What's stubbed
The SLO compliance and burn-rate panels are placeholders until `api_requests_total` carries a `status` label (planned Phase 3, when api adds structured response codes). The dashboard layout, alerts wiring (via the chart's PrometheusRule), and ServiceMonitor flow are real.
