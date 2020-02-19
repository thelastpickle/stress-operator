#!/bin/bash

set -e


if [ -z ${1+x} ]; then
  echo "Usage: prepare-artifacts.sh <stress-operator image>"
  exit 1
fi
image=$1

mkdir -p artifacts

for f in `ls deploy/crds`
do
  cat deploy/crds/$f >> artifacts/stress-operator.yaml
  echo "---" >> artifacts/stress-operator.yaml
done

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
