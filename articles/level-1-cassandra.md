# Develop and Deploy a Level 1 JanusGraph Operator using Cassandra - Part 2

In this tutorial, we will learn how to develop and deploy a Level 1 operator on the OpenShift Container Platform. We will be creating an operator for Janusgraph which will use Cassandra as an storage backend. Cassandra is a distributed database platform which can scale and be highly available, and can perform really well on any commodity hardware or cloud infrastructure. To learn more about Cassandra you can read more [here](https://casandra.apache.org).

When the reader has completed this tutorial, they will understand how to:
* Deploy Apache Cassandra as backend storage.
* Generate Janusgraph image customize to Openshift cluster.
* Deploy Janusgraph operator to openshift cluster.
* Scale Janusgraph instance up/down by modifying and applying the Custom Resource (CR) to Openshift cluster.

> Note: Cassandra deployment is not part of this tutorial. We assume that Cassandra is already available whether its deployed from operator hub or as a standalone deployment.

## Flow

![Architecture](../images/architecture.png)


## Included components

* [Apache Cassandra](https://cassandra.apache.org/) - The Apache Cassandra database is the right choice when you need scalability and high availability without compromising performance.
* [Janusgraph](https://janusgraph.org) - JanusGraph is a scalable graph database optimized for storing and querying graphs containing hundreds of billions of vertices and edges distributed across a multi-machine cluster.
* [Red Hat OpenShift](http://www.openshift.com) - Red Hat OpenShift is the hybrid cloud platform of open possibility: powerful, so you can build anything and flexible, so it works anywhere.


## Featured Technology

* [RedHat OpenShift Operator](https://www.openshift.com/learn/topics/operators): Automate the creation, configuration, and management of instances of Kubernetes-native applications.

## Steps

1. [Deploy Cassandra to OpenShift](#1-deploy-cassandra-to-openshift)
1. [Clone and Modify Janusgraph docker image](#2-clone-and-modify-janusgraph-docker-image)
1. [Deploy Janusgraph operator](#3-deploy-janusgraph-operator)
1. [Load and test retrieve of data using gremlin console](#3-load-and-test-retrieve-of-data-using-gremlin-console)
1. [Test Janusgraph operator for scale up and down](#4-test-janusgraph-operator-for-scale-up-and-down)

### 1. Deploy Cassandra to OpenShift

Clone the `cassandra-openshift` locally. In a terminal, run:

```bash
git clone https://github.com/IBM/janusgraph-operator.git

cd cassandra-openshift
```

We need to update the default configurations of Cassandra so that it can be deployed to OpenShift. The changes are defined in the `Dockerfile`. In order to adapt to openshift environment, we need to change the group ownership and file permission to root. Although OpenShift runs containers using an arbitrarily assigned user ID, the group ID must always be set to the root group (0). And there are other changes that Cassandra needs for it to be successfully deployed which will not be covered part of this tutorial.

You can build and push the Cassandra image to your image repository by running following commands:
```bash
docker build -t cassandra:1.0

docker tag cassandra:1.0 <repository hostname>/<username>/cassandra:1.0

docker push <repository hostname>/<username>/cassandra:1.0

```

> Note: change "repository hostname" and "username" accordingly.

Once the image is built, you can now deploy Cassandra as `Stateful set` in openshift.

Run the following command to deploy Cassandra from the cloned directory in the terminal:

```bash
oc apply -f cassandra-app-v1.yaml -f cassandra-svc-v1.yaml
```

To make sure cassandra is running, it should create one instance of cassandra database. If you want to have multiple replicas, you can modify replicas in the `cassandra-app-v1.yaml`.

![Cassandra Pod](../images/cassandra-deployment.png)

### 2. Clone and Modify Janusgraph docker image

The Janusgraph docker image from the official repo deploys fine with kubernetes but not for openshift. There are few things that needs to be modified before we deploy:

* Fork the repo `https://github.com/JanusGraph/janusgraph-docker`
* Change the file and group ownership to root (0) for related folders. Following modifications are applies to the `Dockerfile`:
```bash
chgrp -R 0 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \
chmod -R g+w ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR}

RUN chmod u+x /opt/janusgraph/bin/gremlin.sh
RUN chmod u+x /opt/janusgraph/conf/remote.yaml

```
* Change the `JANUS_PROPS_TEMPLATE` value to `cql` as we will be using cassandra as backend.
* Since we will only be using the latest version change the version to the latest in `build-images.sh`. We will create a copy of that file to `build-images-ibm.sh` and modify it there. Modifications includes commenting out few lines. The following modification will be applied to the build script:

```bash
# optional version argument
version="${1:-}"
# get all versions
# versions=($(ls -d [0-9]*))
# get the last element of sorted version folders
# latest_version="${versions[${#versions[@]}-1]}"

# override to run the latest version only:
versions="0.5"
latest_version="0.5"

```
* Create `janusgraph-cql-server.properties` in the latest version directory, which in our case is `0.5` and add the following porperties:

```bash
gremlin.graph=org.janusgraph.core.JanusGraphFactory
storage.backend=cql
storage.hostname=cassandra-service
storage.username=cassandra
storage.password=cassandra
storage.cql.keyspace=janusgraph
storage.cassandra.replication-factor=3
storage.cassandra.replication-strategy-class=org.apache.cassandra.locator.NetworkTopologyStrategy
cache.db-cache = true
cache.db-cache-clean-wait = 20
cache.db-cache-time = 180000
cache.db-cache-size = 0.5
storage.directory=/var/lib/janusgraph/data
index.search.backend=lucene
index.search.directory=/var/lib/janusgraph/index

```
These are properties that allows Janusgraph to talk to Cassandra as Cassandra will be storing the data in a distributed fashion.

After these changes, make sure to update `janusgraph-cql-server.properties` with the `cluster-ip` of cassandra service. Update `storage.hostname` with the `Cluster-IP`.

![Cluster IP](../images/cluster-ip.png)

Now we can build and deploy the Janusgraph docker image to openshift by running:

```bash
./build-images-ibm.sh -- if you have created a new file
```
OR

```bash
./build-images.sh -- if you have modified file provided by Janusgraph
```

### 3. Load and test retrieve of data using gremlin console