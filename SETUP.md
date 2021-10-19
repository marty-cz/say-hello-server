## Grafana 

Grafana default login `admin:prom-operator`


sum by (code) (rate(app_http_request_total{app="say-hello-server"}[1m]))

## Loki

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install loki grafana/loki-stack   --set fluentd.enabled=true,promtail.enabled=false
```

### Setup DataSource in Grafana  

https://grafana.com/docs/grafana/v7.5/datasources/loki/

URL of Loki
http://loki.default.svc.cluster.local:3100

## Logging

### Output and Flow

Use Cluster Output and Flow for simplicity

URL of Loki
http://loki.default.svc.cluster.local:3100

## Helm

```shell
helm install say-hello-server helm-chart/ --values helm-chart/values.yaml

helm uninstall say-hello-server
```