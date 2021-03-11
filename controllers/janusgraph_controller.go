/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.ibm.com/TT-ISV-org/janusgraph-operator/api/v1alpha1"
	graphv1alpha1 "github.ibm.com/TT-ISV-org/janusgraph-operator/api/v1alpha1"
)

// JanusgraphReconciler reconciles a Janusgraph object
type JanusgraphReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=graph.ibm.com,resources=janusgraphs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=graph.ibm.com,resources=janusgraphs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=graph.ibm.com,resources=janusgraphs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=pods;deployments;statefulsets;services;persistentvolumeclaims;persistentvolumes;,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;services;persistentvolumeclaims;persistentvolumes;,verbs=get;list;create;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Janusgraph object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *JanusgraphReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("janusgraph", req.NamespacedName)

	// Fetch the Janusgraph instance
	janusgraph := &graphv1alpha1.Janusgraph{}
	err := r.Get(ctx, req.NamespacedName, janusgraph)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Janusgraph resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Janusgraph")
		return ctrl.Result{}, err
	}

	serviceFound := &corev1.Service{}
	log.Info("Checking for service")
	err = r.Get(ctx, types.NamespacedName{Name: janusgraph.Name + "-service", Namespace: janusgraph.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		srv := r.serviceForJanusgraph(janusgraph)
		log.Info("Creating a new headless service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
		err = r.Create(ctx, srv)
		if err != nil {
			log.Error(err, "Failed to create new service", "service.Namespace", srv.Namespace, "service.Name", srv.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		log.Info("Janusgraph service created, requeuing")
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get service")
		return ctrl.Result{}, err
	}

	/*
			// persistent volume
			pvFound := &corev1.PersistentVolume{}
			err = r.Get(ctx, types.NamespacedName{Name: janusgraph.Name + "-pv", Namespace: janusgraph.Namespace}, pvFound)
			if err != nil && errors.IsNotFound(err) {
				pv := r.pvForJanusgraph(janusgraph)
				log.Info("Creating a new pv", "pv.Namespace", pv.Namespace, "pv.Name", pv.Name)
				err = r.Create(ctx, pv)
				if err != nil {
					log.Error(err, "Failed to create new pv", "pv.Namespace", pv.Namespace, "pv.Name", pv.Name)
					return ctrl.Result{}, err
				}
				// Deployment created successfully - return and requeue
				log.Info("Janusgraph persistent volume created, requeuing")
				return ctrl.Result{Requeue: true}, nil
			} else if err != nil {
				log.Error(err, "Failed to get pv")
				return ctrl.Result{}, err
			}

		// persistent volume claim
		pvcFound := &corev1.PersistentVolumeClaim{}
		err = r.Get(ctx, types.NamespacedName{Name: janusgraph.Name + "-pvc", Namespace: janusgraph.Namespace}, pvcFound)
		if err != nil && errors.IsNotFound(err) {
			pvc := r.pvcForJanusgraph(janusgraph)
			log.Info("Creating a new pvc", "pvc.Namespace", pvc.Namespace, "pvc.Name", pvc.Name)
			err = r.Create(ctx, pvc)
			if err != nil {
				log.Error(err, "Failed to create new pvc", "pvc.Namespace", pvc.Namespace, "pvc.Name", pvc.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			log.Info("Janusgraph persistent volume claim created, requeuing")
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get pvc")
			return ctrl.Result{}, err
		}
	*/

	// deployment
	found := &appsv1.StatefulSet{}
	// Check if the deployment already exists, if not create a new one
	err = r.Get(ctx, types.NamespacedName{Name: janusgraph.Name, Namespace: janusgraph.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForJanusgraph(janusgraph)
		log.Info("Creating a new Statefulset", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		log.Info("Deployment created, requeuing")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JanusgraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&graphv1alpha1.Janusgraph{}).
		Complete(r)
}

func labelsForJanusgraph(name string) map[string]string {
	return map[string]string{"app": "Janusgraph", "janusgraph_cr": name}
}

func (r *JanusgraphReconciler) serviceForJanusgraph(m *v1alpha1.Janusgraph) *corev1.Service {
	ls := labelsForJanusgraph(m.Name)
	srv := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-service",
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			// ClusterIP: corev1.ClusterIPNone, //"None",
			Ports: []corev1.ServicePort{{
				Port: 8182,
				Name: "janusgraph",
			},
			},
			Selector: ls,
		},
	}
	ctrl.SetControllerReference(m, srv, r.Scheme)
	return srv
}

func (r *JanusgraphReconciler) pvcForJanusgraph(m *v1alpha1.Janusgraph) *corev1.PersistentVolumeClaim {
	ls := labelsForJanusgraph(m.Name)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-pvc",
			Labels:    ls,
			Namespace: m.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}

	return pvc
}

func (r *JanusgraphReconciler) pvForJanusgraph(m *v1alpha1.Janusgraph) *corev1.PersistentVolume {
	ls := labelsForJanusgraph(m.Name)
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-pv",
			Labels:    ls,
			Namespace: m.Namespace,
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("5Gi"),
			},
		},
	}

	return pv
}

func (r *JanusgraphReconciler) deploymentForJanusgraph(m *v1alpha1.Janusgraph) *appsv1.StatefulSet {
	ls := labelsForJanusgraph(m.Name)
	replicas := m.Spec.Size
	version := m.Spec.Version

	var userID int64 = 999
	trueBool := true

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			ServiceName: m.Name + "-service",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Name:   "janusgraph",
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						SupplementalGroups: []int64{userID},
						// RunAsNonRoot:       &trueBool,
					},
					ServiceAccountName: "janus-custom-sa",
					Containers: []corev1.Container{
						{
							Image: "janusgraph/janusgraph:" + version,
							Name:  "janusgraph",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8182,
									Name:          "janusgraph",
								},
							},
							// ReadinessProbe: &corev1.Probe{
							// 	InitialDelaySeconds: 480,
							// 	PeriodSeconds:       30,
							// 	TimeoutSeconds:      10,
							// 	FailureThreshold:    3,
							// 	Handler: corev1.Handler{
							// 		Exec: &corev1.ExecAction{
							// 			Command: []string{"sh", "/tmp/readiness.sh"},
							// 		},
							// 	},
							// },
							// VolumeMounts: []corev1.VolumeMount{
							// 	{
							// 		Name:      m.Name + "-db",
							// 		MountPath: "/opt/janusgraph/db",
							// 	},
							// },
							Env: []corev1.EnvVar{},
						}},
					RestartPolicy: corev1.RestartPolicyAlways,
					// Volumes: []corev1.Volume{
					// 	{
					// 		Name: m.Name + "-db",
					// 		VolumeSource: corev1.VolumeSource{
					// 			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					// 				ClaimName: m.Name + "-pvc",
					// 			},
					// 		},
					// 	},
					// },
				},
			},
		},
	}
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
