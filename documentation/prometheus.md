# Prometheus Integration
The tlp-stress operator provides Prometheus integration in a couple of ways. First, it exposes metrics for each `TLPStress` instance. Secondly, the operator provisions a Prometheus server using the [Prometheus operator](https://github.com/coreos/prometheus-operator). The Prometheus operator is an optional dependency.

## Exposing Metrics
A service is created for each TLPStress instance and exposes metrics on port 9500. The name of the service will be of the format `<TLPStress-name>-metrics`. Suppose we have a TLPStress instance named `stress-demo`. The operator will create a service that looks like:

```
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2019-10-25T20:37:00Z"
  labels:
    app: tlpstress
    tlpstress: stress-demo
  name: stress-demo-metrics
  namespace: tlpstress
  ownerReferences:
  - apiVersion: thelastpickle.com/v1alpha1
    kind: TLPStress
    name: stress-demo
    uid: 10abc800-b6e1-4552-a5ee-4efc1355debc
  resourceVersion: "81734"
  selfLink: /api/v1/namespaces/tlpstress/services/stress-demo-metrics
  uid: 3d025e0e-1968-4929-a744-cce46554fad8
spec:
  clusterIP: 10.96.170.220
  ports:
  - name: metrics
    port: 9500
    protocol: TCP
    targetPort: 9500
  selector:
    app: tlpstress
    tlpstress: stress-demo
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```
The `labels` and `selector` fields are configured with the TLPStress instance name.

There is an owner reference to the TLPStress instance. When the TLPStress instance is deleted, the service will get deleted as well.

Lastly, the port name on which metrics is exposed is `metrics`.

## Prometheus Operator
The tlp-stress operator itself does not yet provision a Prometheus server. It is currently done when deploying the operator:

```
$ make deploy-all
```

This will first deploy the [Promtheus Operator](https://github.com/coreos/prometheus-operator) into the `prometheus-operator` namespace with cluster-wide permissions. 

The Prometheus operator adds several new CRDs, notably `ServiceMonitor` and `Prometheus`.

### ServiceMonitor
The `ServiceMonitor` CRD defines how a set of services should be monitored. It facilitates service discovery.

If the `ServiceMonitor` CRD exists in the cluster, then the tlp-stress operator will create a `ServiceMonitor` that looks like this:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2019-10-25T20:37:00Z"
  generation: 1
  labels:
    app: tlpstress
  name: tlpstress-servicemonitor
  namespace: tlpstress
  ownerReferences:
  - apiVersion: apps/v1    
    kind: Deployment
    name: tlp-stress-operator-74d5c9c65b
    uid: 10abc800-b6e1-4552-a5ee-4efc1355debc
  resourceVersion: "81736"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/tlpstress/servicemonitors/stress-demo-metrics
  uid: 1e02ee99-a415-489b-bdce-faa9c16932e5
spec:
  endpoints:
  - port: metrics
  namespaceSelector: {}
  selector:
    matchLabels:
      app: tlpstress
``` 

The tlp-stress operator's Deployment is set as an owner reference.

`selector` determines which services will be discovered.

`endpoints` specifies that the `metrics` port for each service should be scraped.

### Prometheus

The `Prometheus` CRD specifies a Prometheus server to be provisioned.

The `Prometheus` instance is a namepace-based deployment with permissions limited to the namespace in which it is deployed. It is deployed in the same namespace as the tlp-stress operator, which defaults to `tlpstress`.

The `Prometheus` object looks like:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: tlpstress
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      app: tlpstress
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```

**Note:** The `Prometheus` object currently is not configurable through the tlp-stress operator. You will need to directly edit [config/prometheus/bundle.yaml](../config/prometheus/bundle.yaml) at deployment time.

#### Accessing Prometheus Console
A service is deployed to expose the Prometheus web UI. It looks like this:

```
$ kubectl get svc prometheus-tlpstress
NAME                   TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
prometheus-tlpstress   ClusterIP   10.80.14.96   <none>        9090/TCP   5h11m
```

This is a ClusterIP service, which can be accessed with `kubectl port-forward`:

```
$ kubectl port-forward svc/prometheus-tlpstress 9090:9090
```