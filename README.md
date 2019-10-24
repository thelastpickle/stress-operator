# tlp-stress Operator
A Kubernetes operator for running tlp-stress in Kubernetes.

**Build Status**  [![CircleCI](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master.svg?style=svg)](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master)

## Overview
tlp-stress is a workload-centric stress tool for Apache Cassandra. tlp-stress already has an impressive and intuitive feature set for modelling different workloads.

If you are not already familiar with tlp-stress, check out the following resources  to learn more:

* [tlp-stress docs](http://thelastpickle.com/tlp-stress/)
* [Introduction to tlp-stress](https://thelastpickle.com/blog/2018/10/31/tlp-stress-intro.html)
* [A Dozen New Features in tlp-stress](https://thelastpickle.com/blog/2019/07/09/tlp-stress-update.html)

The operator provides a number of features including:

* An easy way to configure and deploy tlp-stress as a Kubernetes job
* Integration with Prometheus by:
  * Exposing metrics that can be scraped by a Prometheus server
  * Provisioning a Prometheus server that is configured to monitor tlp-stress jobs
* Integration with Grafana by:
  * Provisioning a Grafana server that is configured to use Prometheus as a data source
  * Adding a dashboard for each tlp-stress job
* Integration with [CassKop](https://github.com/Orange-OpenSource/cassandra-k8s-operator) to provision Cassandra clusters 

**Note:** This operator is pre-alpha and subject to breaking changes on a regular basis.

## Requirements
The following libraries/tools need to be installed in order to build and deploy this operator:

* Go >= 1.11
* Docker client >= 17
* kubectl >= 1.13
* [Operator SDK](https://github.com/operator-framework/operator-sdk) >= 0.9.0

## Installation
You need to clone this git repo. The operator is installed by running the Makefile. 

The project uses Go modules. If you clone the repo under `GOPATH` then make you the `GO111MODULE` environment variable set:

```
$ export GO111MODULE=on
```

The following sections will cover a couple different installation options.

Both options are namespace-based deployments. The operator is configured with permissions only for the namespace in which it is deployed. 

### Basic Installation
To deploy just the operator by itself without anything else, such as Prometheus or Grafana, run `make deploy`:   

```
$ make deploy
./scripts/create-ns.sh tlpstress
Creating namespace tlpstress
namespace/tlpstress created
kubectl -n tlpstress apply -f deploy/crds/thelastpickle_v1alpha1_tlpstress_crd.yaml
customresourcedefinition.apiextensions.k8s.io/tlpstresses.thelastpickle.com created
kubectl -n tlpstress apply -f deploy/service_account.yaml
serviceaccount/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/role.yaml
role.rbac.authorization.k8s.io/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/role_binding.yaml
rolebinding.rbac.authorization.k8s.io/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/operator.yaml
deployment.apps/tlp-stress-operator created
```

**Note:** By default the operator is deployed to the `tlpstress` namespace which will be created if it does not already exist.

### Full, Integrated Installation
If you want the integrations with Prometheus, Grafana, and CassKop, then run
`make deploy-all`:

```
$ make deploy-all
./scripts/create-ns.sh tlpstress
Creating namespace tlpstress
namespace/tlpstress created
kubectl -n tlpstress apply -f config/casskop
customresourcedefinition.apiextensions.k8s.io/cassandraclusters.db.orange.com created
role.rbac.authorization.k8s.io/cassandra-k8s-operator created
rolebinding.rbac.authorization.k8s.io/cassandra-k8s-operator created
serviceaccount/cassandra-k8s-operator created
deployment.apps/cassandra-k8s-operator created
kubectl apply -f config/prometheus-operator/bundle.yaml
namespace/prometheus-operator created
clusterrolebinding.rbac.authorization.k8s.io/prometheus-operator created
clusterrole.rbac.authorization.k8s.io/prometheus-operator created
deployment.apps/prometheus-operator created
serviceaccount/prometheus-operator created
service/prometheus-operator created
until kubectl get crd prometheuses.monitoring.coreos.com > /dev/null 2>&1; do \
		echo "Waiting for prometheuses.monitoring.coreos.com CRD to be deployed"; \
		sleep 1; \
	done;
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
Waiting for prometheuses.monitoring.coreos.com CRD to be deployed
kubectl -n tlpstress apply -f config/prometheus/bundle.yaml
serviceaccount/prometheus created
role.rbac.authorization.k8s.io/prometheus created
rolebinding.rbac.authorization.k8s.io/prometheus created
prometheus.monitoring.coreos.com/tlpstress created
service/prometheus-tlpstress created
kubectl -n tlpstress apply -f deploy/crds/thelastpickle_v1alpha1_tlpstress_crd.yaml
customresourcedefinition.apiextensions.k8s.io/tlpstresses.thelastpickle.com created
kubectl -n tlpstress apply -f deploy/service_account.yaml
serviceaccount/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/role.yaml
role.rbac.authorization.k8s.io/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/role_binding.yaml
rolebinding.rbac.authorization.k8s.io/tlp-stress-operator created
kubectl -n tlpstress apply -f deploy/operator.yaml
deployment.apps/tlp-stress-operator create
```
**Note:** By default the operator is deployed to the `tlpstress` namespace which will be created if it does not already exist.

Everything that gets deployed by `make deploy-all` is optional. The tlp-stress operator can and will utilize resources like Prometheus and Grafana if they are deployed; however, they are not required.

**Verify that the operator is running**
```
$ kubectl -n tlpstress get pods -l name=tlp-stress-operator
NAME                                   READY   STATUS    RESTARTS   AGE
tlp-stress-operator-65fbbbd9f8-9tbcl   1/1     Running   0          30s
```


## TLPStress Custom Resource 
The operator manages TLPStress objects. A TLPStress instance specifies how to configure and run tlp-stress.

Let's look at an example manifest as introduction to the TLPStress CRD:

```yaml
apiVersion: thelastpickle.com/v1alpha1
kind: TLPStress
metadata:
  name: stress-example
spec:
  stressConfig:
    workload: KeyValue
    partitions: 25m
    duration: 60m
    readRate: "0.3"
    replication:
      networkTopologyStrategy:
        dc1: 2
    partitionGenerator: sequence
  jobConfig:
    parallelism: 2
  cassandraConfig:
    cassandraService: stress-example
```
The spec is divided into three sections - `stressConfig`, `jobConfig`, and `cassandraConfig`.

`stressConfig` has properties that correspond to the tlp-stress command line options.

`jobConfig` has properties for configuring the k8s job.

`cassandraConfig` has properties for the Cassandra cluster that tlp-stress will run against, namely how to connect to the cluster.

## CassKop Integration
TODO

## Prometheus Integration
TODO

## Grafana Integration
TODO
