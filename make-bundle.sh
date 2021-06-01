#!/bin/bash
set -x
set -e

REGISTRY=docker.io
USERNAME=sanjeevghimire
REPOSITORY=janusgraph-operator
VERSION=1.0.13

OPERATOR_IMAGE=$REGISTRY/$USERNAME/$REPOSITORY:$VERSION

make bundle IMG=$OPERATOR_IMAGE