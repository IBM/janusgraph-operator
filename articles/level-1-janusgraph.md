# Develop and deploy a level 1 JanusGraph operator using Apache Cassandra

## SUBTITLE: Deploy an instance of the JanusGraph database simply by creating a custom resource

This tutorial shows you how to develop and deploy a level 1 operator on the Red Hat OpenShift Container Platform. You will create an operator for JanusGraph that uses [Apache Cassandra](https://cassandra.apache.org/) as a storage back end. Cassandra is a distributed database platform that can scale and be highly available, and can perform really well on any commodity hardware or cloud infrastructure.

When you have completed this tutorial, you will understand how to:

* Deploy Cassandra as back-end storage.
* Create a JanusGraph image that runs well in OpenShift, not just Kubernetes.
* Deploy a JanusGraph operator to an OpenShift cluster.
* Scale a JanusGraph instance up or down by modifying and applying the Custom Resource (CR) to an OpenShift cluster.

_**Note:** Cassandra deployment is not part of this tutorial. We assume that Cassandra is already available, whether its deployed from operator hub or as a stand-alone deployment._

A level 1 JanusGraph operator has the following capabilities:

* Deploys JanusGraph by creating its Services, Deployments, and RoleBinding
* Ensures that managed resources reach a healthy state, and conveys readiness of the resources to the user through the status block of the custom resource
* Manages scalability by resizing the underlying resources in response to changes in the custom resource

## Flow

![Architecture](../images/architecture.png)


## Included components

* [Apache Cassandra](https://cassandra.apache.org/) -- The Cassandra database is the right choice when you need scalability and high availability without compromising performance.
* [JanusGraph](https://janusgraph.org) -- JanusGraph is a scalable graph database that's optimized for storing and querying graphs containing hundreds of billions of vertices and edges distributed across a multi-machine cluster.
* [Red Hat OpenShift](http://www.openshift.com) -- OpenShift is a powerful, flexible hybrid cloud platform that enables you to build a wide range of solutions that work anywhere.


## Featured technology

* [RedHat OpenShift Operator](https://www.openshift.com/learn/topics/operators): Operator automates the creation, configuration, and management of Kubernetes-native application instances.

## Prerequisites

To complete this tutorial, we assume that you:

* have little or no experience developing operators
* have some knowledge of Kubernetes Operators concepts
* have created a [memcached operator](https://developer.ibm.com/learningpaths/kubernetes-operators/develop-deploy-simple-operator/create-operator/)
* have read [Explanation of memcached operator code](https://developer.ibm.com/learningpaths/kubernetes-operators/develop-deploy-simple-operator/deep-dive-memcached-operator-code/)
* have set up your environment as shown in the [Set up your environment](https://developer.ibm.com/learningpaths/kubernetes-operators/develop-deploy-simple-operator/installation/) tutorial

## Steps

1. [Deploy Cassandra to OpenShift](#1-deploy-cassandra-to-openshift)
1. [Clone and modify the JanusGraph Docker image](#2-clone-and-modify-the-janusgraph-docker-image)
1. [Deploy the JanusGraph operator](#3-deploy-the-janusgraph-operator)
1. [Load and test retrieval of data using the gremlin console](#3-load-and-test-retrieval-of-data-using-the-gremlin-console)
1. [Test the sizing of JanusGraph using the operator](#4-test-the-sizing-of-janusgraph-using-the-operator)

### 1. Deploy Cassandra to OpenShift

Clone the `cassandra-openshift` locally. In a terminal, run:

```bash
$ git clone https://github.com/IBM/janusgraph-operator.git

$ cd cassandra-openshift
```

You need to update the default configurations of Cassandra so that it can be deployed to OpenShift. The changes are defined in the `Dockerfile`. In order to adapt to the OpenShift environment, you need to change the group ownership and file permission to root. (See [Set group ownership and file permission](https://developer.ibm.com/learningpaths/universal-application-image/design-universal-image/#6-set-group-ownership-and-file-permission) in "[Best practices for designing a universal application image](https://developer.ibm.com/learningpaths/universal-application-image/design-universal-image/).") Although OpenShift runs containers using an arbitrarily assigned user ID, the group ID must always be set to the root group (0). And there are other changes that Cassandra needs for it to be successfully deployed which will not be covered in this tutorial. [CAN READERS FIND THIS INFORMATION SOMEWHERE ELSE?]

You can build and push the Cassandra image to your image repository by running following commands:

```bash
docker build -t cassandra:1.0

docker tag cassandra:1.0 <repository hostname>/<username>/cassandra:1.0

docker push <repository hostname>/<username>/cassandra:1.0

```

_**Note:** You need to change "repository hostname" and "username" accordingly._

After the image is built, you can deploy Cassandra as a `StatefulSet` in OpenShift.

Run the following command to deploy Cassandra from the cloned directory in the terminal:

```bash
$ oc apply -f cassandra-app-v1.yaml -f cassandra-svc-v1.yaml
```

To ensure that Cassandra is running, it should create one instance of the Cassandra database. If you want to have multiple replicas, you can modify replicas in the `cassandra-app-v1.yaml`.

![Cassandra Pod](../images/cassandra-deployment.png)

### 2. Clone and modify the JanusGraph Docker image

The JanusGraph Docker image from the official repo deploys fine into Kubernetes but runs into errors when deployed into OpenShift. There are few things that need to be modified before you can deploy:

* Fork the repo `https://github.com/JanusGraph/janusgraph-docker`.
* [Change the file and group ownership](https://developer.ibm.com/learningpaths/universal-application-image/design-universal-image/#6-set-group-ownership-and-file-permission) to root (0) for related folders. The following modifications apply to the `Dockerfile`:
```bash
chgrp -R 1001:0 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \
chmod -R g+w ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR}

RUN chmod u+x /opt/janusgraph/bin/gremlin.sh
RUN chmod u+x /opt/janusgraph/conf/remote.yaml
```
* Change the `JANUS_PROPS_TEMPLATE` value to `cql` as you will be using Cassandra as the back end.
* Since you will only be using the latest version, change the version to the latest in `build-images.sh`. You will create a copy of that file to `build-images-ibm.sh` and modify it there. These modifications include commenting out a few lines. The following modifications are applied to the build script:

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
* Create `janusgraph-cql-server.properties` in the latest version directory (which in this case is `0.5`) and add the following properties:

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

These are properties that allows JanusGraph to talk to Cassandra as Cassandra will be storing the data in a distributed fashion.

After these changes, make sure to update `janusgraph-cql-server.properties` with the `cluster-ip` of the Cassandra service. Update `storage.hostname` with the `Cluster-IP`.

![Cluster IP](../images/cluster-ip.png)

Now you can build and deploy the JanusGraph Docker image to OpenShift by running:

```bash
$ ./build-images-ibm.sh -- if you have created a new file
```

or...

```bash
$ ./build-images.sh -- if you have modified the file provided by JanusGraph
```

### 3. Deploy the JanusGraph operator

Use the Operator SDK to create the operator project, and you can initialize and create the project structure using the SDK. To make things easier, we have already created a project structure using the SDK. If you want to learn more about Operator SDK and controller code structure, you can go to our operator articles [here](level-1-operator.md).  [LINK TO PREVIOUS TUTORIAL HERE.]

The custom resource (CR) instance and spec definition in your API should look like the following:


<table>
<tr>
<th>Custom Resource (CR)</th>
<th>Spec API definition</th>
</tr>
<tr>
<td>

```yaml
apiVersion: graph.ibm.com/v1alpha1
kind: Janusgraph
metadata:
  name: janusgraph-sample
spec:
  # Add fields here
  size: 3
  version: latest
``` 
</td>
<td>

```go
type JanusgraphSpec struct {}	
	Size    int32  `json:"size"`
	Version string `json:"version"`
}
```

</td>
</tr>
</table>


From the cloned project root directory, open the `build-and-deploy.sh` script in an editor and change following parameters:

```bash
img="<image repo name>/<username>/<image name>:<tag>"
namespace="<namespace>"
```

And finally, run the following from your terminal:

```bash
$ ./build-and-deploy.sh
```

For more information, you can check out the controller code. The operator controller code is responsible for the following tasks:

* Creates the Kubernetes Service that exposes the JanusGraph database with an IP
* Creates the Kubernetes Deployments containing JanusGraph images, configuring them based on the specification in the custom resource 
* Sets the status block in the custom resource to show the readiness of the JanusGraph database

Let's take a look at all of the resources that the operator has deployed for JanusGraph. Run the following command in your terminal:

```bash
$ oc get all
```

The output should look like this:

![Oc get all](../images/oc-get-all.png)


### 4. Load and test retrieval of data using the Gremlin console

To load the data, use your Gremlin console to run the Groovy script [load_data.groovy](https://github.ibm.com/TT-ISV-org/janusgraph-operator/blob/main/data/load_data.groovy). To do so, first, download the [Gremlin console](https://tinkerpop.apache.org/downloads.html) if you haven't already done so.

Once it's downloaded and unzipped, go to `conf/remote.yaml` and update it with the following configuration:

_**NOTE:** `HOST_NAME` is the external IP from your cluster and it can be retrieved using `oc get svc`. Copy the `EXTERNAL-IP` for `jansugraph-sample-service` and replace it._

```yaml
hosts: [HOST_NAME]
port: 8182
serializer: { 
  className: org.apache.tinkerpop.gremlin.driver.ser.GryoMessageSerializerV3d0, config: { serializeResultToString: true }
}
connectionPool: {
  enableSsl: false,
  maxInProcessPerConnection: 16,
  # The maximum number of times that a connection can be borrowed from the pool simultaneously.
  maxSimultaneousUsagePerConnection: 32,
  # The maximum size of a connection pool for a host.
  maxSize: 32,
  maxContentLength: 81928192
}
# Size of the pool for handling background work. default : available processors * 2
workerPoolSize: 16
# Size of the pool for handling request/response operations. # default : available processors
nioPoolSize: 8
```

Copy the Groovy script and paste it into the Gremlin console data directory. Then, from the terminal, run the following from the root of your Gremlin console:

```bash
$ bin/gremlin.sh

$ :remote connect tinkerpop.server conf/remote.yaml
```

Then run the following command to load the Groovy script that you copied and pasted to the data directory:

```bash
$ :load data/load_data.groovy
```

This will load the data. To confirm that the data has been successfully loaded, you should run a Gremlin query to get all the airlines:

```bash
gremlin> g.V().has("object_type", "flight").limit(30000).values("airlines").dedup().toList()
==>MilkyWay Airlines
==>Spartan Airlines
==>Phoenix Airlines
```

You have now successfully loaded your data.

### 5. Test the sizing of JanusGraph using the operator

Next, let's confirm that the operator can successfully scale the JanusGraph database up or down. This can be performed using the OpenShift console.

To open the OpenShift console in IBM cloud, in the IBM Cloud console, navigate to the cluster you provisioned your operator and JanusGraph into, and press the **OpenShift web console** button at the top-right corner of the page.

![OpenShift](../images/openshift-2.png)

In the OpenShift console, select your project in the **Project** combo box along the top of the window. Your project is the namespace that you deployed you operator into.

Then, from the left navigation menu, select **Workloads** and **Stateful Sets**. Click on the one named `janusgraph-sample`.

![stateful set](../images/statefulset.png)

This will bring you to a screen that shows the number of replicas that have been deployed.

![Replicas](../images/replicas.png)

To test the sizing of JanusGraph, you can increase the number of pods by clicking the up arrow next to pods, and decrease it by clicking the down arrow.

After each increment and decrement, you can go to the terminal where you connected to JanusGraph using the Gremlin console from your local machine, and run `get` commands to retrieve the data. On all resizing, you should consistently see the same amount of data retrieved. Run the following query to receive all the airlines with any duplicate data removed:

```bash
$ gremlin> g.V().has("object_type", "flight").limit(30000).values("airlines").dedup().toList()
```

This should consistently retrieve the same data regardless of how many times you've resized. The output should look like this: 

```bash
[
  "Spartan Airlines",
  "Phoenix Airlines",
  "MilkyWay Airline"
]
```

## Conclusion

**Congratulations!** You've now successfully deployed a JanusGraph operator level 1, and you have tested the deployment by resizing the replicas, and checked the integrity of the data in the new pods.
