# Grafana Integration
The tlp-stress operator provides integration with Grafana through the [Grafana Operator](https://github.com/integr8ly/grafana-operator) in the following ways:

* Deploy a Grafana server
* Deploy a data source for Prometheus
* Deploy a dashboard for each `TLPStress` instance

## Grafana Operator
The tlp-stress operator itself does not yet provision the Grafana server or data source. It is currently done when deploying the operator:

```
$ make deploy-all
```

This will first deploy the [Grafana Operator](https://github.com/integr8ly/grafana-operator) into the same namespace in which the tlps-stress operator is deploy. This is `tlpstress` by default.

The Grafana Operator adds serveral new CRDs including:

* `Grafana`
* `GrafanaDataSource`
* `GrafanaDashboard`

## Grafana
TODO

## GrafanaDataSource
TODO

## GrafanaDashboard