#!/bin/bash
set -x
set -e

img="sanjeevghimire/janusgraph-operator:1.0.21"
namespace="sanjeev-janus-demo"

make deploy IMG=$img

kubectl apply -f config/samples/graph_v1alpha1_janusgraph.yaml
