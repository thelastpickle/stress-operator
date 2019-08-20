#!/bin/bash

set -e

init_helm() {
  echo "Creating ServiceAccount and ClusterRoleBinding for tiller"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: tiller
    namespace: kube-system
EOF

  echo "Initializing helm"
  helm init --service-account tiller

  # TODO wait for tiller deployment to be ready
}

init_casskop() {
  repo="https://Orange-OpenSource.github.io/cassandra-k8s-operator/helm"

  echo "Creating namespace cassandra"
  kubectl create namespace cassandra

  echo "Installing casskop from $repo"
  helm repo add casskop https://Orange-OpenSource.github.io/cassandra-k8s-operator/helm
  helm install --wait --name casskop casskop/cassandra-operator

  # TODO wait for casskop deployment to be ready
}

init_tlp-stress() {
  echo "Creating tlp-stress-operator resources"
  
  cd $(dirname "$BASH_SOURCE[0]")
  cd ..

  kubectl apply -f deploy/crds/thelastpickle_v1alpha1_tlpstress_crd.yaml
  kubectl apply -f deploy

  # TODO wait for tlp-stress-operator deployment to be ready
}

if [ -z $(which helm) ]; then
  echo "You must have helm installed in order to use this script."
  echo "See https://helm.sh/docs/using_helm/#installing-helm for details."
  exit 1
fi

init_helm
init_casskop
init_tlp-stress

