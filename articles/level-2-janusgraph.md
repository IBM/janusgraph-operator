# LEVEL 2 : Seamless upgrade for JanusGraph Operator

In this tutorial, we will learn how to develop and deploy a Level 2 operator on the OpenShift Container Platform. This is the continuation of the level 1 operator where it has the functionality to upgrade version of your operand. In this case our operand is JanusGraph.

When the reader has completed this tutorial, they will understand how to:
* Build and push a latest version of Janusgraph image
* Upgrade version of the operand (JanusGraph application)


## What is Operators Level II Capability?

This capabillity is referred as seamless upgrade. The operator should support seamless upgrade of the operator and the operand. An upgrade of the operator means that the CR instance are in new desired state and would upgrade the operand. Upgrade might also mean upgrading the application that operator manages along with other internals such as schema migrations. It should clearly mentioned what is upgraded when this takes place and what is not. 

To upgrade the current version of the operand, which in this case is Janusgraph, to the desired version using go, you need to:
* Check for current version of the container image.
* Compare the version of the Custom Resource (CR) instance with the container image.
* If the version is not the same and is lower than the one in CR, then update the image of the container. If the version is lower, ignore the version upgrade.

## Steps

1. [Implement version upgrade in the operator controller](#1_implement_version_upgrade_in_the_controller)
1. [Deploy the operator](#2_deploy_the_operator)
1. [Update CR version and apply](#2_update_cr_version)


### 1. Implement version upgrade in the operator controller

The operator should be capable of upgrading the version to the desired version higher than the current one. To do so, following cases should be implemented in the operator's controller.

1. Get the desired version from the Custom Resource (CR)
1. Get the current version from the container specification
1. Check to see if the container image needs upgrade
1. Upgrade container image if current version is less than desired version

Lets look at the code implementation.
>Note: check the comments to see what the code is doing.

```go
  // version upgrade/downgrade
  //1. Get the desired version from the Custom Resource (CR)
	desiredVersion := janusgraph.Spec.Version	
	crImage := fmt.Sprintf("%s:%s", JANUS_IMAGE, desiredVersion)
  //2. Get the current version from the container specification
  manifestImage := *&found.Spec.Template.Spec.Containers[0].Image
	versionSplit := strings.Split(manifestImage, ":")
	currentVersion := versionSplit[1]
	log.Info("Current version is", ":", currentVersion)

  //3. Check to see if the container image needs upgrade
	// only version with  x.x.x format will work
	// ex: 1.1.1 or 1.1 or 2
	vA, err := semver.NewVersion(desiredVersion)
	if err != nil {
		fmt.Println(err.Error())
	}
	log.Info("Desired Version ", ":", vA.String())
	vB, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Println(err.Error())
	}
  //4. Upgrade container image if current version is less than desired version
	if vB.LessThan(*vA) {
		log.Info("Uprading to new version", ":", vA.String())
		found.Spec.Template.Spec.Containers[0].Image = crImage
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update version")
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else {
		log.Info("Not Upgrading : ", "Already at latest version", currentVersion)
	}

```

### 2. Deploy the operator

After the code has been changed, you can deploy the operator to your cluster using previous steps described in [here](level-1-janusgraph.md)

### 3. Update CR version and apply

To upgrade the version, the CR instance should contain the version that you are upgrading to. Modify the `version` attribute in `spec`.

```yaml
apiVersion: graph.example.com/v1alpha1
kind: Janusgraph
metadata:
  name: janusgraph-sample
spec:
  # Add fields here
  size: 3
  version: 1.0.1

```

Once the version is changed to the desired version, then you can apply the CR to your cluster.

```bash
oc apply -f config/samples/graph_v1alpha1_janusgraph.yaml

```

### 4. Test the version upgrade

There are multiple ways to test if the version of the operand has changed.

* Test your application and make sure the feature changes has been applied.
* Check your pods to see if the container image has latest version. You can do that describing the pods and check `Containers.janusgraph.Image` and make sure it has the right image and version.

```bash
oc describe <pod name>

```

```bash
Containers:
  janusgraph:
    Container ID:   cri-o://f6d4584e850308d988f13cfdf76ab89b4da77687763c64e6640b50b0bcc8ae4b
    Image:          sanjeevghimire/janusgraph:1.0.1
    Image ID:       docker.io/sanjeevghimire/janusgraph@sha256:609ed2aa1c802f4ea377855f0add33e024572c0ddfa728dcded24f0f1164eaa8
    Port:           8182/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Wed, 09 Jun 2021 12:04:17 -0700


```


At this point, you have successfully implemented Level II operator for JanusGraph.                  