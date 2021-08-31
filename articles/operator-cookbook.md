# The Operator Cookbook: How to make an operator

## Operators ensure that certain Kubernetes resources are created and configured properly, and relay status information back to the user

This tutorial examines common building blocks for level 1 operators, and shows you the logic a service vendor needs to write in order
to build a level 1 operator. It uses the [Operator Capability Levels](https://operatorframework.io/operator-capabilities/) as a guideline for what is considered a level 1 operator.

By developing and deploying the [Memcached Operator](https://developer.ibm.com/learningpaths/kubernetes-operators/develop-deploy-simple-operator/create-operator/) 
and the [JanusGraph Operator](https://github.com/IBM/create-and-deploy-memcached-operator-using-go/blob/main/articles/level-1-operator.md) [**NEED TO UPDATE THIS LINK**], you can 
analyze the similarities in the controller code and think about what each operator must do at a high level.

## Characteristics of an operator

An operator ensures that certain Kubernetes resources (the ones that are required to run your service) are created and configured 
properly. It also relays status information back to the user to communicate when certain resources are running.

The Memcached example shows you how to create a Deployment resource for the manager, which is the operator itself. And then, once you have  
applied your custom resource (CR) using `kubectl`, you can create a Memcached Deployment, which is the operand, or the application you  
are deploying. Similarly, in the JanusGraph operator, you create a StatefulSet instead of a Deployment, and then create a service. 

Here are the main characteristics of a level 1 operator covered in this tutorial:

1. [Define the API](#define-the-api)
2. [Create Kubernetes resources if they do not exist](#1-check-if-a-resource-exists-and-create-one-if-it-does-not)
3. [Update replicas in your controller code](#2-replicas-should-be-set-in-the-custom-resource-and-updated-in-the-controller-code)
4. [Update the status](#3-update-the-status)
5. [Scale up and down via custom resource](#4-ensure-that-the-operator-can-scale-up-and-down-via-the-custom-resource)

## Define the API

When building an operator, the easiest way to get started is by using the [Operator SDK](https://sdk.operatorframework.io/). Once you've 
finished the first steps, such as using the [`operator sdk init`](https://github.com/IBM/create-and-deploy-memcached-operator-using-go/blob/main/BEGINNER_TUTORIAL.md#1-create-a-new-project-using-operator-sdk) and [`operator sdk create api`](https://github.com/IBM/create-and-deploy-memcached-operator-using-go/blob/main/BEGINNER_TUTORIAL.md#2-create-api-and-custom-controller) [**SHOULD THIS LINK TO A PREVIOUSLY PUBLISHED TUTORIAL?**] commands, you'll want to update the API.

This is where you design the structure of your custom resource. For simple cases, you'll likely use something like the `Size` and `Version` fields 
in the `Spec` section of your custom resource.

The Operator SDK generates the following code for your API:

```go
package v1alpha1
import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
type ExampleSpec struct {
	Foo string `json:"foo,omitempty"`
}
type ExampleStatus struct {
}
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Example struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MemcachedSpec   `json:"spec,omitempty"`
	Status MemcachedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Example `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Example{}, &ExampleList{})
}
```

First, you should update the `Spec` section, like so:

```go
// ExampleSpec defines the desired state of Example database
type ExampleSpec struct {
	Size    int32  `json:"size"`
	Version string `json:"version"`
}
```

Next, update the `Status` section: 

```go
// ExampleStatus defines the observed state of Example database
type ExampleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}
```

And finally, you need to specify the fields for your `Example` custom resource:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Example is the Schema for the example API
type Example struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleSpec   `json:"spec,omitempty"`
	Status ExampleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExampleList contains a list of Example
type ExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Example `json:"items"`
}
```

That's it for your API.

## The main logic for your operator

The main logic when checking for different types of Kubernetes resources (such as Service, StatefulSet, and Deployment) is as follows.

First, you need to get a reference to a certain type of Kubernetes resource that you want to create:

```go	
found := &appsv1.Deployment{}
```

Then, use the `Get` function to find resources of that type in your namespace:

```go	
err = r.Get(ctx, req.NamespacedName, found)
```

The main logic is shown below, and this is similar no matter what resource you want to ensure is running (Deployment, StatefulSet, Service, or PVC).

### 1. Check if a resource exists, and create one if it does not

First, you should confirm that the error is not nil. If there is no error, that implies that the resource you want to create is already created, so you do not 
need to create another one.

Next, check for an `IsNotFound` error, which indicates that this resource doesn't exist at all -- in which case, you need to create one. 

After you create it, and the deployment or StatefulSet has been created successfully, then you can return and `Requeue`. Otherwise, you need to
return an error:

```go

	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMemcached(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

```
 

### 2. Replicas should be set in the custom resource and updated in the controller code

From the memcached example, you can see that you set a variable to be what the `size` is from the custom resource. From there, you need to 
check if the deployment's spec section has the same number of replicas as what is specified in the custom resource. 
If the numbers don't match, then you need to update the replicas. 

```go
// Ensure the deployment size is the same as the spec
size := memcached.Spec.Size
if *found.Spec.Replicas != size {
	found.Spec.Replicas = &size
	err = r.Update(ctx, found)
	if err != nil {
		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, err
	}
	// Spec updated - return and requeue
	return ctrl.Result{Requeue: true}, nil
}
```

### 3. Update the status

The last thing you need to do in any operator is to update the status. This can be done by using the
reconciler [`Status().Update()`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/client#StatusWriter.Update) function. You will see 
this below, but first you need to format your pods such that you can quickly compare the current state with the desired 
state. In this example, you use the reconciler `List` function to retrieve the pods that match your labels and are in your namespace. 

```go
podList := &corev1.PodList{}
listOpts := []client.ListOption{
	client.InNamespace(memcached.Namespace),
	client.MatchingLabels(labelsForMemcached(memcached.Name)),
}
if err = r.List(ctx, podList, listOpts...); err != nil {
	log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
	return ctrl.Result{}, err
}
```

You then have a function that returns an array of strings. This is done so that you can easily compare pods in the current state 
versus the desired state.

```go
podNames := getPodNames(podList.Items)
```

Next, you update the status if necessary. Check if the pod names you've retrieved from the `List` function match the 
custom resource's status. The status that is defined in the API is as follows: 

```go
Nodes []string `json:"nodes"`
```

These nodes are the pod names that are currently in the cluster. If those nodes are different than the ones you've retrieved 
from the `List` function, then you should update the status using the reconciler [`Status().Update()`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/client#StatusWriter.Update) function. 

```
// Update status.Nodes if needed
if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
	memcached.Status.Nodes = podNames
	err := r.Status().Update(ctx, memcached)
	if err != nil {
		log.Error(err, "Failed to update Memcached status")
		return ctrl.Result{}, err
	}
}
```

Once you've updated the status, you are ready to test your operator!

###4. Ensure that the operator can scale up and down via the custom resource

The last thing to check is to make sure that you can scale your operand up and down via the custom resource. 

You can do this by changing the `size` value in your custom resource:

```go
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
spec:
  size: 3
```

Change the `size` from 3 to 1:

```go
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
spec:
  size: 1
```

Once you issue a `kubectl apply -f` command on the custom resource, you should see two pods terminating. As long as your 
application continues to work and is able to scale up and down via the custom resource, then you have a properly working level 1 
operator.

## Conclusion

Let's recap. To build an operator for your Kubernetes service, you need to complete three main tasks:

1. Implement functions to check if the desired resource exists, and then create it if it does not exist.
2. Set your replicas in your custom resource, and update them within your controller code. 
3. Update your status, which communicates to the user what state the pods are in.

After that, you want to test your operator by scaling it up and down via the custom resource. If it can scale up and down 
successfully via custom resource and your application still runs smoothly, then you are done in terms of a level 1 operator.

**Congratulations!** You now understand the main concepts behind building a level 1 operator. Stay tuned for subsequent tutorials 
that cover level 2 operators.
