# tlp-stress Operator
A Kubernetes operator for running tlp-stress in Kubernetes.

**Build Status**  [![CircleCI](https://circleci.com/gh/jsanda/stress-operator/tree/master.svg?style=svg)](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master)

## Overview
tlp-stress is a workload-centric stress tool for Apache Cassandra. It provides a rich feature set for modelling different workloads with Cassandra.

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

* Go >= 1.13.0
* Docker client >= 17
* kubectl >= 1.13
* [Operator SDK](https://github.com/operator-framework/operator-sdk) = 0.14.0

## Installation
You need to clone this git repo. The operator is installed by running the Makefile. 

The following sections will cover a couple different installation options.

Both options are namespace-based deployments. The operator is configured with permissions only for the namespace in which it is deployed. 

By default the operator is deployed to the `tlpstress` namespace which will be created if it does not already exist.

Both of the installations options discussed belowed install the latest published version of the operator image. If you want to deploy a different version, for now you need to change the image tag in `deploy/operator.yaml`:

`image: docker.io/thelastpickle/stress-operator:latest`

### Basic Installation
To deploy the operator by itself without anything else, such as Prometheus or Grafana, run `make deploy`:   

```
$ make deploy
./scripts/create-ns.sh tlpstress
Creating namespace tlpstress
namespace/tlpstress created
kubectl -n tlpstress apply -f deploy/crds/
customresourcedefinition.apiextensions.k8s.io/stresses.thelastpickle.com created
kubectl -n tlpstress apply -f deploy/service_account.yaml
serviceaccount/stress-operator created
kubectl -n tlpstress apply -f deploy/role.yaml
role.rbac.authorization.k8s.io/stress-operator created
kubectl -n tlpstress apply -f deploy/role_binding.yaml
rolebinding.rbac.authorization.k8s.io/stress-operator created
kubectl -n tlpstress apply -f deploy/operator.yaml
deployment.apps/stress-operator created
```

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
kubectl -n tlpstress apply -f config/prometheus/bundle.yaml
serviceaccount/prometheus created
role.rbac.authorization.k8s.io/prometheus created
rolebinding.rbac.authorization.k8s.io/prometheus created
prometheus.monitoring.coreos.com/tlpstress created
service/prometheus-tlpstress created
kubectl -n tlpstress apply -f deploy/crds/
customresourcedefinition.apiextensions.k8s.io/stresses.thelastpickle.com created
kubectl -n tlpstress apply -f deploy/service_account.yaml
serviceaccount/stress-operator created
kubectl -n tlpstress apply -f deploy/role.yaml
role.rbac.authorization.k8s.io/stress-operator created
kubectl -n tlpstress apply -f deploy/role_binding.yaml
rolebinding.rbac.authorization.k8s.io/stress-operator created
kubectl -n tlpstress apply -f deploy/operator.yaml
deployment.apps/stress-operator create
```
**Note:** By default the operator is deployed to the `tlpstress` namespace which will be created if it does not already exist.

Please see [the documentation](./documentation/integrated-install.md) for details on everything that gets installed and deployed with `make deploy-all`.

Everything that gets deployed by `make deploy-all` is optional. The tlp-stress operator can and will utilize resources like Prometheus and Grafana if they are deployed; however, they are not required.

**Verify that the operator is running**
```
$ kubectl -n tlpstress get pods -l name=tlp-stress-operator
NAME                                   READY   STATUS    RESTARTS   AGE
tlp-stress-operator-65fbbbd9f8-9tbcl   1/1     Running   0          30s
```

## Running tlp-stress
We run tlp-stress by creating a `TLPStress` custom resource. Consider the following example, [example/stress-demo.yaml](./example/stress-demo.yaml):

```yaml
apiVersion: thelastpickle.com/v1alpha1
kind: TLPStress
metadata:
  name: tlpstress-demo
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
    parallelism: 1
  cassandraConfig:
    cassandraService: tlpstress-demo
  image: thelastpickle/tlp-stress:3.0.0
  imagePullPolicy: IfNotPresent
```

The spec is divided into three sections - `stressConfig`, `jobConfig`, and `cassandraConfig`.

`stressConfig` has properties that map directly to the tlp-stress command line options.

`jobConfig` has properties for configuring the k8s job that the operator deploys.

`cassandraConfig` has properties for the Cassandra cluster that tlp-stress will run against, namely how to connect to the cluster.

**Note:** Please see [the documenation](documentation/stress.md) for a complete overview of the TLPStress custom resource.

Now let's create the TLPStress instance:

```
$ kubectl -n tlpstress apply -f example/stress-demo.yaml
tlpstress.thelastpickle.com/stress-demo created
```

Verify that the object was created:

```
$ kubectl -n tlpstress get tlpstress
NAME          AGE
stress-demo   17s
```

The operator should create a k8s Job:

```
$ kubectl -n tlpstress get jobs --selector=tlpstress=tlpstress-demo
NAME             COMPLETIONS   DURATION   AGE
tlpstress-demo   1/1           9s         14s
```

The job completed quickly because there is no Cassandra cluster running. Please see [the documentation](./documentation/cassandra.md) for options on how to deploy Cassandra in Kubernetes.

Now let's delete the TLPStress instance:

```
$ kubectl -n tlpstress delete tlpstress tlpstress-demo
tlpstress.thelastpickle.com "tlpstress-demo" deleted
```

Verify that the object was deleted:

```
$ kubectl -n tlpstress get tlpstress
No resources found.
```

Verify that the job was also deleted:

```
$ kubectl -n tlpstress get jobs --selector=tlpstress=tlpstress-demo
No resources found.
```

## Documentation
The operator provides a lot of functionaly for provisioning Cassandra clusters and for monitoring tlp-stress. 

Please check out [the docs](./documentation/README.md) for a complete overview.