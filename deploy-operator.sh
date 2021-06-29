#!/bin/bash
set -x
set -e

img="sanjeevghimire/janusgraph-operator:1.0.16"
namespace="sanjeev-janus"

make undeploy

make deploy IMG=$img

kubectl apply -f config/samples/graph_v1alpha1_janusgraph.yaml
