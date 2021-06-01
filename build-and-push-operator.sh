#!/bin/bash
set -x
set -e

export img="sanjeevghimire/janusgraph-operator:1.0.14"
export namespace="sanjeev-janus"

cd config/manager
kustomize edit set namespace $namespace
kustomize edit set image controller=$img
cd ../../
cd config/default
kustomize edit set namespace $namespace
cd ../../


make docker-build IMG=$img
make docker-push IMG=$img