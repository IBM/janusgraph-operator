#!/bin/bash
set -x
set -e


export USERNAME="sanjeevghimire"
export BUNDLE_IMG="docker.io/$USERNAME/janusgraph-operator-bundle:v0.0.1"

make bundle-build IMG=$BUNDLE_IMG

make docker-push IMG=$BUNDLE_IMG