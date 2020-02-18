#!/bin/bash

set -e

image=$1

mkdir artifacts

cat deploy/service_account.yaml >> artifacts/stress-operator.yaml
echo "---" >> artifacts/stress-operator.yaml
cat deploy/role.yaml >> artifacts/stress-operator.yaml
echo "---" >> artifacts/stress-operator.yaml
cat deploy/role_binding.yaml >> artifacts/stress-operator.yaml
echo "---" >> artifacts/stress-operator.yaml
cat deploy/operator.yaml >> artifacts/stress-operator.yaml

cp -R deploy/crds artifacts
cp config/casskop/casskop.yaml artifacts/
cp config/grafana/grafana-operator.yaml artifacts/
cp config/prometheus-operator/prometheus-operator.yaml artifacts

sed -i -e 's@'docker.io/thelastpickle/stress-operator:latest'@ '$image'@' artifacts/stress-operator.yaml
