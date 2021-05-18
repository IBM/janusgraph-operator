# Certifying Janusgraph Image

In this tutorial, we will learn about how to prepare and certify your JanusGraph containter so that you could deploy Janusgraph Operator to RedHat OpenShift.

# Prerequisites

* Follow the prerequisite steps as mentioned in the [Program Prerequisites](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/prerequisites). These prerequisites are part of [Certification Workflow](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/certification-workflow).


## Steps

1. The base image that is used to build Janusgraph image should be supported by RedHat. In the Janusgraph project that you have cloned, find the `Dockerfile` and use the following RedHat supported `OpenJDK` image:

```bash
FROM registry.access.redhat.com/ubi8/openjdk-8:1.3-9.1617297653
```

> NOTE: There are higher versions of OpenJDK available and can be used as well.

2. The JanusGraph image should run as `non-root` user but part of the `root` group. To do this, we have added the following changes to the existing Janusgraph `Dockerfile`

* Add a non-root user `9999` and assign that user the folders
```bash
RUN groupadd -r janusgraph --gid=9999 && \
    useradd -r -g janusgraph --uid=9999 -d ${JANUS_DATA_DIR} janusgraph && \
```

* Change the group of the folders to `root` group.
```bash
chown -R 9999:9999 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \

chgrp -R 0 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \

chmod -R g+w ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR}

```

3. Add following labels to the Janusgraph Operator Dockerfile.

```bash
LABEL name="JanusGraph Operator Using Cassandra" \
  vendor="IBM" \
  version="v0.0.1" \
  release="1" \
  summary="This is a JanusGraph operator that ensures stateful deployment in an OpenShift cluster." \
  description="This operator will deploy JanusGraph in OpenShift cluster."
```

4. Copy the licenses folder for JanusGraph Operator to your container:

```bash
  # Required Licenses for Red Hat build service and scanner
COPY licenses /licenses
```

5. Build and deploy the operator image by running the following script:

```bash
./build-and-deploy.sh
```

6. From the RedHat Partner Connect portal, create the container application project by following the link:

https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-application/creating-a-container-application-project

7. 