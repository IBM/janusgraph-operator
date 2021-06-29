# Certifying Janusgraph Image

In this tutorial, we will learn about how to prepare and certify your JanusGraph Operator so that you could deploy Janusgraph Operator to RedHat OpenShift market place or at https://operatorhub.io.

If you want to learn more about operator certification, you can click the link [here](https://github.ibm.com/TT-ISV-org/operator/blob/main/certification/cert-info.md).

# Prerequisites

Follow the prerequisite steps as mentioned in the [Program Prerequisites](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/prerequisites). These prerequisites are part of [Certification Workflow](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/certification-workflow).

Following steps are required to certify JanusGraph operator: 
1. JanusGraph container image certification.
1. JanusGraph operator image certification.
1. JanusGraph operator bundle image certification.
1. Preview in OperatorHub (Optional).

## 1. JanusGraph container image certification

Following are the steps to certify container image. At the end of the steps there is a complete Dockerfile with all the changes required for certification.
### Dockerfile changes for container image certification

The following steps from 1-4 are Dockerfile changes that is required for certification.

1. The base image that is used to build Janusgraph image should be supported by RedHat. In the Janusgraph docker project that you have cloned, find the `Dockerfile` and use the following RedHat supported `OpenJDK` image:

```bash
FROM registry.access.redhat.com/ubi8/openjdk-8:1.3-12
```

> NOTE: There are higher versions of OpenJDK available and can be used as well.

2. The JanusGraph image should run as `non-root` user but part of the `root` group. To do this, we have added the following changes to the existing Janusgraph `Dockerfile`.

* Comment this section out as the user 999 is already part of the base image and since command `apt-get` is not part of the base image we replace that with `dnf`.
```bash
# RUN groupadd -r janusgraph --gid=999 && \
#     useradd -r -g janusgraph --uid=999 -d ${JANUS_DATA_DIR} janusgraph && \
#     apt-get update -y && \
#     DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends krb5-user && \
#     rm -rf /var/lib/apt/lists/*

RUN dnf -y upgrade-minimal --security --sec-severity=Important --sec-severity=Critical && \
    rm -rf /var/lib/apt/lists/*
```

* Change the group of the folders to `root` group.
```bash
chown -R 999:0 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \

chmod -R g+w ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR}

```

3. Add following labels to the Janusgraph Operator Dockerfile. These labels are required labels that will be checked part of certification process.

```bash
  LABEL name="JanusGraph Operator Using Cassandra" \
  maintainer="sanjeev.ghimire@ibm.com" \
  vendor="JanusGraph" \
  version=${JANUS_VERSION} \
  release="1" \
  summary="This is a JanusGraph operator that ensures stateful deployment in an OpenShift cluster." \
  description="This operator will deploy JanusGraph in OpenShift cluster."  
```

4. Copy the licenses folder for JanusGraph Operator to your container:

```bash
  # Required Licenses for Red Hat build service and scanner
COPY licenses /licenses
```

Here is the complete Dockerfile for JanusGraph (container image) with all the changes mentioned above for it to be certified:

```bash
#
# NOTE: THIS FILE IS GENERATED VIA "update.sh"
# DO NOT EDIT IT DIRECTLY; CHANGES WILL BE OVERWRITTEN.
#
# Copyright 2019 JanusGraph Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM debian:buster-slim as builder

ARG JANUS_VERSION=0.5.3
ARG YQ_VERSION=3.4.1

ENV JANUS_VERSION=${JANUS_VERSION} \
    JANUS_HOME=/opt/janusgraph

WORKDIR /opt

RUN apt update -y && apt install -y gpg unzip curl && \
    curl -fSL https://github.com/JanusGraph/janusgraph/releases/download/v${JANUS_VERSION}/janusgraph-${JANUS_VERSION}.zip -o janusgraph.zip && \
    curl -fSL https://github.com/JanusGraph/janusgraph/releases/download/v${JANUS_VERSION}/janusgraph-${JANUS_VERSION}.zip.asc -o janusgraph.zip.asc && \
    curl -fSL https://github.com/JanusGraph/janusgraph/releases/download/v${JANUS_VERSION}/KEYS -o KEYS && \
    curl -fSL https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 -o yq && \
    gpg --import KEYS && \
    gpg --batch --verify janusgraph.zip.asc janusgraph.zip && \
    unzip janusgraph.zip && \
    mv janusgraph-${JANUS_VERSION} /opt/janusgraph && \
    rm -rf ${JANUS_HOME}/elasticsearch && \
    rm -rf ${JANUS_HOME}/javadocs && \
    rm -rf ${JANUS_HOME}/log && \
    rm -rf ${JANUS_HOME}/examples

COPY conf/janusgraph-berkeleyje-lucene-server.properties conf/log4j-server.properties ${JANUS_HOME}/conf/gremlin-server/
COPY conf/janusgraph-cql-server.properties conf/log4j-server.properties ${JANUS_HOME}/conf/gremlin-server/
COPY scripts/remote-connect.groovy ${JANUS_HOME}/scripts/
# COPY conf/gremlin-server.yaml ${JANUS_HOME}/conf/gremlin-server/

## 1. Use the UBI for open jdk
FROM registry.access.redhat.com/ubi8/openjdk-8:1.3-12

ARG CREATED=test
ARG REVISION=test
ARG JANUS_VERSION=0.5.3

## 3. Add following labels to the Janusgraph Operator Dockerfile.
LABEL name="JanusGraph" \
      maintainer="sanjeev.ghimire@ibm.com" \
      vendor="JanusGraph" \
      version=${JANUS_VERSION} \
      release="1" \
      summary="A distributed graph database" \
      description="A distributred graph database"

ENV JANUS_VERSION=${JANUS_VERSION} \
    JANUS_HOME=/opt/janusgraph \
    JANUS_CONFIG_DIR=/etc/opt/janusgraph \
    JANUS_DATA_DIR=/var/lib/janusgraph \
    JANUS_SERVER_TIMEOUT=30 \
    JANUS_STORAGE_TIMEOUT=60 \
    # JANUS_PROPS_TEMPLATE=berkeleyje-lucene \
    JANUS_PROPS_TEMPLATE=cql \
    JANUS_INITDB_DIR=/docker-entrypoint-initdb.d \
    janusgraph.index.search.directory=/var/lib/janusgraph/index \
    janusgraph.storage.directory=/var/lib/janusgraph/data \
    gremlinserver.graphs.graph=/etc/opt/janusgraph/janusgraph.properties \
    gremlinserver.threadPoolWorker=1 \
    gremlinserver.gremlinPool=8 \
    gremlinserver.host=0.0.0.0 \
    gremlinserver.channelizer=org.apache.tinkerpop.gremlin.server.channel.WsAndHttpChannelizer

USER root

## Note that this section is commented out because the base image used already has user 999
# and also the  apt-get command is not available in the base image.
# RUN groupadd -r janusgraph --gid=999 && \
#     useradd -r -g janusgraph --uid=999 -d ${JANUS_DATA_DIR} janusgraph && \
#     apt-get update -y && \
#     DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends krb5-user && \
#     rm -rf /var/lib/apt/lists/*


RUN dnf -y upgrade-minimal --security --sec-severity=Important --sec-severity=Critical && \
    rm -rf /var/lib/apt/lists/*


COPY --from=builder /opt/janusgraph/ /opt/janusgraph/
COPY --from=builder /opt/yq /usr/bin/yq
COPY docker-entrypoint.sh /usr/local/bin/
COPY load-initdb.sh /usr/local/bin/

##4. Copy the licenses folder for JanusGraph Operator to your container:
COPY licenses /licenses

## 2. Change the group of the folders to root group.
RUN chmod 755 /usr/local/bin/docker-entrypoint.sh && \
    chmod 755 /usr/local/bin/load-initdb.sh && \
    chmod 755 /usr/bin/yq && \
    mkdir -p ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \
    chown -R 999:0 ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \    
    chmod -R g+w ${JANUS_HOME} ${JANUS_INITDB_DIR} ${JANUS_CONFIG_DIR} ${JANUS_DATA_DIR} && \
    chmod u+x /opt/janusgraph/bin/gremlin.sh && \
    chmod u+x /opt/janusgraph/conf/remote.yaml

EXPOSE 8182

WORKDIR ${JANUS_HOME}
USER janusgraph

ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "janusgraph" ]

LABEL org.opencontainers.image.title="JanusGraph Docker Image" \
      org.opencontainers.image.description="Official JanusGraph Docker image" \
      org.opencontainers.image.url="https://janusgraph.org/" \
      org.opencontainers.image.documentation="https://docs.janusgraph.org/v0.5/" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.source="https://github.com/JanusGraph/janusgraph-docker/" \
      org.opencontainers.image.vendor="JanusGraph" \
      org.opencontainers.image.version="${JANUS_VERSION}" \
      org.opencontainers.image.created="${CREATED}" \
      org.opencontainers.image.license="Apache-2.0"


```


### Build and Deploy

Before you push your images for scanning, it is recommended to test the container image. To build and deploy the container image, run the following script:

```bash
$ ./build-images-ibm.sh -- if you have created a new file
```
OR

```bash
$ ./build-images.sh -- if you have modified file provided by Janusgraph
```

### Create container application project in RedHat Connect Portal

From the RedHat Partner Connect portal, create the container application project before uploading your image by following below link:

https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-operator/creating-an-operator-project/creating-container-project

Make sure the certification checklist are all completed and you see green check mark.

![container image checklist](../images/ccp-checklist.png)

>NOTE: The sales contact information and distribution approval from Redhat in the checklist items, which can be added later, are optional for container image certification.

Now, you can push your container image for certification. It can be done manually from your local build or configure the build service in `RedHat Connect Portal`.


### Push Container Image manually

You can follow the instructions from the RedHat Connect Portal, by clicking the `Push Images Manually` link from your project page.

![push manually](../images/push-manually-link.png)

which brings to this page:

![push manually](../images/push-manually.png)

>NOTE that the page you see is using podman cli command, but you can use docker instead of podman.

Follow the steps to push container image manually:
* Login to the RedHat Connect Registry: 
`docker login -u unused scan.connect.redhat.com -p <registry key>`
* Find the image id of your container image using: `docker images |grep janusgraph-docker`
* Tag your container image using: 
`docker tag <image id> scan.connect.redhat.com/<OSPID>/janusgraph-docker:<version>` . Make sure to use the correct URL thats given in `Push your image` page from RedHat Connect Portal.
>Replace `image id` and `version` with respective values.
* Push your container image for scanning using: 
`docker push scan.connect.redhat.com/<OSPID>/janusgraph-docker:<version>`
>Replace `version` with the values used when tagging.

>NOTE: You can find your project registry key and ospid by going to your project in RedHat Connect Portal and clicking `Push Images Manually` or you can copy the full URL instead of copying just the OSPID

Then you can see your container image in RedHat Connect Portal scanning for any issues in your image. If no issues found you can see the image `Certification test` as being `passed`.

![Certification pass](../images/cert-pass.png)


## 2. JanusGraph operator image certification

In previous step, you already created a container project and certified the container image. You can use the same project to certify the JanusGraph operator image. Before we create JanusGraph operator image few things needs to be changed:

* Update Dockerfile to use RedHat supported base image in `Dockerfile`

Replace:
```bash
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
```

With

```bash
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
```

* Add the following labels to `Dockerfile` right after base image:

```bash
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

LABEL name="JanusGraph Operator Using Cassandra" \
  maintainer="Sanjeev Ghimire:sanjeev.ghimire@ibm.com" \
  vendor="IBM" \
  version="v0.0.1" \
  release="1" \
  summary="This is a JanusGraph operator that ensures stateful deployment in an OpenShift cluster." \
  description="This operator will deploy JanusGraph in OpenShift cluster."

```

* Add appropriate licenses to `licenses` folder, create licenses folder if you don't have one and update `Dockerfile` to copy the license directory:

```bash
# Use UBI image as base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
LABEL name="JanusGraph Operator Using Cassandra" \
  maintainer="Sanjeev Ghimire:sanjeev.ghimire@ibm.com" \
  vendor="IBM" \
  version="v0.0.1" \
  release="1" \
  summary="This is a JanusGraph operator that ensures stateful deployment in an OpenShift cluster." \
  description="This operator will deploy JanusGraph in OpenShift cluster."

# Required Licenses for Red Hat build service and scanner
COPY licenses /licenses
```

* Make sure to push the operator and test in your cluster by running:

```bash
./build-and-push-operator.sh

./deploy-operator.sh

```

* Upload operator image to RedHat project registry:

First find the image id of the operator image you built in last step. From terminal run: 

```bash
docker images | grep `janusgraph-operator`
```

Then, login to your project

```bash
docker login -u unused scan.connect.redhat.com -p <project registry key>

docker tag <image id> scan.connect.redhat.com/<ospid-id>/janusgraph-operator:v0.0.1

# Push the iamge to the RH registry
$ docker push scan.connect.redhat.com/<ospid-id>/janusgraph-operator:v0.0.1

```

>NOTE: You can find your project registry key and ospid by going to your project in RedHat Connect Portal and clicking `Push Images Manually`

* Go to `https://connect.redhat.com/project/<project_id>/images` to check the status of certification.

It might take a while for the image to appear. You then need to wait for the certification process to finish.

If "certification test" passed then continue to next step. Otherwise, check the scan logs for errors and update the image accordingly.

## 3. JanusGraph operator bundle image certification

### Steps to create operator bundle and submit for certification

1. Create an operator bundle image project using the following link:
https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-operator/certify-your-operator-bundle-image/creating-operator-bundle-image-project

2. Make sure the certification checklist are all completed and you see green check mark.

![container image checklist](../images/ccp-checklist.png)

>NOTE: The sales contact information and distribution approval from Redhat in the checklist items, which can be added later, are optional for container image certification.

3. Create the bundle by running the following script:
Update the following before running the script:

```bash
REGISTRY=docker.io
USERNAME=sanjeevghimire
REPOSITORY=janusgraph-operator
VERSION=1.0.13
```
Then run:

```bash
./make-bundle.sh
```

which asks series of questions, whose answers will be added to the generated ClusterServiceVersion (CSV). The questions are:

```bash
* Display name for the operator (required):
* Description for the operator (required):
* Provider's name for the operator (required):
* Any relevant URL for the provider name (optional):
* Comma-separated list of keywords for your operator (required):
* Comma-separated list of maintainers and their emails (e.g. 'name1:email1, name2:email2') (required):

```

The make bundle command used above, automates several tasks, including running the following operator-sdk subcommands in order:

```bash
generate kustomize manifests

generate bundle

bundle validate
```

Once this process is done, you will see a folder `bundle` in your project root directory which contains CSV, copy of CRDs, and generated metadata in the bundle format.

To learn in detail about the bundle process you can go to this [link](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/creating-the-metadata-bundle).

4. Make sure following files are updated: 
Update `bundle/manifests/janusgraph-operator.clusterserviceversion.yaml` by:

Replacing: 

`base64data: ""` with a base64 image of JanusGraph logo.

`mediatype: ""` With: `mediatype: "image/png"`


Add the following line to `bundle/metadata/annotations.yaml`:

`operators.operatorframework.io.bundle.channel.default.v1: alpha`


Add the following labels to bundle.Dockerfile:
```yaml
LABEL operators.operatorframework.io.bundle.channel.default.v1=alpha
LABEL com.redhat.openshift.versions="v4.6"
LABEL com.redhat.delivery.operator.bundle=true
```

5. To build the bundle image and push to registry, run the following script, make sure to replace the following with appropriate values:
```bash
export USERNAME="sanjeevghimire"
export BUNDLE_IMG="docker.io/$USERNAME/janusgraph-operator-bundle:v0.0.1"
```

```bash
./build-and-push-operator-bundle.sh

```
6. Pushing Operator bundle image for Certification test

### Push Operator bundle image manually

You can follow the instructions from the RedHat Connect Portal, by clicking the `Push Images Manually` link from your bundle project page.

![push manually](../images/push-manually-link.png)

which brings to this page:

![push manually](../images/push-manually.png)

>NOTE that the page you see is using podman cli command, but you can use docker instead of podman.

Follow the steps to push container image manually:
* Login to the RedHat Connect Registry: 
`docker login -u unused scan.connect.redhat.com -p <registry key>`
* Find the image id of your bundle image using: `docker images |grep janusgraph-operator-bundle`
* Tag your container image using: 
`docker tag <image id> scan.connect.redhat.com/<OSPID>/janusgraph-operator-bundle:<version>` . Make sure to use the correct URL thats given in `Push your image` page from RedHat Connect Portal.
* Push your container image for scanning using: 
`docker push scan.connect.redhat.com/<OSPID>/janusgraph-operator-bundle:<version>`


>NOTE: You can find your project registry key and ospid by going to your project in RedHat Connect Portal and clicking `Push Images Manually` or you can copy the full URL instead of copying just the OSPID

Then you can see your images in RedHat Connect Portal scanning for any issues in your image. If no issues found, you can see the image `Certification test` as being `passed`.

![Bundle Certification](../images/bundle-certification.png)


### 4. Preview in OperatorHub (Optional)

Go to the preview link: https://operatorhub.io/preview and paste the content of [janusgraph-operator.clusterserviceversion.yaml](../bundle/manifests/janusgraph-operator.clusterserviceversion.yaml) and you should see the following preview:

![CSV yaml preview](../images/csvyaml-preview.png)

![Operator hub preview](../images/operatorhub-preview.png)

Finally, you have successfully certified your operator. Now you can publish your operator in RedHat market place and operatorshub.io.

## Next Steps

At this point, you have successfuly certified your operator image and bundle. The next step is to publish the JanusGraph operator at RedHat Market place or operatorhubs.io.