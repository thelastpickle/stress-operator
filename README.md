# tlp-stress Operator
A Kubernetes operator for running tlp-stress in Kubernetes.

tlp-stress is a workload-centric stress tool for Apache Cassandra.  To learn more about tlp-stress see the [project site](http://thelastpickle.com/tlp-stress/) as well as these articles:

* [Introduction to tlp-stress](https://thelastpickle.com/blog/2018/10/31/tlp-stress-intro.html)
* [A Dozen New Features in tlp-stress](https://thelastpickle.com/blog/2019/07/09/tlp-stress-update.html)

The operator provides a number of features including:

* Configure and deploy tlp-stress with the `TLPStress` custom resource
* Expose metrics and services for integration with Prometheus for monitoring
* Deploy Prometheus
* Deploy Grafana and dasboard for view Prometheus metrics
* Provision Cassandra cluster

**Note:** This operator is pre-alpha and subject to breaking changes on a regular basis.

## Requirements
The following libraries/tools need to be installed in order to build and deploy this operator:

* Go >= 1.11
* Docker client >= 17
* kubectl >= 1.13
* [Operator SDK](https://github.com/operator-framework/operator-sdk) >= 0.9.0 

## Installation
The operator is installed by running the Makefile. This section will cover a couple different installation options. 

**Note:** Helm charts may made available in the future.

### Install Only the Operator
To deploy only the operator without accompanying resources, like Prometheus, run `make deploy`:   

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

**Note:** The operator is deployed to the `tlpstress` namespace which will be created if it does not already exist.

**Verify that the operator is running**
```
$ kubectl -n tlpstress get pods -l name=tlp-stress-operator
NAME                                   READY   STATUS    RESTARTS   AGE
tlp-stress-operator-65fbbbd9f8-9tbcl   1/1     Running   0          30s
```


### Install the Operator with Accompanying Resources
To deploy the operator with accompanying resources run `make deploy-all`:

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

## TLPStress Custom Resource 
TODO

## CassKop Integration
TODO

## Prometheus Integration
TODO

## Grafana Integration
TODO

## CircleCI Status
[![CircleCI](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master.svg?style=svg)](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master)
