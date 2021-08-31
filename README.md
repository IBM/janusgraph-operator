# Kubernetes Operators Learning Path - Advanced Level

**Note: If you want to see beginner and intermediate level tutorials, you can find them [here](https://github.com/IBM/create-and-deploy-memcached-operator-using-go/blob/main/README.md).**

1. <b>Develop and Deploy a Level 1 JanusGraph Operator on OpenShift Container Platform</b>: 
In this tutorial, we will discuss how to develop and deploy a Level 1 operator on the OpenShift Container Platform. We will use the 
[Operator SDK Capability Levels](https://operatorframework.io/operator-capabilities/) as our guidelines for what is considered a 
level 1 operator. 
- [Part 1](https://github.com/IBM/create-and-deploy-memcached-operator-using-go/blob/main/articles/level-1-operator.md) of the tutorial we will deploy JanusGraph using the default (BerkeleyDB) backend storage. This will be a simple approach, and only recommended for testing purposes.
- [Part 2](https://github.com/IBM/janusgraph-operator/blob/main/articles/level-1-janusgraph.md) will feature Cassandra as the backend storage for JanusGraph, which is more suitable for for production use cases. 

2. [The Operator Cookbook: How to make an operator from scratch](https://github.com/IBM/janusgraph-operator/blob/main/articles/operator-cookbook.md): In this article, we will discuss common building blocks for level 1 operators, and what logic a service vendor would need to write themselves in order to build a level 1 operator.

3. [LEVEL 2 : Seamless upgrade for JanusGraph Operator](https://github.com/IBM/janusgraph-operator/blob/main/articles/level-2-janusgraph.md): This tutorial builds on the Level 1 JanusGraph operator to add Level 2 capabilities.

4. [Certifying Janusgraph Image](https://github.com/IBM/janusgraph-operator/blob/main/articles/janusgraph-certification.md) shows the steps to submit the JanusGraph image and its operator for Red Hat certification.
