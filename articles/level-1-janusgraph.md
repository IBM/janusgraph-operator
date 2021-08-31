# Develop and deploy a level 1 JanusGraph operator using Apache Cassandra

## SUBTITLE: Deploy an instance of the JanusGraph database simply by creating a custom resource

This tutorial shows you how to develop and deploy a level 1 operator on the Red Hat OpenShift Container Platform. You will create an operator for JanusGraph that uses [Apache Cassandra](https://cassandra.apache.org/) as a storage back end. Cassandra is a distributed database platform that can scale and be highly available, and can perform really well on any commodity hardware or cloud infrastructure.

When you have completed this tutorial, you will understand how to:

* Deploy Cassandra as back-end storage.
* Create a JanusGraph image that runs well in OpenShift, not just Kubernetes.
* Deploy a JanusGraph operator to an OpenShift cluster.
* Scale a JanusGraph instance up or down by modifying and applying the custom resource (CR) to an OpenShift cluster.

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
1. [Clone and modify Janusgraph docker image](#2-clone-and-modify-janusgraph-docker-image)
1. [Deploy Janusgraph operator](#3-deploy-janusgraph-operator)
1. [Load and test retrieve of data using gremlin console](#4-load-and-test-retrieve-of-data-using-gremlin-console)
1. [Scaling Janusgraph](#5-scaling-janusgraph)

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

The JanusGraph Docker image from the official repo deploys fine into Kubernetes but runs into errors when deployed into OpenShift. There are few things that need to be modified before you can deploy.

Follow this [link](https://github.com/IBM/janusgraph-docker-openshift/blob/main/README-openshift.md) to create an JanusGraph image that can be deployable to OpenShift.

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

To load the data, use your Gremlin console to run the Groovy script [load_data.groovy](https://github.com/IBM/janusgraph-operator/blob/main/data/load_data.groovy). To do so, first, download the [Gremlin console](https://tinkerpop.apache.org/downloads.html) if you haven't already done so.

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

To retrieve the data and test to make sure the data has been successfully loaded, we will run a gremlin query to get all the airlines:

```bash
gremlin> g.V().has("object_type", "flight").limit(30000).values("airlines").dedup().toList()
==>MilkyWay Airlines
==>Spartan Airlines
==>Phoenix Airlines
```

You have now successfully loaded your data.

In the next section, we will scale the JanusGraph instance by changing the number of pod replicas. As you do so, rerun this Gremlin query to show that the set of data in the database remains the same, that the starting or stopping pod replicas does not duplicate or lose data.

### 5. Scaling JanusGraph

The JanusGraph instance can scale to run more pods to handle more client load and spread it across more cluster nodes. However, scaling is adjusted differently when an operator is managing an instance. We'll look at two approaches a developer can use to scale a set of pods:

* First, we'll look at how a developer can typically scale set of pods manually, and see how that doesn't work quite the same with an operator.
* Second, we'll look at how a developer can use an operator to scale a set of pods that the operator is managing.

#### Manually adjust the number of pod replicas

From your provisioned cluster which which you have already setup part of prerequisites, select the cluster and go to `OpenShift web console` by clicking the button from top right corner of the page.

![OpenShift](../images/openshift-2.png)

In the OpenShift console, select your project in the **Project** combo box along the top of the window. Your project is the namespace that you deployed you operator into.

Then, from the left navigation menu, select **Workloads** and **Stateful Sets**. Click on the one named `janusgraph-sample`.

![stateful set](../images/statefulset.png)

This will bring you to a screen that shows the number of replicas that have been deployed.

![Replicas](../images/replicas.png)

Typically, a developer can use this view to manually change the number of pod replicas, but this works a bit differently in a Deployment or StatefulSet being managed by an operator. In this view, you can use the up and down arrows next to the set of pods to increase or decrease the number of pods. Indeed, if you try that here, the view shows that the number of replicas does change. But wait a minute and the number changes back to 3 again. Why? Because this resource is being managed by the JanusGraph operator, and its CR says that the size is 3. So when the size differs from 3, the operator puts it back.

To adjust the number of pod replicas and have the change stick, we'll need to use the operator.

#### Use the operator to adjust the number of pod replicas

To tell the operator to adjust the number of pod replicas, change that setting in the custom resource. Because the CR describes the instance's configuration, changing the settings in the CR causes the operator to change the configuration in the instance.

To scale the number of pod replicas in the JanusGraph instance, change the `spec` in your custom resource instance. Change the `Size` in the following spec:

```bash
  apiVersion: graph.example.com/v1alpha1
  kind: Janusgraph
  metadata:
    name: janusgraph-sample
  spec:
    # update the size to scale/descale Janusgraph instances
    size: 3
    version: 1.0.1
```

And apply to the cluster using:

```bash
oc apply -f samples/graph_v1alpha1_janusgraph.yaml
```

In the view of the stateful set in the OpenShift console, watch the number of pod replicas. After a minute, the number will adjust to the new size you specified in the CR. This is because the operator saw the new size and made the necessary adjustments to the instance it's managing.

**Congratulations!** You've successfully deployed an Janusgraph operator `Level I`. And you also have tested the deployment with resizing the replicas and  checked the integrity of the data in new pods.

# License

This code pattern is licensed under the Apache Software License, Version 2.  Separate third party code objects invoked within this code pattern are licensed by their respective providers pursuant to their own separate licenses. Contributions are subject to the [Developer Certificate of Origin, Version 1.1 (DCO)](https://developercertificate.org/) and the [Apache Software License, Version 2](https://www.apache.org/licenses/LICENSE-2.0.txt).
