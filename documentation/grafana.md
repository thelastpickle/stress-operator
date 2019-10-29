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
The `Grafana` CRD specifies a Grafana server to be provisioned.

The `Grafana` instance is namespace-based deployment with permissions limited to the namespace in which it is deployed. It is deployed in the same namespace as the tlp-stress operator, which defaults to `tlpstress`.

The `Grafana` instance looks like:

```
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: tlpstress
spec:
  service:
    labels:
      app: grafana
  config:
    log:
      mode: "console"
      level: "debug"
    security:
      admin_user: "root"
      admin_password: "secret"
    auth:
      disable_login_form: False
      disable_signout_menu: True
    auth.basic:
      enabled: False
    auth.anonymous:
      enabled: True
  dashboardLabelSelector:
    - matchExpressions:
        - {key: app, operator: In, values: [grafana]}
```

**Note:** The `Grafana` object currently is not configurable through the tlp-stress operator. You will need to directly edit [config/grafana/02_grafana.yaml](../config/grafana/02_grafana.yaml) at deployment time.

### Accessing the UI
The Grafana Operator deploys a service to expose the UI:

```
$ kubectl get svc grafana-service
NAME              TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
grafana-service   ClusterIP   10.80.3.122   <none>        3000/TCP   4h58m
```

By default this is a ClusterIP service, which can be accessed with  `kubectl port-forward`:

```
$ kubectl port-forward svc/grafana-service 3000:3000
```

## GrafanaDataSource
TODO

## GrafanaDashboard