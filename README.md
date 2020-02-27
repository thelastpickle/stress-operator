# tlp-stress Operator
A Kubernetes operator for running tlp-stress in Kubernetes.

**Build Status**  [![CircleCI](https://circleci.com/gh/jsanda/stress-operator/tree/master.svg?style=svg)](https://circleci.com/gh/thelastpickle/stress-operator/tree/master)

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
* Integration with [casskop](https://github.com/Orange-OpenSource/cassandra-k8s-operator) to provision Cassandra clusters 

**Note:** This operator is pre-alpha and subject to breaking changes on a regular basis.

## Requirements
The following libraries/tools need to be installed in order to build and deploy this operator:

* Go >= 1.13.0
* Docker client >= 17
* kubectl >= 1.13
* [Operator SDK](https://github.com/operator-framework/operator-sdk) = 0.14.0

## Installation
Go to the [releases](https://github.com/thelastpickle/stress-operator/releases) on GitHub and download `stress-operator.yaml` from the desired version. Then run:

`$ kubectl apply -f stress-operator.yaml`

Verify that the operator is running:

```
$ kubectl get pods -l name=stress-operator
NAME                                   READY   STATUS    RESTARTS   AGE
stress-operator-65fbbbd9f8-9tbcl       1/1     Running   0          30s
```

### Provisioning Cassandra
stres-operator can optionally provision Cassandra cluster using [casskop](https://github.com/Orange-OpenSource/casskop). If you want to enable this feature, download `casskop.yaml` and run:

`$ kubectl apply -f casskop.yaml`

### Monitoring with Prometheus and Grafana
stress-operator can optionally provision and condfigure Prometheus and Grafana. tlp-stress metrics will be stored in Prometheus and dashboards will get created in Grafana. If you want to enable this feature, download `prometheus-operator.yaml` and `grafana-operator.yaml` and then run:

```
$ kubectl apply -f prometheus-operator.yaml

$ kubectl apply -f grafana-operator.yaml
```

## Running tlp-stress
We run tlp-stress by creating a `Stress` custom resource. Consider the following example, `stress.yaml`:

```yaml
apiVersion: thelastpickle.com/v1alpha1
kind: Stress
metadata:
  name: stress-demo
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
    cassandraService: stress-demo
```

The spec is divided into three sections - `stressConfig`, `jobConfig`, and `cassandraConfig`.

`stressConfig` has properties that map directly to the tlp-stress command line options.

`jobConfig` has properties for configuring the k8s job that the operator deploys.

`cassandraConfig` has properties for the Cassandra cluster that tlp-stress will run against, namely how to connect to the cluster.

**Note:** Please see [the documenation](documentation/stress.md) for a complete overview of the Stress custom resource.

Now let's create the Stress instance:

```
$ kubectl apply -f stress.yaml
stress.thelastpickle.com/stress-demo created
```

Verify that the object was created:

```
$ kubectl get stress
NAME          AGE
stress-demo   17s
```

The operator should create a k8s Job:

```
$ kubectl get jobs --selector=stress=stress-demo
NAME             COMPLETIONS   DURATION   AGE
stress-demo   1/1           9s         14s
```

The job completed quickly because there is no Cassandra cluster running. Please see [the documentation](./documentation/cassandra.md) for options on how to deploy Cassandra in Kubernetes.

Now let's delete the Stress instance:

```
$ kubectl delete stress stress-demo
stress.thelastpickle.com "stress-demo" deleted
```

Verify that the object was deleted:

```
$ kubectl get stress
No resources found.
```

Verify that the job was also deleted:

```
$ kubectl get jobs --selector=stress=stress-demo
No resources found.
```

## Documentation
The operator provides a lot of functionaly for provisioning Cassandra clusters and for monitoring tlp-stress with Grafana and Prometheus.

Please check out [the docs](./documentation/README.md) for a complete overview.
