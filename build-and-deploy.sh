#!/bin/bash
set -x
set -e

img="sanjeevghimire/janusgraph-operator:latest"
namespace="sanjeev-janus"

cd config/manager
kustomize edit set namespace $namespace
kustomize edit set image controller=$img
cd ../../
cd config/default
kustomize edit set namespace $namespace
cd ../../

make docker-build IMG=$img
make docker-push IMG=$img
make deploy IMG=$img

kubectl apply -f config/samples/graph_v1alpha1_janusgraph.yaml
