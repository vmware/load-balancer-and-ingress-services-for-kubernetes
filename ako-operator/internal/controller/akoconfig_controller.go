/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.

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

package controller

import (
	"context"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/go-logr/logr"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1beta1"
)

const (
	CleanupFinalizer = "ako.vmware.com/cleanup"
)

// AKOConfigConditionType is a valid value for AKOConfigStatus.State
type AKOConfigConditionType string

const (
	AKOConfigStatusReady      AKOConfigConditionType = "Ready"
	AKOConfigStatusError      AKOConfigConditionType = "Error"
	AKOConfigStatusProcessing AKOConfigConditionType = "Processing"
)

var (
	rebootRequired = false
	objectList     map[types.NamespacedName]client.Object
	objListOnce    sync.Once
)

// AKOConfigReconciler reconciles a AKOConfig object
type AKOConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *rest.Config
	//Log    logr.Logger
}

func getObjectList() map[types.NamespacedName]client.Object {
	objListOnce.Do(func() {
		objectList = make(map[types.NamespacedName]client.Object)
	})
	return objectList
}

func finalizerInList(finalizers []string, key string) bool {
	for _, f := range finalizers {
		if f == key {
			return true
		}
	}
	return false
}

func removeFinalizer(finalizers []string, key string) (result []string) {
	for _, f := range finalizers {
		if f == key {
			continue
		}
		result = append(result, f)
	}
	return result
}

// RBAC permissions based on config/rbac/role.yaml
// +kubebuilder:rbac:groups="",resources=configmaps;configmaps/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=serviceaccounts;serviceaccounts/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs;akoconfigs/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=akoconfigs/status,verbs=get;patch;update
// +kubebuilder:rbac:groups=apps,resources=statefulsets;statefulsets/status;statefulsets/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;clusterroles;clusterrolebindings/finalizers;clusterroles/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=extensions;networking.k8s.io,resources=ingresses;ingresses/status,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups="",resources=services;services/status,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch;update
// +kubebuilder:rbac:groups=crd.projectcalico.org,resources=blockaffinities;blockaffinities/status,verbs=get;watch;list
// +kubebuilder:rbac:groups=network.openshift.io,resources=hostsubnets,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/status,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups=ako.vmware.com,resources=hostrules;hostrules/status;hostrules/finalizers;httprules;httprules/status;httprules/finalizers;aviinfrasettings;aviinfrasettings/status;aviinfrasettings/finalizers;l4rules;l4rules/status;l4rules/finalizers;ssorules;ssorules/status;ssorules/finalizers;l7rules;l7rules/status;l7rules/finalizers,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions;customresourcedefinitions/status;customresourcedefinitions/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gateways;gateways/status;gatewayclasses;gateways/finalizers;gatewayclasses/status;gatewayclasses/finalizers,verbs=create;delete;get;watch;list;patch;update
// +kubebuilder:rbac:groups="",resources=*,verbs=get;watch;list
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses;ingressclasses/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=secrets;secrets/status;secrets/finalizers,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;get;update
// +kubebuilder:rbac:groups=cilium.io,resources=ciliumnodes,verbs=get;watch;list
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses;gatewayclasses/status;gateways;gateways/status;httproutes;httproutes/status,verbs=get;watch;list;patch;update;create;delete
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the AKOConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *AKOConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Verify if a CR of AKOConfig exists")
	var akoBeta akov1beta1.AKOConfig
	err := r.Get(ctx, req.NamespacedName, &akoBeta)
	if err != nil {
		if errors.IsNotFound(err) {
			// akoconfig object got deleted, before we come here, so just return
			logger.Info("AKOConfig resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "unable to fetch AKOConfig object")
		return ctrl.Result{}, err
	}

	if !akoBeta.GetDeletionTimestamp().IsZero() {
		if finalizerInList(akoBeta.GetFinalizers(), CleanupFinalizer) {
			if err := r.CleanupArtifacts(ctx, logger); err != nil {
				return ctrl.Result{}, err
			}

			patch := client.MergeFrom(akoBeta.DeepCopy())
			akoBeta.Finalizers = removeFinalizer(akoBeta.Finalizers, CleanupFinalizer)
			if err := r.Patch(context.TODO(), &akoBeta, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
		// return from here, no more reconciliation as the AKOConfig is being deleted
		return ctrl.Result{}, nil
	}

	// Update status to Processing
	if akoBeta.Status.State != string(AKOConfigStatusProcessing) && akoBeta.Status.State != string(AKOConfigStatusReady) {
		patch := client.MergeFrom(akoBeta.DeepCopy())
		akoBeta.Status.State = string(AKOConfigStatusProcessing)
		logger.Info("Patching akoconfig object with processing state")
		if err := r.Status().Patch(ctx, &akoBeta, patch); err != nil {
			logger.Error(err, "unable to update AKOConfig status to Processing")
			return ctrl.Result{}, err
		}
	}
	// reconcile all objects
	err = r.ReconcileAllArtifacts(ctx, akoBeta, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status to Ready
	if akoBeta.Status.State != string(AKOConfigStatusReady) {
		patch := client.MergeFrom(akoBeta.DeepCopy())
		akoBeta.Status.State = string(AKOConfigStatusReady)
		logger.Info("Patching akoconfig object with ready state")
		if err := r.Status().Patch(ctx, &akoBeta, patch); err != nil {
			logger.Error(err, "unable to update AKOConfig status to Ready")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *AKOConfigReconciler) ReconcileAllArtifacts(ctx context.Context, ako akov1beta1.AKOConfig, log logr.Logger) error {
	// If an error occurs during artifact reconciliation, update status to Error
	var reconcileErr error
	defer func() {
		if reconcileErr != nil {
			log.Error(reconcileErr, "Got error while reconciling")
			patch := client.MergeFrom(ako.DeepCopy())
			ako.Status.State = string(AKOConfigStatusError)
			log.Info("Patching akoconfig object with error state")
			if err := r.Status().Patch(ctx, &ako, patch); err != nil {
				log.Error(err, "unable to update AKOConfig status to Error during artifact reconciliation")
			}
		}
	}()

	secretNamespacedName := types.NamespacedName{Namespace: AviSystemNS, Name: AviSecretName}
	var aviSecret v1.Secret
	err := r.Get(ctx, secretNamespacedName, &aviSecret)
	if err != nil {
		log.Error(err, "secret named avi-secret is must for starting AKO controller")
		return err
	}

	// reconcile all the required artifacts for AKO
	err = createOrUpdateConfigMap(ctx, ako, log, r)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createOrUpdateServiceAccount(ctx, ako, log, r)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createOrUpdateClusterRole(ctx, ako, log, r)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createOrUpdateClusterroleBinding(ctx, ako, log, r)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createCRDs(r.Config, log)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createOrUpdateGatewayClass(ctx, ako, log, r)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	err = createOrUpdateStatefulSet(ctx, ako, log, r, aviSecret)
	if err != nil {
		reconcileErr = err
		return reconcileErr
	}

	return nil
}

func (r *AKOConfigReconciler) CleanupArtifacts(ctx context.Context, log logr.Logger) error {
	log.V(0).Info("cleaning up all the artifacts")
	objList := getObjectList()
	if len(objList) == 0 {
		// AKOConfig was deleted, but during the same time, the operator was restarted
		var cm corev1.ConfigMap
		if err := r.Get(ctx, getConfigMapName(), &cm); err != nil {
			log.V(0).Info("error getting configmap", "error", err)
		} else {
			objList[getConfigMapName()] = &cm
		}
		var sf appsv1.StatefulSet
		if err := r.Get(ctx, getSFNamespacedName(), &sf); err != nil {
			log.V(0).Info("error getting statefulset", "error", err)
		} else {
			objList[getSFNamespacedName()] = &sf
		}
		var cr rbacv1.ClusterRole
		if err := r.Get(ctx, getCRName(), &cr); err != nil {
			log.V(0).Info("error getting clusterrole", "error", err)
		} else {
			objList[getCRName()] = &cr
		}
		var crb rbacv1.ClusterRoleBinding
		if err := r.Get(ctx, getCRBName(), &crb); err != nil {
			log.V(0).Info("error getting clusterrolebinding", "error", err)
		} else {
			objList[getCRBName()] = &crb
		}
		var sa v1.ServiceAccount
		if err := r.Get(ctx, getSAName(), &sa); err != nil {
			log.V(0).Info("error getting serviceaccount", "error", err)
		} else {
			objList[getSAName()] = &sa
		}

		var gwClass gatewayv1.GatewayClass
		if err := r.Get(ctx, getGWClassName(), &gwClass); err != nil {
			log.V(0).Info("error getting gatewayclass", "error", err)
		} else {
			objList[getGWClassName()] = &gwClass
		}
	}
	for objName, obj := range objList {
		if err := r.deleteIfExists(ctx, objName, obj); err != nil {
			log.Error(err, "error while deleting object")
			return err
		}
	}
	err := deleteCRDs(r.Config, log)
	if err != nil {
		log.Error(err, "error while deleting crds")
		return err
	}
	return nil
}

func (r *AKOConfigReconciler) deleteIfExists(ctx context.Context, objNsName types.NamespacedName, object client.Object) error {
	err := r.Client.Get(ctx, objNsName, object)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if err == nil {
		if err := r.Client.Delete(ctx, object); err != nil {
			return err
		}
	}
	return nil
}

// mapClusterScopedToAKOConfig is a mapping function used by Watches to find the AKOConfig
// instance to reconcile when a cluster-scoped resource changes. This function lists all
// AKOConfig objects in the operator's namespace and returns a reconciliation request for each one.
// This ensures that changes to shared, cluster-scoped resources trigger a reconciliation
// of the active AKOConfig instance, regardless of its name.
func (r *AKOConfigReconciler) mapClusterScopedToAKOConfig(ctx context.Context, obj client.Object) []reconcile.Request {
	log := log.FromContext(ctx)
	akoConfigList := &akov1beta1.AKOConfigList{}
	// List all AKOConfig objects in the avi-system namespace.
	// Typically, there should only be one.
	if err := r.List(ctx, akoConfigList, client.InNamespace(AviSystemNS)); err != nil {
		log.Error(err, "unable to list AKOConfigs to map cluster-scoped resource change")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0, len(akoConfigList.Items))
	for _, item := range akoConfigList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		})
	}
	if len(requests) > 1 {
		log.Info("found multiple AKOConfig instances, which is not a typical configuration. Reconciliation will be triggered for all of them.", "count", len(requests))
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *AKOConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1beta1.AKOConfig{}).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == AviSystemNS && obj.GetName() == ConfigMapName
		}))).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == AviSystemNS && obj.GetName() == StatefulSetName
		}))).
		Owns(&corev1.ServiceAccount{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == AviSystemNS && obj.GetName() == AKOServiceAccount
		}))).
		// Watches cluster-scoped resources that are managed by this controller.
		// A namespaced AKOConfig cannot be a controller owner of cluster-scoped resources
		// via OwnerReference, so Owns() is not applicable here.
		// Instead, we use Watches() with a map function to explicitly trigger reconciliation
		// of the AKOConfig instance when these cluster-scoped objects change.
		Watches(
			&rbacv1.ClusterRole{},
			handler.EnqueueRequestsFromMapFunc(r.mapClusterScopedToAKOConfig),
			builder.WithPredicates(
				predicate.And(
					predicate.ResourceVersionChangedPredicate{},
					predicate.NewPredicateFuncs(func(obj client.Object) bool {
						return obj.GetName() == AKOCR
					}),
				),
			),
		).
		Watches(
			&rbacv1.ClusterRoleBinding{},
			handler.EnqueueRequestsFromMapFunc(r.mapClusterScopedToAKOConfig),
			builder.WithPredicates(
				predicate.And(
					predicate.ResourceVersionChangedPredicate{},
					predicate.NewPredicateFuncs(func(obj client.Object) bool {
						return obj.GetName() == CRBName
					}),
				),
			),
		).
		Watches(
			&gatewayv1.GatewayClass{},
			handler.EnqueueRequestsFromMapFunc(r.mapClusterScopedToAKOConfig),
			builder.WithPredicates(
				predicate.And(
					predicate.ResourceVersionChangedPredicate{},
					predicate.NewPredicateFuncs(func(obj client.Object) bool {
						return obj.GetName() == GWClassName
					}),
				),
			),
		).
		// Watch the CRDs that this operator is responsible for creating and maintaining.
		// This ensures that if a CRD is modified or deleted, the operator will reconcile
		// and restore it to the desired state.
		Watches(
			&apiextensionv1.CustomResourceDefinition{},
			handler.EnqueueRequestsFromMapFunc(r.mapClusterScopedToAKOConfig),
			builder.WithPredicates(
				predicate.And(
					predicate.ResourceVersionChangedPredicate{},
					predicate.NewPredicateFuncs(func(obj client.Object) bool {
						managedCRDs := map[string]struct{}{
							"hostrules.ako.vmware.com":        {},
							"httprules.ako.vmware.com":        {},
							"aviinfrasettings.ako.vmware.com": {},
							"l4rules.ako.vmware.com":          {},
							"ssorules.ako.vmware.com":         {},
							"l7rules.ako.vmware.com":          {},
						}
						_, ok := managedCRDs[obj.GetName()]
						return ok
					}),
				),
			),
		).
		Complete(r)
}
