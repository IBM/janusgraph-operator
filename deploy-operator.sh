#!/bin/bash
set -x
set -e

img="sanjeevghimire/janusgraph-operator:1.0.11"
namespace="sanjeev-janus"

make deploy IMG=$img

kubectl apply -f config/samples/graph_v1alpha1_janusgraph.yaml
