# Certifying Janusgraph Image

In this tutorial, we will learn about how to prepare and certify your JanusGraph container so that you could deploy Janusgraph Operator to RedHat OpenShift.

# Prerequisites

* Follow the prerequisite steps as mentioned in the [Program Prerequisites](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/prerequisites). These prerequisites are part of [Certification Workflow](https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/program-on-boarding/certification-workflow).


The certification of an operator is done in 2 stages as follows: 
1. Operator image container certification
2. Operator bundle image certification

## 1. Operator Image Container Certification
### Steps

1. The base image that is used to build Janusgraph image should be supported by RedHat. In the Janusgraph docker project that you have cloned, find the `Dockerfile` and use the following RedHat supported `OpenJDK` image:

```bash
FROM registry.access.redhat.com/ubi8/openjdk-8:1.3-12
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

5. Build and deploy the operator image by running the following script:

```bash
./build-and-deploy.sh
```

6. From the RedHat Partner Connect portal, create the container application project by following the link:

https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-operator/creating-an-operator-project/creating-container-project

7. Make sure the certification checklist are all completed and you see green check mark.

![container image checklist](../images/ccp-checklist.png)

>NOTE: The sales contact information and distribution approval from Redhat in the checklist items, which can be added later, are optional for container image certification.

8. Now, you can push your container image for certification. It can be done manually from your local build or configure the build service in `RedHat Connect Portal`.

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
`docker tag <image id> scan.connect.redhat.com/ospid-f7cc59be-157e-48ca-a817-7dc43c616c41/janusgraph-docker:1.0.6` . Make sure to use the correct URL thats given in `Push your image` page from RedHat Connect Portal.
* Push your container image for scanning using: 
`docker push scan.connect.redhat.com/ospid-f7cc59be-157e-48ca-a817-7dc43c616c41/janusgraph-docker:1.0.6`

Then you can see your images in RedHat Connect Portal scanning for any issues in your image. If no issues found you can see the image `Certification test` as being `passed`.

![Certification pass](../images/cert-pass.png)


## 2. Operator bundle image certification

### Steps

1. Create an operator bundle image project using teh following link:
https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-operator/certify-your-operator-bundle-image/creating-operator-bundle-image-project

2. Make sure the certification checklist are all completed and you see green check mark.

![container image checklist](../images/ccp-checklist.png)

>NOTE: The sales contact information and distribution approval from Redhat in the checklist items, which can be added later, are optional for container image certification.

3. You must create metadata for your operator as part of the build process. These metadata files are the packaging format used by the Operator Lifecycle Manager (OLM) to deploy your operator onto OpenShift (OLM comes pre-installed in OpenShift 4.x).

* Change the CRD to use v1beta1: The operator-sdk uses the latest version of CustomResourceDefinition, v1, by default. Older versions of OpenShift only support v1beta1, so if your operator is going to be listed on OCP 4.5 and earlier, you'll need to convert to the older format.

```bash
$ vi config/crd/bases/<your CRD filename>
```
and change line
`apiVersion: apiextensions.k8s.io/v1` 
to
`apiVersion: apiextensions.k8s.io/v1beta1`

* Create the bundle: SDK projects are scaffolded with a Makefile containing the bundle recipe by default, which wraps generate kustomize manifests, generate bundle, and other related commands. Run: 

```bash
make bundle
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

Once this process is done, you will see a folder `bundle` in your project root directory which contains CSV, copy of CRDs, and generated metadata in the bundle format.

To learn in detail about the bundle process you can go to this [link](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/creating-the-metadata-bundle).

4. The generated ClusterServiceVersion (CSV) file needs some additional information as below:

Fields to add under metadata.annotations are: 

* `categories` - Comma separated string of these applicable category names 
* `description` - Short description of the operator
* `containerImage` - The full location (registry, repository, name and tag) of the operator image
* `createdAt` - A rough (to the day) timestamp of when the operator image was created
* `support` - Name of the supporting vendor (eg: ExampleCo)
* `repository` -  URL of the operator's source code repository (this field is optional)

Fields to adjust under spec are:  

* `description` - Long description of the operator's owned customresourcedefinitions in Markdown format. Usage instructions and relevant info for the user goes here
* `icon.base64data` - A base64 encoded PNG, JPEG or SVG image will need to be added
* `icon.mediatype` - The corresponding MIME type of the image (eg: image/png)

For the final changes for Janusgraph Operator ClusterServiceVersion you can check [janusgraph-operator.clusterserviceversion.yaml](../bundle/manifests/janusgraph-operator.clusterserviceversion.yaml)


5. Verifying your metadata bundle can be done using `Operator SDK`. Run the following:

```bash
$ operator-sdk bundle validate ./bundle

INFO[0001] Found annotations file                        bundle-dir=bundle container-tool=docker
INFO[0001] Could not find optional dependencies file     bundle-dir=bundle container-tool=docker
INFO[0001] All validation tests have completed successfully
```

```bash
$ operator-sdk bundle validate ./bundle --select-optional suite=operatorframework

INFO[0000] Found annotations file                        bundle-dir=bundle container-tool=docker
INFO[0000] Could not find optional dependencies file     bundle-dir=bundle container-tool=docker
WARN[0000] Warning: Value : (janusgraph-operator-v1.0.2) csv.metadata.Name janusgraph-operator-v1.0.2 is not following the recommended naming convention: <operator-name>.v<semver> e.g. memcached-operator.v0.0.1
INFO[0000] All validation tests have completed successfully

```

6. Previewing your CSV on OperatorHub.io

Go to the preview link: https://operatorhub.io/preview and paste the content of [janusgraph-operator.clusterserviceversion.yaml](../bundle/manifests/janusgraph-operator.clusterserviceversion.yaml) and you should see the following preview:

![CSV yaml preview](../images/csvyaml-preview.png)

![Operator hub preview](../images/operatorhub-preview.png)

7. Pushing Operator bundle image for Certification test

### Push Operator bundle image manually

You can follow the instructions from the RedHat Connect Portal, by clicking the `Push Images Manually` link from your bundle project page.

![push manually](../images/push-manually-link.png)

which brings to this page:

![push manually](../images/push-manually.png)

>NOTE that the page you see is using podman cli command, but you can use docker instead of podman.

Follow the steps to push container image manually:
* Login to the RedHat Connect Registry: 
`docker login -u unused scan.connect.redhat.com -p <registry key>`
* Find the image id of your container image using: `docker images |grep janusgraph-operator`
* Tag your container image using: 
`docker tag <image id> scan.connect.redhat.com/ospid-f7cc59be-157e-48ca-a817-7dc43c616c41/janusgraph-operator:1.0.6` . Make sure to use the correct URL thats given in `Push your image` page from RedHat Connect Portal.
* Push your container image for scanning using: 
`docker push scan.connect.redhat.com/ospid-f7cc59be-157e-48ca-a817-7dc43c616c41/janusgraph-operator:1.0.6`

Then you can see your images in RedHat Connect Portal scanning for any issues in your image. If no issues found, you can see the image `Certification test` as being `passed`.

![Certification pass](../images/cert-pass.png)





