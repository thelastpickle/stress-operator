# tlp-stress Operator
This operator runs [tlp-stress](https://github.com/thelastpickle/tlp-stress), which is a workload-centric stress tool for Apache Cassandra. 

The operator is build with the [Operator SDK](https://github.com/operator-framework/operator-sdk).

**Note:** This operator is pre-alpha and subject to breaking changes on a regular basis.

The operator runs tlp-stress as a [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion).

## Requirements
You need to have the [Operator SDK](https://github.com/operator-framework/operator-sdk) installed in order to build this project.

## Usage
Currently you need to clone the git repo in order to deploy the operator.

**Clone the repo**

```
$ git clone https://github.com/jsanda/tlp-stress-operator
$ cd tlp-stress-operator
```

**Deploy resources**

```
$ kubectl apply -f deploy/crds/thelastpickle_v1alpha1_tlpstress_crd.yaml
customresourcedefinition.apiextensions.k8s.io/tlpstresses.thelastpickle.com created

$ kubectl apply -f deploy/
deployment.apps/tlp-stress-operator created
role.rbac.authorization.k8s.io/tlp-stress-operator created
rolebinding.rbac.authorization.k8s.io/tlp-stress-operator created
serviceaccount/tlp-stress-operator created
```

**Verify that the operator is running**

```
$ kubectl get pods -l name=tlp-stress-operator
NAME                                   READY   STATUS    RESTARTS   AGE
tlp-stress-operator-65fbbbd9f8-9tbcl   1/1     Running   0          30s
```

## TLPStress Custom Resource 
TODO

## CircleCI Status
[![CircleCI](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master.svg?style=svg)](https://circleci.com/gh/jsanda/tlp-stress-operator/tree/master)
