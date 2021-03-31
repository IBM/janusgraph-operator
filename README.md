# janusgraph-operator

This repo holds the code for deploying the JanusGraph operator. It also holds the tutorials to describe the steps to deploy
the operator, and what the controller code in the operator is doing.

1. [Develop and Deploy a Level 1 JanusGraph Operator on OpenShift Container Platform](https://github.ibm.com/TT-ISV-org/janusgraph-operator/blob/main/articles/level-1-operator.md): 
In this tutorial, we will discuss how to develop and deploy a Level 1 operator on the OpenShift Container Platform. We will use the 
[Operator SDK Capability Levels](https://operatorframework.io/operator-capabilities/) as our guidelines for what is considered a 
level 1 operator. 
* Part 1 of the tutorial we will deploy JanusGraph using the default (BerkeleyDB) 
backend storage. This will be a simple approach, and only recommended for testing purposes.
* Part 2 will feature Cassandra as the backend storage for JanusGraph, which is more suitable for for production use cases. 

